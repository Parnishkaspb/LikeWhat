// Package db is the single point of access to PostgreSQL shared by all
// repositories. Repositories embed Client to reuse the connection pool and
// a unified "not found" mapping instead of repeating that plumbing.
package db

import (
	"context"
	"errors"
	"time"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/errs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// queryTimeout bounds every store query independently of the incoming
// request context, so a slow statement cannot hold a request forever.
const queryTimeout = 5 * time.Second

// Client owns the pgx connection pool and exposes the query primitives
// repositories need, hiding the concrete pool type behind one object.
type Client struct {
	pool *pgxpool.Pool
}

func NewClient(pool *pgxpool.Pool) *Client {
	return &Client{pool: pool}
}

// Row is a one-row query result that releases its timeout context on Scan.
type Row struct {
	row    pgx.Row
	cancel context.CancelFunc
}

// Scan delegates to the underlying row and releases the query timeout.
func (r *Row) Scan(dest ...any) error {
	defer r.cancel()
	return r.row.Scan(dest...)
}

// QueryRow runs a one-row query. Rows that do not match anything surface as
// errs.ErrNotFound via NotFound, so every repository returns the same sentinel.
func (c *Client) QueryRow(ctx context.Context, query string, args ...any) *Row {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	return &Row{row: c.pool.QueryRow(ctx, query, args...), cancel: cancel}
}

// Query runs a multi-row query. The returned rows carry the timeout context;
// it is released when the caller closes the rows.
func (c *Client) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	rows, err := c.pool.Query(ctx, query, args...)
	if err != nil {
		cancel()
		return nil, err
	}
	return &Rows{Rows: rows, cancel: cancel}, nil
}

// Rows wraps query results so closing them releases the timeout context.
type Rows struct {
	pgx.Rows
	cancel context.CancelFunc
}

func (r *Rows) Close() {
	r.cancel()
	r.Rows.Close()
}

func (c *Client) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
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
