package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/he-end/verify-reverse/verify/repository"
)

type VerificationCode struct {
	bun.BaseModel `bun:"table:verification_codes,alias:vc"`
	ID            uuid.UUID  `bun:"id,pk,type:uuid"`
	Contact       string     `bun:"contact,notnull"`
	ContactType   string     `bun:"contact_type,notnull"`
	Code          string     `bun:"code,notnull"`
	Name          string     `bun:"name,notnull"`
	PasswordHash  string     `bun:"password_hash"`
	ExpiresAt     time.Time  `bun:"expires_at,notnull"`
	CreatedAt     time.Time  `bun:"created_at,notnull"`
	UsedAt        *time.Time `bun:"used_at"`
	UserID        *uuid.UUID `bun:"user_id"`
	IsPhantom     bool       `bun:"is_phantom,notnull,default:false"`
}

type VerificationRepository struct {
	*repository.BaseRepository[VerificationCode]
	db *bun.DB
}

func NewVerificationRepository(db *bun.DB) *VerificationRepository {
	return &VerificationRepository{
		BaseRepository: repository.NewBaseRepository[VerificationCode](db),
		db:             db,
	}
}

func (r *VerificationRepository) FindByCodeAndContact(ctx context.Context, code, contact string) (*VerificationCode, error) {
	var vc VerificationCode
	err := r.db.NewSelect().Model(&vc).
		Where("vc.code = ?", code).
		Where("vc.contact = ?", contact).
		Where("vc.used_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, repository.MapError(err)
	}
	if vc.ExpiresAt.Before(time.Now()) {
		return nil, repository.ErrVerificationExpired
	}
	return &vc, nil
}

func (r *VerificationRepository) MarkUsed(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	now := time.Now()
	_, err := r.db.NewUpdate().Model((*VerificationCode)(nil)).
		Set("used_at = ?", now).
		Set("user_id = ?", userID).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *VerificationRepository) DeleteExpired(ctx context.Context) (int64, error) {
	res, err := r.db.NewDelete().Model((*VerificationCode)(nil)).
		Where("expires_at < now()").
		Where("used_at IS NULL").
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *VerificationRepository) ExistsPending(ctx context.Context, contact, contactType string) (bool, error) {
	exists, err := r.db.NewSelect().Model((*VerificationCode)(nil)).
		Where("contact = ?", contact).
		Where("contact_type = ?", contactType).
		Where("used_at IS NULL").
		Where("expires_at > now()").
		Exists(ctx)
	if err != nil {
		return false, repository.MapError(err)
	}
	return exists, nil
}
