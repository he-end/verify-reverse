package auth

import (
	"context"
	"time"

	"github.com/uptrace/bun"

	"github.com/he-end/verify-reverse/verify/repository"
)

type VerificationAttempt struct {
	bun.BaseModel `bun:"table:verification_attempts,alias:va"`
	Contact       string     `bun:"contact,pk"`
	ContactType   string     `bun:"contact_type,pk"`
	Attempts      int        `bun:"attempts,notnull,default:1"`
	LastAttempt   time.Time  `bun:"last_attempt,notnull"`
	BlockedUntil  *time.Time `bun:"blocked_until"`
}

type AttemptRepository struct {
	*repository.BaseRepository[VerificationAttempt]
	db *bun.DB
}

func NewAttemptRepository(db *bun.DB) *AttemptRepository {
	return &AttemptRepository{
		BaseRepository: repository.NewBaseRepository[VerificationAttempt](db),
		db:             db,
	}
}

func (r *AttemptRepository) RecordFailed(ctx context.Context, contact, contactType string) error {
	now := time.Now()
	thirtyMin := now.Add(30 * time.Minute)
	twoHr := now.Add(2 * time.Hour)
	twentyFourHr := now.Add(24 * time.Hour)

	_, err := r.db.NewInsert().
		Model(&VerificationAttempt{
			Contact:     contact,
			ContactType: contactType,
			Attempts:    1,
			LastAttempt: now,
		}).
		On("CONFLICT (contact, contact_type) DO UPDATE").
		Set("attempts = va.attempts + 1").
		Set("last_attempt = EXCLUDED.last_attempt").
		Set("blocked_until = CASE WHEN va.attempts + 1 = 5 THEN ?::timestamptz WHEN va.attempts + 1 = 10 THEN ?::timestamptz WHEN va.attempts + 1 >= 15 THEN ?::timestamptz ELSE va.blocked_until END",
			thirtyMin, twoHr, twentyFourHr).
		Exec(ctx)
	return repository.MapError(err)
}

func (r *AttemptRepository) IsBlocked(ctx context.Context, contact, contactType string) (bool, error) {
	exists, err := r.db.NewSelect().
		Model((*VerificationAttempt)(nil)).
		Where("contact = ?", contact).
		Where("contact_type = ?", contactType).
		Where("blocked_until > NOW()").
		Exists(ctx)
	if err != nil {
		return false, repository.MapError(err)
	}
	return exists, nil
}

func (r *AttemptRepository) ResetAttempts(ctx context.Context, contact, contactType string) error {
	_, err := r.db.NewDelete().
		Model((*VerificationAttempt)(nil)).
		Where("contact = ?", contact).
		Where("contact_type = ?", contactType).
		Exec(ctx)
	return repository.MapError(err)
}
