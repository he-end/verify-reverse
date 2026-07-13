package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/he-end/verify-reverse/verify/repository"
)

type Session struct {
	bun.BaseModel `bun:"table:sessions,alias:s"`
	ID            uuid.UUID `bun:"id,pk,type:uuid"`
	UserID        uuid.UUID `bun:"user_id,notnull"`
	RefreshToken  string    `bun:"refresh_token,unique,notnull"`
	AccessTokenID string    `bun:"access_token_id,notnull"`
	ExpiresAt     time.Time `bun:"expires_at,notnull"`
	CreatedAt     time.Time `bun:"created_at,notnull"`
	User          *User     `bun:"rel:belongs-to,join:user_id=id"`
}

type SessionRepository struct {
	*repository.BaseRepository[Session]
	db *bun.DB
}

func NewSessionRepository(db *bun.DB) *SessionRepository {
	return &SessionRepository{
		BaseRepository: repository.NewBaseRepository[Session](db),
		db:             db,
	}
}

func (r *SessionRepository) FindByRefreshToken(ctx context.Context, token string) (*Session, error) {
	var session Session
	err := r.db.NewSelect().Model(&session).Where("s.refresh_token = ?", token).Scan(ctx)
	if err != nil {
		return nil, repository.MapError(err)
	}
	return &session, nil
}

func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.NewDelete().Model((*Session)(nil)).Where("user_id = ?", userID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete sessions by user: %w", err)
	}
	return nil
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	res, err := r.db.NewDelete().Model((*Session)(nil)).Where("expires_at < now()").Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("delete expired sessions: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *SessionRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := r.db.NewSelect().Model((*Session)(nil)).Where("user_id = ?", userID).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count sessions by user: %w", err)
	}
	return count, nil
}

func (r *SessionRepository) DeleteOldestByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.NewDelete().Model((*Session)(nil)).
		Where("id = (SELECT id FROM sessions WHERE user_id = ? ORDER BY created_at ASC LIMIT 1)", userID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete oldest session: %w", err)
	}
	return nil
}
