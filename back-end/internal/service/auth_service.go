package service

import (
	"context"
	"database/sql"
	"errors"
	"mm/internal/storage/repository"
	"mm/pkg/apperrors"
	"mm/pkg/mailer"
	"mm/pkg/mtype"
)

const (
	EmailVerificationCodeTopic = "Service: Verification Code"
)

type AuthService struct {
	CodeRepository repository.CodeRepository
	UserRepository *repository.UserRepository
	Mailer         mailer.Mailer
}

func NewAuthService(
	codeRepository repository.CodeRepository,
	mailer mailer.Mailer,
	userRepository *repository.UserRepository,
) *AuthService {
	return &AuthService{
		CodeRepository: codeRepository,
		UserRepository: userRepository,
		Mailer:         mailer,
	}
}

func (s *AuthService) SendCodeToRegisterUser(ctx context.Context, template string, email mtype.Email) error {
	user, err := s.UserRepository.FindByEmail(ctx, email)
	if err != nil {
		return err
	}

	if user == nil {
		return apperrors.NotFound("user not found")
	}

	return s.SendCodeOnEmail(ctx, template, email)
}

func (s *AuthService) SendCodeOnEmail(ctx context.Context, template string, email mtype.Email) error {
	code, err := generateVerificationCode()
	if err != nil {
		return err
	}

	err = s.CodeRepository.Save(ctx, email.String(), code)
	if err != nil {
		return err
	}

	mail := mailer.NewMail(email, template, EmailVerificationCodeTopic, code)

	err = s.Mailer.SendEmail(ctx, mail)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) CheckUsersEmailCode(ctx context.Context, email mtype.Email, code string) (bool, error) {

	storedCode, err := s.CodeRepository.GetCodeByEmail(ctx, email.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, apperrors.Unauthorized("code not found")
		}
		return false, apperrors.Internal("failed to get code by email", err)
	}

	if storedCode != code {
		return false, apperrors.Unauthorized("code is invalid")
	}

	err = s.CodeRepository.DeleteCodeByEmail(ctx, email.String())
	if err != nil {
		return false, apperrors.Internal("failed to delete code by email", err)
	}

	return true, nil
}
