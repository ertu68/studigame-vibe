// Package db opens the SQLite database and applies migrations.
package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Open opens the SQLite database and applies pending migrations.
func Open(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)", path)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	// SQLite allows one writer at a time; a single connection avoids SQLITE_BUSY
	// under concurrent goroutines without needing a connection-pool tuneup yet.
	conn.SetMaxOpenConns(1)

	if err := applyMigrations(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("apply migrations: %w", err)
	}
	return conn, nil
}

// applyMigrations applies all pending migrations in order.
func applyMigrations(db *sql.DB) error {
	// Ensure migrations table exists - CREATE TABLE IF NOT EXISTS to handle existing table
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP)`); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	// Get list of applied migrations
	applied := make(map[string]bool)
	rows, err := db.Query(`SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return fmt.Errorf("query applied migrations: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("scan migration version: %w", err)
		}
		applied[version] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate migrations: %w", err)
	}

	// Get all migration files
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	// Sort by filename to ensure correct order
	filenames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			filenames = append(filenames, entry.Name())
		}
	}
	sort.Strings(filenames)

	// Apply pending migrations
	for _, filename := range filenames {
		version := strings.TrimSuffix(filename, ".sql")
		if applied[version] {
			continue
		}

		// Read migration content
		content, err := fs.ReadFile(migrationsFS, "migrations/"+filename)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", filename, err)
		}

		// Apply migration in a transaction
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin transaction for %s: %w", filename, err)
		}

		// Execute migration SQL
		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", filename, err)
		}

		// Record successful application
		if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", filename, err)
		}

		fmt.Printf("Applied migration: %s\n", filename)
	}

	return nil
}
