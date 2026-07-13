package user

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/he-end/verify-reverse/auth/repository"
	"github.com/he-end/verify-reverse/auth/repository/auth"
)

type UserService struct {
	repo        *auth.AuthRepository
	sessionRepo *auth.SessionRepository
	verifyRepo  *auth.VerificationRepository
}

func NewUserService(repo *auth.AuthRepository, sessionRepo *auth.SessionRepository, verifyRepo *auth.VerificationRepository) *UserService {
	return &UserService{
		repo:        repo,
		sessionRepo: sessionRepo,
		verifyRepo:  verifyRepo,
	}
}

func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*auth.User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	return user, nil
}

type UpdateProfileInput struct {
	Name string
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, input UpdateProfileInput) (*auth.User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	user.Name = input.Name
	user.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	return user, nil
}

func (s *UserService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return repository.ErrWrongPassword
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user.PasswordHash = string(hashed)
	user.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	if err := s.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("delete sessions: %w", err)
	}

	return nil
}

func (s *UserService) InitiateWANumberChange(ctx context.Context, userID uuid.UUID, newNumber string) (*string, error) {
	exists, err := s.repo.ExistsByNumber(ctx, newNumber)
	if err != nil {
		return nil, fmt.Errorf("check number exists: %w", err)
	}
	if exists {
		user, err := s.repo.FindByNumber(ctx, newNumber)
		if err != nil && !errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("find user by number: %w", err)
		}
		if user != nil && user.ID == userID {
			return nil, repository.ErrDuplicateKey
		}
		return nil, repository.ErrNumberTaken
	}

	code, err := generateChangeWACode()
	if err != nil {
		return nil, fmt.Errorf("generate verification code: %w", err)
	}

	expiresAt := time.Now().Add(15 * time.Minute)
	vcID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate verification code ID: %w", err)
	}

	vc := &auth.VerificationCode{
		ID:          vcID,
		Contact:     newNumber,
		ContactType: "wa",
		Code:        code,
		Name:        "",
		ExpiresAt:   expiresAt,
		UserID:      &userID,
		Purpose:     "change_wa",
	}

	if err := s.verifyRepo.Create(ctx, vc); err != nil {
		return nil, fmt.Errorf("create verification code: %w", err)
	}

	return &code, nil
}

func (s *UserService) CompleteWANumberChange(ctx context.Context, code, contact string) error {
	vc, err := s.verifyRepo.FindByCodeAndContactAndPurpose(ctx, code, contact, "change_wa")
	if err != nil {
		return fmt.Errorf("find verification code: %w", err)
	}

	if vc.UserID == nil {
		return repository.ErrVerificationNotValid
	}

	user, err := s.repo.FindByID(ctx, *vc.UserID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}

	user.Number = &contact
	user.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("update user number: %w", err)
	}

	if err := s.verifyRepo.MarkUsed(ctx, vc.ID, *vc.UserID); err != nil {
		return fmt.Errorf("mark verification used: %w", err)
	}

	return nil
}

func generateChangeWACode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("generate random index: %w", err)
		}
		b[i] = charset[n.Int64()]
	}
	return "CHGWA-" + string(b), nil
}
