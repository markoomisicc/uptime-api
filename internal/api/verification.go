package api

import (
	"context"
	"crypto/rand"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/markoomisicc/uptime-api/internal/store"
)

const (
	purposeEmailVerification = "email_verification"
	verificationTokenTTL     = 24 * time.Hour
	resendCooldown           = 60 * time.Second
)

func (s *Server) sendVerificationEmail(ctx context.Context, user store.User) error {
	if err := s.queries.DeleteUserTokens(ctx, store.DeleteUserTokensParams{
		UserID:  user.ID,
		Purpose: purposeEmailVerification,
	}); err != nil {
		return err
	}

	token := rand.Text()

	if err := s.queries.CreateUserToken(ctx, store.CreateUserTokenParams{
		UserID:    user.ID,
		Purpose:   purposeEmailVerification,
		TokenHash: hashToken(token),
		ExpiresAt: time.Now().Add(verificationTokenTTL),
	}); err != nil {
		return err
	}

	link := s.cfg.AppURL + "/verify-email?token=" + url.QueryEscape(token)
	subject, body := verificationEmail(user.Language, user.FirstName, link)

	return s.mailer.Send(user.Email, subject, body)
}

// Verify email

type verifyEmailRequest struct {
	Token string `json:"token"`
}

func (s *Server) handleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_body", "Request body must be valid JSON")
		return
	}

	ctx := r.Context()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		writeInternalError(w, "begin transaction", err)
		return
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	userID, err := qtx.ConsumeUserToken(ctx, store.ConsumeUserTokenParams{
		TokenHash: hashToken(req.Token),
		Purpose:   purposeEmailVerification,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusBadRequest, "invalid_token", "Verification link is invalid or has expired")
			return
		}
		writeInternalError(w, "consume verification token", err)
		return
	}

	user, err := qtx.MarkUserEmailVerified(ctx, userID)
	if err != nil {
		writeInternalError(w, "mark email verified", err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeInternalError(w, "commit transaction", err)
		return
	}

	token, expiresAt, err := s.createSession(r, user.ID)
	if err != nil {
		writeInternalError(w, "create session", err)
		return
	}

	s.setSessionCookie(w, token, expiresAt)
	writeJSON(w, http.StatusOK, newUserResponse(user))
}

// Resend verification email

type resendVerificationRequest struct {
	Email string `json:"email"`
}

func (s *Server) handleResendVerification(w http.ResponseWriter, r *http.Request) {
	var req resendVerificationRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_body", "Request body must be valid JSON")
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	if !s.resendLimiter.Allow(email) {
		w.Header().Set("Retry-After", strconv.Itoa(int(resendCooldown.Seconds())))
		writeError(w, http.StatusTooManyRequests, "too_many_requests", "Please wait before requesting another email")
		return
	}

	user, err := s.queries.GetUserByEmail(r.Context(), email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		writeInternalError(w, "get user by email", err)
		return
	}

	if err == nil && user.EmailVerifiedAt == nil {
		if err := s.sendVerificationEmail(r.Context(), user); err != nil {
			writeInternalError(w, "send verification email", err)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
