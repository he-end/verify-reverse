package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/driver/pgdriver"
)

type QueryOption func(*bun.SelectQuery) *bun.SelectQuery

type Repository[T any] interface {
	FindByID(ctx context.Context, id any) (*T, error)
	FindAll(ctx context.Context, opts ...QueryOption) ([]T, error)
	FindOne(ctx context.Context, opts ...QueryOption) (*T, error)
	Create(ctx context.Context, entity *T) error
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id any) error
	Count(ctx context.Context, opts ...QueryOption) (int, error)
}

type BaseRepository[T any] struct {
	DB *bun.DB
}

func NewBaseRepository[T any](db *bun.DB) *BaseRepository[T] {
	return &BaseRepository[T]{DB: db}
}

func (r *BaseRepository[T]) FindByID(ctx context.Context, id any) (*T, error) {
	var entity T
	err := r.DB.NewSelect().Model(&entity).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, MapError(err)
	}
	return &entity, nil
}

func (r *BaseRepository[T]) FindAll(ctx context.Context, opts ...QueryOption) ([]T, error) {
	var entities []T
	q := r.DB.NewSelect().Model(&entities)
	for _, opt := range opts {
		q = opt(q)
	}
	err := q.Scan(ctx)
	if err != nil {
		return nil, MapError(err)
	}
	return entities, nil
}

func (r *BaseRepository[T]) FindOne(ctx context.Context, opts ...QueryOption) (*T, error) {
	var entity T
	q := r.DB.NewSelect().Model(&entity)
	for _, opt := range opts {
		q = opt(q)
	}
	err := q.Limit(1).Scan(ctx)
	if err != nil {
		return nil, MapError(err)
	}
	return &entity, nil
}

func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	_, err := r.DB.NewInsert().Model(entity).Exec(ctx)
	return MapError(err)
}

func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) error {
	_, err := r.DB.NewUpdate().Model(entity).WherePK().Exec(ctx)
	return MapError(err)
}

func (r *BaseRepository[T]) Delete(ctx context.Context, id any) error {
	var entity T
	_, err := r.DB.NewDelete().Model(&entity).Where("id = ?", id).Exec(ctx)
	return MapError(err)
}

func (r *BaseRepository[T]) Count(ctx context.Context, opts ...QueryOption) (int, error) {
	var entity T
	q := r.DB.NewSelect().Model(&entity)
	for _, opt := range opts {
		q = opt(q)
	}
	count, err := q.Count(ctx)
	if err != nil {
		return 0, MapError(err)
	}
	return count, nil
}

func MapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr pgdriver.Error
	if errors.As(err, &pgErr) {
		if pgErr.Field('C') == "23505" {
			return ErrDuplicateKey
		}
	}
	return err
}
