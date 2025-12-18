package domain

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type authService struct {
	repo   ports.AuthRepo
	secret string
}

func NewAuthService(repo ports.AuthRepo, secret string) ports.AuthService {
	return &authService{
		repo:   repo,
		secret: secret,
	}
}

func (s *authService) Login(ctx context.Context, password string) (string, error) {
	realPass, err := s.repo.GetPassword(ctx)
	if err != nil {
		return "", err
	}
	if realPass == "" || password != realPass {
		return "", errors.New("invalid password")
	}
	return s.sign("allowed"), nil
}

func (s *authService) ValidateToken(ctx context.Context, token string) (bool, error) {
	return token == s.sign("allowed"), nil
}

func (s *authService) sign(msg string) string {
	h := hmac.New(sha256.New, []byte(s.secret))
	h.Write([]byte(msg))
	return hex.EncodeToString(h.Sum(nil))
}
