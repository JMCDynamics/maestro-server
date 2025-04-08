package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/JMCDynamics/maestro-server/internal/interfaces"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type postgreDatabaseAdapter struct {
	pool             *pgxpool.Pool
	connectionString string
}

type PGXResultRow struct {
	rows pgx.Rows
}

func (r *PGXResultRow) Next() bool {
	return r.rows.Next()
}

func (r *PGXResultRow) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

func (r *PGXResultRow) Close() {
	r.rows.Close()
}

func (r *PGXResultRow) Err() error {
	return r.rows.Err()
}

func NewDatabaseGateway(connString string) (interfaces.IDatabaseGateway, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	return &postgreDatabaseAdapter{pool: pool, connectionString: connString}, nil
}

func (pg *postgreDatabaseAdapter) Query(ctx context.Context, query string, args ...any) (interfaces.ResultSet, error) {
	rows, err := pg.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &PGXResultRow{
		rows: rows,
	}, nil
}

func (pg *postgreDatabaseAdapter) Exec(ctx context.Context, query string, args ...any) error {
	_, err := pg.pool.Exec(ctx, query, args...)
	return err
}

func (pg *postgreDatabaseAdapter) QueryRow(ctx context.Context, query string, dest any, args ...any) error {
	row := pg.pool.QueryRow(ctx, query, args...)
	return row.Scan(&dest)
}

func (pg *postgreDatabaseAdapter) Close() {
	pg.pool.Close()
}

func (pg *postgreDatabaseAdapter) RunMigrations() error {
	// db, err := sql.Open("postgres", pg.connectionString)
	// if err != nil {
	// 	return fmt.Errorf("failed to open database: %w", err)
	// }
	// defer db.Close()

	// driver, err := postgres.WithInstance(db, &postgres.Config{})
	// if err != nil {
	// 	return fmt.Errorf("failed to create driver instance: %w", err)
	// }

	// m, err := migrate.(
	// 	"file://../../migrations",
	// 	driver,
	// 	pg.connectionString,
	// )

	// if err != nil {
	// 	return fmt.Errorf("failed to create migration instance: %w", err)
	// }

	// if err := m.Up(); err != nil && err != migrate.ErrNoChange {
	// 	return fmt.Errorf("failed to run migrations: %w", err)
	// }

	return nil
}
