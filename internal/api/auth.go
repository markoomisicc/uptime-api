package api

import (
	"errors"
	"log"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/markoomisicc/uptime-api/internal/store"
	"golang.org/x/crypto/bcrypt"
)

var (
	usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`)
	phonePattern    = regexp.MustCompile(`^\+[1-9][0-9]{6,14}$`)
)

type registerRequest struct {
	Email     string  `json:"email"`
	Username  string  `json:"username"`
	Password  string  `json:"password"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Phone     *string `json:"phone"`
	Timezone  string  `json:"timezone"`
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

	if req.Phone != nil {
		phone := strings.TrimSpace(*req.Phone)
		if phone == "" {
			req.Phone = nil
		} else {
			req.Phone = &phone
		}
	}
}

func (req registerRequest) validate() map[string]string {
	errs := map[string]string{}

	if addr, err := mail.ParseAddress(req.Email); err != nil || addr.Address != req.Email || len(req.Email) > 254 {
		errs["email"] = "must be a valid email address"
	}

	if !usernamePattern.MatchString(req.Username) {
		errs["username"] = "must be 3-30 characters long and contain only letters, numbers, _ or -"
	}

	if len(req.Password) < 8 || len(req.Password) > 72 {
		errs["password"] = "must be between 8 and 72 characters long"
	}

	if req.FirstName == "" || len(req.FirstName) > 100 {
		errs["first_name"] = "is required and must be at most 100 characters long"
	}

	if req.LastName == "" || len(req.LastName) > 100 {
		errs["last_name"] = "is required and must be at most 100 characters long"
	}

	if req.Phone != nil && !phonePattern.MatchString(*req.Phone) {
		errs["phone"] = "must be in international format, e.g. +385911234567"
	}

	if _, err := time.LoadLocation(req.Timezone); err != nil {
		errs["timezone"] = "must be a valid time zone, e.g. Europe/Zagreb"
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
		EmailVerified: user.EmailVerifiedAt != nil,
		CreatedAt:     user.CreatedAt,
	}
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.normalize()
	if errs := req.validate(); len(errs) > 0 {
		writeFieldErrors(w, http.StatusUnprocessableEntity, "validation failed", errs)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("hash password: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
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
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_email_unique":
				writeFieldErrors(w, http.StatusConflict, "already exists", map[string]string{"email": "is already registered"})
				return
			case "users_username_unique":
				writeFieldErrors(w, http.StatusConflict, "already exists", map[string]string{"username": "is already taken"})
				return
			}
		}

		log.Printf("create user: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, newUserResponse(user))
}
