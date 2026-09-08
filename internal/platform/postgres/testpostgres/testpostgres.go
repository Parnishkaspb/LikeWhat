// Package testpostgres provisions a disposable PostgreSQL for tests.
//
// Tests are backed by a real database instead of an in-memory repository.
// Each test binary starts one PostgreSQL container via testcontainers,
// applies the project migrations, and shares a single pool across the package.
package testpostgres

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	pool *pgxpool.Pool
	stop func()
)

// Main wraps m.Run with a single disposable PostgreSQL lifecycle. Call it from
// a package TestMain:
//
//	func TestMain(m *testing.M) { os.Exit(testpostgres.Main(m)) }
func Main(m *testing.M) int {
	start()
	defer shutdown()
	return m.Run()
}

// Pool returns the shared pool created by Main.
func Pool() *pgxpool.Pool { return pool }

func start() {
	ctx := context.Background()

	container, err := postgres.Run(ctx, "postgres:16-alpine")
	if err != nil {
		panic("testpostgres: start container: " + err.Error())
	}
	stop = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = container.Terminate(ctx)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("testpostgres: connection string: " + err.Error())
	}

	pool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		panic("testpostgres: open pool: " + err.Error())
	}

	applyMigrations(dsn)
}

func shutdown() {
	if pool != nil {
		pool.Close()
	}
	if stop != nil {
		stop()
	}
}

func applyMigrations(dsn string) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		panic("testpostgres: open migrations connection: " + err.Error())
	}
	defer db.Close()

	// internal/platform/postgres/testpostgres -> repo root migrations dir.
	fsys := os.DirFS(filepath.Join("..", "..", "..", "..", "migrations"))
	provider, err := goose.NewProvider(goose.DialectPostgres, db, fsys)
	if err != nil {
		panic("testpostgres: migrations provider: " + err.Error())
	}
	if _, err := provider.Up(context.Background()); err != nil {
		panic("testpostgres: apply migrations: " + err.Error())
	}
}
