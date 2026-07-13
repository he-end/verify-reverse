package auth

import (
	"context"

	"github.com/uptrace/bun"

	"github.com/he-end/verify-reverse/auth/repository"
)

type AuthRepository struct {
	*repository.BaseRepository[User]
	db *bun.DB
}

func NewAuthRepository(db *bun.DB) *AuthRepository {
	return &AuthRepository{
		BaseRepository: repository.NewBaseRepository[User](db),
		db:             db,
	}
}

func (r *AuthRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.NewSelect().Model(&user).Where("u.email = ?", email).Scan(ctx)
	if err != nil {
		return nil, repository.MapError(err)
	}
	return &user, nil
}

func (r *AuthRepository) FindByNumber(ctx context.Context, number string) (*User, error) {
	var user User
	err := r.db.NewSelect().Model(&user).Where("u.number = ?", number).Scan(ctx)
	if err != nil {
		return nil, repository.MapError(err)
	}
	return &user, nil
}

func (r *AuthRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	exists, err := r.db.NewSelect().Model((*User)(nil)).Where("u.email = ?", email).Exists(ctx)
	if err != nil {
		return false, repository.MapError(err)
	}
	return exists, nil
}

func (r *AuthRepository) ExistsByNumber(ctx context.Context, number string) (bool, error) {
	exists, err := r.db.NewSelect().Model((*User)(nil)).Where("u.number = ?", number).Exists(ctx)
	if err != nil {
		return false, repository.MapError(err)
	}
	return exists, nil
}
