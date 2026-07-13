package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/he-end/verify-reverse/auth/repository"
	"github.com/he-end/verify-reverse/auth/repository/auth"
)

type AuthService struct {
	repo              *auth.AuthRepository
	sessionRepo       *auth.SessionRepository
	verifyRepo        *auth.VerificationRepository
	jwt               *JWTService
	allowMultiSession bool
	maxSession        int
}

func NewAuthService(repo *auth.AuthRepository, sessionRepo *auth.SessionRepository, verifyRepo *auth.VerificationRepository, jwt *JWTService, allowMultiSession bool, maxSession int) *AuthService {
	return &AuthService{
		repo:              repo,
		sessionRepo:       sessionRepo,
		verifyRepo:        verifyRepo,
		jwt:               jwt,
		allowMultiSession: allowMultiSession,
		maxSession:        maxSession,
	}
}

type RegisterInput struct {
	Email  *string
	Number *string
	Name   string
	Pwd    string
}

type LoginInput struct {
	Email  *string
	Number *string
	Pwd    string
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*auth.User, *TokenPair, error) {
	if input.Email == nil && input.Number == nil {
		return nil, nil, repository.ErrMissingContact
	}

	if input.Email != nil {
		exists, err := s.repo.ExistsByEmail(ctx, *input.Email)
		if err != nil {
			return nil, nil, fmt.Errorf("check email exists: %w", err)
		}
		if exists {
			return nil, nil, repository.ErrDuplicateKey
		}
	}

	if input.Number != nil {
		exists, err := s.repo.ExistsByNumber(ctx, *input.Number)
		if err != nil {
			return nil, nil, fmt.Errorf("check number exists: %w", err)
		}
		if exists {
			return nil, nil, repository.ErrDuplicateKey
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Pwd), 12)
	if err != nil {
		return nil, nil, fmt.Errorf("hash password: %w", err)
	}

	userID, err := uuid.NewV7()
	if err != nil {
		return nil, nil, fmt.Errorf("generate user ID: %w", err)
	}

	user := &auth.User{
		ID:           userID,
		Email:        input.Email,
		Number:       input.Number,
		Name:         input.Name,
		PasswordHash: string(hashedPassword),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, nil, fmt.Errorf("create user: %w", err)
	}

	userHash := UserHash(user.PasswordHash)
	tokens, err := s.jwt.GenerateTokenPair(ctx, user.ID, userHash)
	if err != nil {
		return nil, nil, fmt.Errorf("generate token pair: %w", err)
	}

	sessionID, err := uuid.NewV7()
	if err != nil {
		return nil, nil, fmt.Errorf("generate session ID: %w", err)
	}

	session := &auth.Session{
		ID:            sessionID,
		UserID:        user.ID,
		RefreshToken:  tokens.RefreshToken,
		AccessTokenID: uuid.New().String(),
		ExpiresAt:     tokens.ExpiresAt,
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, nil, fmt.Errorf("create session: %w", err)
	}

	return user, tokens, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*auth.User, *TokenPair, error) {
	if input.Email == nil && input.Number == nil {
		return nil, nil, repository.ErrMissingContact
	}

	var user *auth.User
	var err error

	if input.Email != nil {
		user, err = s.repo.FindByEmail(ctx, *input.Email)
	} else {
		user, err = s.repo.FindByNumber(ctx, *input.Number)
	}
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, repository.ErrInvalidCredentials
		}
		return nil, nil, fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Pwd)); err != nil {
		return nil, nil, repository.ErrInvalidCredentials
	}

	userHash := UserHash(user.PasswordHash)
	tokens, err := s.jwt.GenerateTokenPair(ctx, user.ID, userHash)
	if err != nil {
		return nil, nil, fmt.Errorf("generate token pair: %w", err)
	}

	if !s.allowMultiSession {
		if err := s.sessionRepo.DeleteByUserID(ctx, user.ID); err != nil {
			return nil, nil, fmt.Errorf("delete old sessions: %w", err)
		}
	} else if s.maxSession > 0 {
		count, err := s.sessionRepo.CountByUserID(ctx, user.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("count sessions by user: %w", err)
		}
		if count >= s.maxSession {
			if err := s.sessionRepo.DeleteOldestByUserID(ctx, user.ID); err != nil {
				return nil, nil, fmt.Errorf("delete oldest session: %w", err)
			}
		}
	}

	sessionID, err := uuid.NewV7()
	if err != nil {
		return nil, nil, fmt.Errorf("generate session ID: %w", err)
	}

	session := &auth.Session{
		ID:            sessionID,
		UserID:        user.ID,
		RefreshToken:  tokens.RefreshToken,
		AccessTokenID: uuid.New().String(),
		ExpiresAt:     tokens.ExpiresAt,
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, nil, fmt.Errorf("create session: %w", err)
	}

	return user, tokens, nil
}

func (s *AuthService) InitiateWAVerify(ctx context.Context, number, name, pwd string) (*string, time.Time, error) {
	isPhantom := false

	exists, err := s.repo.ExistsByNumber(ctx, number)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("check number exists: %w", err)
	}
	if exists {
		isPhantom = true
	}

	if !isPhantom {
		pending, err := s.verifyRepo.ExistsPending(ctx, number, "wa")
		if err != nil {
			return nil, time.Time{}, fmt.Errorf("check pending verification: %w", err)
		}
		if pending {
			isPhantom = true
		}
	}

	code, err := generateVerificationCode()
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("generate verification code: %w", err)
	}

	var passwordHash string
	if pwd != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(pwd), 12)
		if err != nil {
			return nil, time.Time{}, fmt.Errorf("hash password: %w", err)
		}
		passwordHash = string(hashed)
	}

	expiresAt := time.Now().Add(15 * time.Minute)
	vcID, err := uuid.NewV7()
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("generate verification code ID: %w", err)
	}

	vc := &auth.VerificationCode{
		ID:           vcID,
		Contact:      number,
		ContactType:  "wa",
		Code:         code,
		Name:         name,
		PasswordHash: passwordHash,
		ExpiresAt:    expiresAt,
		IsPhantom:    isPhantom,
	}

	if err := s.verifyRepo.Create(ctx, vc); err != nil {
		return nil, time.Time{}, fmt.Errorf("create verification code: %w", err)
	}

	return &code, expiresAt, nil
}

func (s *AuthService) IsAlreadyVerified(ctx context.Context, contact string) (bool, error) {
	_, err := s.repo.FindByNumber(ctx, contact)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("find user by number: %w", err)
	}
	return true, nil
}

func (s *AuthService) CompleteWAVerify(ctx context.Context, code, contact string) (*auth.User, error) {
	vc, err := s.verifyRepo.FindByCodeAndContact(ctx, code, contact)
	if err != nil {
		return nil, fmt.Errorf("find verification code: %w", err)
	}

	if vc.IsPhantom {
		return nil, repository.ErrVerificationNotValid
	}

	userID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate user ID: %w", err)
	}

	user := &auth.User{
		ID:           userID,
		Number:       &vc.Contact,
		Name:         vc.Name,
		PasswordHash: vc.PasswordHash,
		Status:       "active",
	}
	now := time.Now()
	user.VerifiedAt = &now

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	if err := s.verifyRepo.MarkUsed(ctx, vc.ID, user.ID); err != nil {
		return nil, fmt.Errorf("mark verification used: %w", err)
	}

	return user, nil
}

func generateVerificationCode() (string, error) {
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
	return "VRFY-" + string(b), nil
}

func UserHashFallback(userID uuid.UUID, status string) string {
	h := sha256.Sum256([]byte(userID.String() + status))
	return hex.EncodeToString(h[:])
}

func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	if err := s.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("logout: %w", err)
	}
	return nil
}

func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*auth.User, *TokenPair, error) {
	session, err := s.sessionRepo.FindByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, nil, fmt.Errorf("find session: %w", err)
	}

	if time.Now().After(session.ExpiresAt) {
		if err := s.sessionRepo.Delete(ctx, session.ID); err != nil {
			return nil, nil, fmt.Errorf("delete expired session: %w", err)
		}
		return nil, nil, repository.ErrTokenExpired
	}

	user, err := s.repo.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("find user: %w", err)
	}

	userHash := UserHash(user.PasswordHash)
	tokens, err := s.jwt.GenerateTokenPair(ctx, user.ID, userHash)
	if err != nil {
		return nil, nil, fmt.Errorf("generate token pair: %w", err)
	}

	if err := s.sessionRepo.Delete(ctx, session.ID); err != nil {
		return nil, nil, fmt.Errorf("delete old session: %w", err)
	}

	newSessionID, err := uuid.NewV7()
	if err != nil {
		return nil, nil, fmt.Errorf("generate session ID: %w", err)
	}

	newSession := &auth.Session{
		ID:            newSessionID,
		UserID:        user.ID,
		RefreshToken:  tokens.RefreshToken,
		AccessTokenID: uuid.New().String(),
		ExpiresAt:     tokens.ExpiresAt,
	}
	if err := s.sessionRepo.Create(ctx, newSession); err != nil {
		return nil, nil, fmt.Errorf("create session: %w", err)
	}

	return user, tokens, nil
}
