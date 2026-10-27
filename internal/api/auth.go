package api

import (
	"errors"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/markoomisicc/uptime-api/internal/store"
	"golang.org/x/crypto/bcrypt"
)

var (
	usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`)
	phonePattern    = regexp.MustCompile(`^\+[1-9][0-9]{6,14}$`)
)

var dummyPasswordHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)

func isSupportedLanguage(language string) bool {
	return language == "en" || language == "de"
}

// Register

type registerRequest struct {
	Email     string  `json:"email"`
	Username  string  `json:"username"`
	Password  string  `json:"password"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Phone     *string `json:"phone"`
	Timezone  string  `json:"timezone"`
	Language  string  `json:"language"`
}

func (req *registerRequest) normalize() {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Username = strings.TrimSpace(req.Username)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)

	req.Timezone = strings.TrimSpace(req.Timezone)
	if req.Timezone == "" {
		req.Timezone = "UTC"
	}

	req.Language = strings.TrimSpace(req.Language)
	if req.Language == "" {
		req.Language = "en"
	}

	if req.Phone != nil {
		phone := strings.TrimSpace(*req.Phone)
		if phone == "" {
			req.Phone = nil
		} else {
			req.Phone = &phone
		}
	}
}

func (req registerRequest) validate() map[string]fieldError {
	errs := map[string]fieldError{}

	if addr, err := mail.ParseAddress(req.Email); err != nil || addr.Address != req.Email || len(req.Email) > 254 {
		errs["email"] = fieldError{"invalid", "must be a valid email address"}
	}

	if !usernamePattern.MatchString(req.Username) {
		errs["username"] = fieldError{"invalid", "must be 3-30 characters long and contain only letters, numbers, _ or -"}
	}

	if len(req.Password) < 8 || len(req.Password) > 72 {
		errs["password"] = fieldError{"invalid_length", "must be between 8 and 72 characters long"}
	}

	if req.FirstName == "" {
		errs["first_name"] = fieldError{"required", "is required"}
	} else if len(req.FirstName) > 100 {
		errs["first_name"] = fieldError{"too_long", "must be at most 100 characters long"}
	}

	if req.LastName == "" {
		errs["last_name"] = fieldError{"required", "is required"}
	} else if len(req.LastName) > 100 {
		errs["last_name"] = fieldError{"too_long", "must be at most 100 characters long"}
	}

	if req.Phone != nil && !phonePattern.MatchString(*req.Phone) {
		errs["phone"] = fieldError{"invalid", "must be in international format, e.g. +385911234567"}
	}

	if _, err := time.LoadLocation(req.Timezone); err != nil {
		errs["timezone"] = fieldError{"invalid", "must be a valid time zone, e.g. Europe/Zagreb"}
	}

	if !isSupportedLanguage(req.Language) {
		errs["language"] = fieldError{"invalid", "must be one of: en, de"}
	}

	return errs
}

type userResponse struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	Username      string    `json:"username"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Phone         *string   `json:"phone"`
	Timezone      string    `json:"timezone"`
	Language      string    `json:"language"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

func newUserResponse(user store.User) userResponse {
	return userResponse{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Phone:         user.Phone,
		Timezone:      user.Timezone,
		Language:      user.Language,
		EmailVerified: user.EmailVerifiedAt != nil,
		CreatedAt:     user.CreatedAt.UTC(),
	}
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_body", "Request body must be valid JSON")
		return
	}

	req.normalize()
	if errs := req.validate(); len(errs) > 0 {
		writeFieldErrors(w, http.StatusUnprocessableEntity, "validation_failed", "Validation failed", errs)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeInternalError(w, "hash password", err)
		return
	}

	user, err := s.queries.CreateUser(r.Context(), store.CreateUserParams{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hash),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		Timezone:     req.Timezone,
		Language:     req.Language,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_email_unique":
				writeFieldErrors(w, http.StatusConflict, "already_exists", "Account already exists",
					map[string]fieldError{"email": {"taken", "is already registered"}})
				return
			case "users_username_unique":
				writeFieldErrors(w, http.StatusConflict, "already_exists", "Account already exists",
					map[string]fieldError{"username": {"taken", "is already taken"}})
				return
			}
		}

		writeInternalError(w, "create user", err)
		return
	}

	writeJSON(w, http.StatusCreated, newUserResponse(user))
}

// Login and logout

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_body", "Request body must be valid JSON")
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	user, err := s.queries.GetUserByEmail(r.Context(), email)
	userFound := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		writeInternalError(w, "get user by email", err)
		return
	}

	hash := dummyPasswordHash
	if userFound {
		hash = []byte(user.PasswordHash)
	}

	passwordErr := bcrypt.CompareHashAndPassword(hash, []byte(req.Password))
	if !userFound || passwordErr != nil {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password")
		return
	}

	token, expiresAt, err := s.createSession(r, user.ID)
	if err != nil {
		writeInternalError(w, "create session", err)
		return
	}

	if err := s.queries.UpdateUserLastLogin(r.Context(), user.ID); err != nil {
		// Not critical: the user is logged in even if this update fails.
		logError("update last login", err)
	}

	s.setSessionCookie(w, token, expiresAt)
	writeJSON(w, http.StatusOK, newUserResponse(user))
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		if err := s.queries.DeleteSession(r.Context(), hashToken(cookie.Value)); err != nil {
			writeInternalError(w, "delete session", err)
			return
		}
	}

	s.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// Current user

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, newUserResponse(userFromContext(r.Context())))
}

type updateMeRequest struct {
	Language string `json:"language"`
}

func (s *Server) handleUpdateMe(w http.ResponseWriter, r *http.Request) {
	var req updateMeRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_body", "Request body must be valid JSON")
		return
	}

	if !isSupportedLanguage(req.Language) {
		writeFieldErrors(w, http.StatusUnprocessableEntity, "validation_failed", "Validation failed",
			map[string]fieldError{"language": {"invalid", "must be one of: en, de"}})
		return
	}

	user := userFromContext(r.Context())

	updated, err := s.queries.UpdateUserLanguage(r.Context(), store.UpdateUserLanguageParams{
		ID:       user.ID,
		Language: req.Language,
	})
	if err != nil {
		writeInternalError(w, "update user language", err)
		return
	}

	writeJSON(w, http.StatusOK, newUserResponse(updated))
}
