package auth_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/he-end/verify-reverse/verify/repository"
	"github.com/he-end/verify-reverse/verify/repository/auth"
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

func newTestSetup(t *testing.T) *auth.AuthRepository {
	t.Helper()
	db := testhelper.NewTestDB(t)
	repo := auth.NewAuthRepository(db)
	t.Cleanup(func() {
		testhelper.TruncateAll(context.Background(), db)
	})
	return repo
}

func TestCreateAndFindUserByEmail(t *testing.T) {
	repo := newTestSetup(t)
	ctx := t.Context()

	email := "test@example.com"
	user := &auth.User{
		ID:           uuid.Must(uuid.NewV7()),
		Email:        &email,
		Name:         "Test User",
		PasswordHash: "hashed_password",
		Status:       "active",
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByEmail(ctx, email)
	if err != nil {
		t.Fatalf("FindByEmail failed: %v", err)
	}
	if found.Name != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", found.Name)
	}
	if found.Email == nil || *found.Email != email {
		t.Errorf("expected email '%s', got %v", email, found.Email)
	}
}

func TestCreateUserDuplicateEmail(t *testing.T) {
	repo := newTestSetup(t)
	ctx := t.Context()

	email := "dupe@example.com"
	user1 := &auth.User{
		ID:           uuid.Must(uuid.NewV7()),
		Email:        &email,
		Name:         "User One",
		PasswordHash: "hash1",
		Status:       "active",
	}
	if err := repo.Create(ctx, user1); err != nil {
		t.Fatalf("Create first user failed: %v", err)
	}

	user2 := &auth.User{
		ID:           uuid.Must(uuid.NewV7()),
		Email:        &email,
		Name:         "User Two",
		PasswordHash: "hash2",
		Status:       "active",
	}
	err := repo.Create(ctx, user2)
	if !errors.Is(err, repository.ErrDuplicateKey) {
		t.Errorf("expected ErrDuplicateKey, got: %v", err)
	}
}

func TestFindByEmailNotFound(t *testing.T) {
	repo := newTestSetup(t)
	ctx := t.Context()

	_, err := repo.FindByEmail(ctx, "nonexistent@example.com")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestExistsByEmail(t *testing.T) {
	repo := newTestSetup(t)
	ctx := t.Context()

	email := "exists@example.com"
	user := &auth.User{
		ID:           uuid.Must(uuid.NewV7()),
		Email:        &email,
		Name:         "Exists User",
		PasswordHash: "hash",
		Status:       "active",
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	exists, err := repo.ExistsByEmail(ctx, email)
	if err != nil {
		t.Fatalf("ExistsByEmail failed: %v", err)
	}
	if !exists {
		t.Error("expected user to exist")
	}

	exists, err = repo.ExistsByEmail(ctx, "missing@example.com")
	if err != nil {
		t.Fatalf("ExistsByEmail failed: %v", err)
	}
	if exists {
		t.Error("expected user to not exist")
	}
}

func TestDeleteUser(t *testing.T) {
	repo := newTestSetup(t)
	ctx := t.Context()

	email := "delete-me@example.com"
	user := &auth.User{
		ID:           uuid.Must(uuid.NewV7()),
		Email:        &email,
		Name:         "Delete Me",
		PasswordHash: "hash",
		Status:       "active",
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := repo.Delete(ctx, user.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := repo.FindByEmail(ctx, email)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got: %v", err)
	}
}
