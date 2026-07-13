package auth_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/he-end/verify-reverse/verify/repository"
	"github.com/he-end/verify-reverse/verify/repository/auth"
	authsvc "github.com/he-end/verify-reverse/verify/service/auth"
	"github.com/he-end/verify-reverse/verify/testhelper"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	cleanup, err := testhelper.StartPGContainer(ctx)
	if err != nil {
		os.Exit(1)
	}
	defer cleanup()
	os.Exit(m.Run())
}

type serviceTestSetup struct {
	db         *bun.DB
	authRepo   *auth.AuthRepository
	verifyRepo *auth.VerificationRepository
	svc        *authsvc.AuthService
}

func newServiceTestSetup(t *testing.T) *serviceTestSetup {
	t.Helper()
	db := testhelper.NewTestDB(t)
	authRepo := auth.NewAuthRepository(db)
	verifyRepo := auth.NewVerificationRepository(db)
	svc := authsvc.NewAuthService(authRepo, nil, verifyRepo, nil, true, 5)
	t.Cleanup(func() {
		testhelper.TruncateAll(context.Background(), db)
	})
	return &serviceTestSetup{
		db:         db,
		authRepo:   authRepo,
		verifyRepo: verifyRepo,
		svc:        svc,
	}
}

func TestInitiateWAVerifyPhantomForExistingNumber(t *testing.T) {
	ts := newServiceTestSetup(t)
	ctx := t.Context()

	email := "phantom@example.com"
	number := "6281234567890"
	user := &auth.User{
		ID:           uuid.Must(uuid.NewV7()),
		Email:        &email,
		Number:       &number,
		Name:         "Existing User",
		PasswordHash: "hash",
		Status:       "active",
	}
	if err := ts.authRepo.Create(ctx, user); err != nil {
		t.Fatalf("Create user failed: %v", err)
	}

	code, expiresAt, err := ts.svc.InitiateWAVerify(ctx, number, "Test", "password")
	if err != nil {
		t.Fatalf("InitiateWAVerify should not return error for existing number, got: %v", err)
	}
	if code == nil {
		t.Fatal("expected a verification code, got nil")
	}
	if expiresAt.IsZero() {
		t.Fatal("expected a non-zero expiry time")
	}
}

func TestInitiateWAVerifyNoPhantomForNewNumber(t *testing.T) {
	ts := newServiceTestSetup(t)
	ctx := t.Context()

	number := "6289876543210"

	code, expiresAt, err := ts.svc.InitiateWAVerify(ctx, number, "New User", "password")
	if err != nil {
		t.Fatalf("InitiateWAVerify failed: %v", err)
	}
	if code == nil {
		t.Fatal("expected a verification code, got nil")
	}
	if expiresAt.IsZero() {
		t.Fatal("expected a non-zero expiry time")
	}

	user, err := ts.svc.CompleteWAVerify(ctx, *code, number)
	if err != nil {
		t.Fatalf("CompleteWAVerify should succeed for non-phantom code, got: %v", err)
	}
	if user.Number == nil || *user.Number != number {
		t.Errorf("expected number '%s', got %v", number, user.Number)
	}
}

func TestCompleteWAVerifyPhantomReturnsError(t *testing.T) {
	ts := newServiceTestSetup(t)
	ctx := t.Context()

	email := "phantom-verify@example.com"
	number := "6281111111111"
	user := &auth.User{
		ID:           uuid.Must(uuid.NewV7()),
		Email:        &email,
		Number:       &number,
		Name:         "Already Registered",
		PasswordHash: "hash",
		Status:       "active",
	}
	if err := ts.authRepo.Create(ctx, user); err != nil {
		t.Fatalf("Create user failed: %v", err)
	}

	code, _, err := ts.svc.InitiateWAVerify(ctx, number, "Test", "password")
	if err != nil {
		t.Fatalf("InitiateWAVerify failed: %v", err)
	}

	_, err = ts.svc.CompleteWAVerify(ctx, *code, number)
	if !errors.Is(err, repository.ErrVerificationNotValid) {
		t.Errorf("expected ErrVerificationNotValid, got: %v", err)
	}
}

func TestCompleteWAVerifySuccess(t *testing.T) {
	ts := newServiceTestSetup(t)
	ctx := t.Context()

	number := "6282222222222"

	code, _, err := ts.svc.InitiateWAVerify(ctx, number, "New User", "password")
	if err != nil {
		t.Fatalf("InitiateWAVerify failed: %v", err)
	}

	user, err := ts.svc.CompleteWAVerify(ctx, *code, number)
	if err != nil {
		t.Fatalf("CompleteWAVerify failed: %v", err)
	}
	if user == nil {
		t.Fatal("expected a user, got nil")
	}
	if user.Number == nil || *user.Number != number {
		t.Errorf("expected number '%s', got %v", number, user.Number)
	}
}
