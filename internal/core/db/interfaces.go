package db

import "context"

// DB is the core interface for our database connection.
// It can execute queries or spawn transactions.
type DB interface {
	Executor
	RunInTransaction(ctx context.Context, fn func(tx Transaction) error) error
}

// Transaction represents an active database transaction.
// It has the same capabilities as the main Executor.
type Transaction interface {
	Executor
}

// Executor defines the generic methods for raw SQL execution.
// By using named parameters (e.g. {:id}), the underlying implementation
// can translate them to PostgreSQL ($1) or SQLite (?) automatically.
type Executor interface {
	Execute(ctx context.Context, query string, params map[string]any) error
	QueryRow(ctx context.Context, query string, params map[string]any, dest any) error
	QueryRows(ctx context.Context, query string, params map[string]any, dest any) error
}
