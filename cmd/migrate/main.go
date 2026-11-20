package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	dir := os.Getenv("MIGRATIONS_DIR")
	if dir == "" {
		dir = "db/migrations"
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()
	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS public.schema_migrations (
		version text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		log.Fatalf("migration table: %v", err)
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		log.Fatalf("scan migrations: %v", err)
	}
	sort.Strings(files)
	if len(files) == 0 {
		log.Fatalf("no migrations found in %s", dir)
	}
	applied := 0
	for _, file := range files {
		version := strings.TrimSuffix(filepath.Base(file), ".sql")
		var exists bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM public.schema_migrations WHERE version=$1)`,
			version).Scan(&exists); err != nil {
			log.Fatalf("check %s: %v", version, err)
		}
		if exists {
			continue
		}
		raw, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("read %s: %v", file, err)
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			log.Fatalf("begin %s: %v", version, err)
		}
		if _, err := tx.Exec(ctx, string(raw)); err != nil {
			_ = tx.Rollback(ctx)
			log.Fatalf("apply %s: %v", version, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO public.schema_migrations (version) VALUES ($1)`, version); err != nil {
			_ = tx.Rollback(ctx)
			log.Fatalf("record %s: %v", version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			log.Fatalf("commit %s: %v", version, err)
		}
		fmt.Printf("applied %s\n", version)
		applied++
	}
	if applied == 0 {
		fmt.Println("database is up to date")
	}
}
