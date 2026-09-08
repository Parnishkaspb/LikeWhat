// Package db is the single point of access to PostgreSQL shared by all
// repositories. Repositories embed Client to reuse the connection pool and
// a unified "not found" mapping instead of repeating that plumbing.
package db

import (
	"context"
	"errors"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/errs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Client owns the pgx connection pool and exposes the query primitives
// repositories need, hiding the concrete pool type behind one object.
type Client struct {
	pool *pgxpool.Pool
}

func NewClient(pool *pgxpool.Pool) *Client {
	return &Client{pool: pool}
}

// QueryRow runs a one-row query. Rows that do not match anything surface as
// errs.ErrNotFound via NotFound, so every repository returns the same sentinel.
func (c *Client) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	return c.pool.QueryRow(ctx, query, args...)
}

func (c *Client) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	return c.pool.Query(ctx, query, args...)
}

func (c *Client) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	return c.pool.Exec(ctx, query, args...)
}

// ErrNotFound is the sentinel returned by Store methods when a record is missing.
var ErrNotFound = errs.ErrNotFound

// NotFound converts a no-row result into errs.ErrNotFound and leaves any other
// error untouched.
func NotFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrNotFound
	}
	return err
}
