// cmd/seed/rss_seed/main.go
package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
)

//go:embed rss.json
var rssJSON []byte

type SeedCategory struct {
	ID         string     `json:"id"` // hanya untuk mapping internal seed
	Name       string     `json:"name"`
	DeletedAt  *time.Time `json:"deletedAt"`
	CreatedAt  string     `json:"createdAt"`
	UpdatedAt  string     `json:"updatedAt"`
	MasterRsss []SeedRSS  `json:"masterRsses"`
}

type SeedRSS struct {
	ID                  string     `json:"id"` // tidak dipakai untuk PK DB
	Title               string     `json:"title"`
	URL                 string     `json:"url"`
	Publisher           string     `json:"publisher"`
	MasterRssCategoryID string     `json:"masterRssCategoryId"` // tidak perlu, sudah nested
	DeletedAt           *time.Time `json:"deletedAt"`
	CreatedAt           string     `json:"createdAt"`
	UpdatedAt           string     `json:"updatedAt"`
}

func main() {
	_ = godotenv.Load()

	var (
		dsn           = flag.String("dsn", os.Getenv("DATABASE_URL"), "Postgres DSN (default: env DATABASE_URL)")
		runMigrate    = flag.Bool("migrate", true, "Run goose up migrations before seeding")
		migrationsDir = flag.String("migrations-dir", "./migrations", "Path to migrations directory")
	)
	flag.Parse()

	if strings.TrimSpace(*dsn) == "" {
		log.Fatal("DATABASE_URL empty. Set env DATABASE_URL or pass -dsn.")
	}

	ctx := context.Background()

	db, err := sql.Open("pgx", *dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	if *runMigrate {
		if err := runGooseUp(db, *migrationsDir); err != nil {
			log.Fatalf("goose up failed: %v", err)
		}
	}

	var categories []SeedCategory
	if err := json.Unmarshal(rssJSON, &categories); err != nil {
		log.Fatalf("unmarshal rss.json: %v", err)
	}

	if err := seedRSS(ctx, db, categories); err != nil {
		log.Fatalf("seed failed: %v", err)
	}

	log.Println("✅ rss seed done")
}

func runGooseUp(db *sql.DB, migrationsDir string) error {
	goose.SetDialect("postgres")
	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("goose.Up(%s): %w", migrationsDir, err)
	}
	return nil
}

func seedRSS(ctx context.Context, db *sql.DB, categories []SeedCategory) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// mapping seed category id (string) => DB category id (int64)
	seedCatIDToDBCatID := make(map[string]int64)

	var insertedCat, updatedCat, insertedRSS, updatedRSS int

	for _, c := range categories {
		catName := strings.TrimSpace(c.Name)
		if catName == "" {
			return fmt.Errorf("category name empty for seed id=%s", c.ID)
		}

		// 1) find by name (dedupe)
		dbCatID, found, err := findCategoryIDByName(ctx, tx, catName)
		if err != nil {
			return err
		}

		if found {
			// update existing category
			if err := updateCategory(ctx, tx, dbCatID, catName, c.DeletedAt); err != nil {
				return err
			}
			updatedCat++
		} else {
			// insert new category, let DB generate id
			newID, err := insertCategory(ctx, tx, catName, c.DeletedAt)
			if err != nil {
				// kalau race/seed ulang dan ada unique constraint name, fallback ke find+update
				if isUniqueViolation(err) {
					dbCatID, found2, err2 := findCategoryIDByName(ctx, tx, catName)
					if err2 != nil {
						return err2
					}
					if !found2 {
						return fmt.Errorf("unique violation but category not found by name: %w", err)
					}
					if err := updateCategory(ctx, tx, dbCatID, catName, c.DeletedAt); err != nil {
						return err
					}
					updatedCat++
				} else {
					return err
				}
			} else {
				dbCatID = newID
				insertedCat++
			}
		}

		seedCatIDToDBCatID[c.ID] = dbCatID

		// seed RSS items under category (FK pasti nyambung karena pakai dbCatID)
		for _, r := range c.MasterRsss {
			title := strings.TrimSpace(r.Title)
			url := strings.TrimSpace(r.URL)
			publisher := strings.TrimSpace(r.Publisher)
			if title == "" || url == "" || publisher == "" {
				return fmt.Errorf("rss has empty field (seed id=%s title=%q url=%q publisher=%q)", r.ID, title, url, publisher)
			}

			// dedupe RSS by url (lebih stabil daripada seed UUID)
			if existingID, ok, err := findRssIDByURL(ctx, tx, url); err != nil {
				return err
			} else if ok {
				if err := updateRssByID(ctx, tx, existingID, title, url, publisher, dbCatID, r.DeletedAt); err != nil {
					return err
				}
				updatedRSS++
				continue
			}

			if _, err := insertRss(ctx, tx, title, url, publisher, dbCatID, r.DeletedAt); err != nil {
				// kalau url unique dan seed ulang
				if isUniqueViolation(err) {
					existingID, ok, err2 := findRssIDByURL(ctx, tx, url)
					if err2 != nil {
						return err2
					}
					if !ok {
						return fmt.Errorf("unique violation but rss not found by url: %w", err)
					}
					if err := updateRssByID(ctx, tx, existingID, title, url, publisher, dbCatID, r.DeletedAt); err != nil {
						return err
					}
					updatedRSS++
					continue
				}
				return err
			}
			insertedRSS++
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	log.Printf("✅ categories: inserted=%d updated=%d | rss: inserted=%d updated=%d\n",
		insertedCat, updatedCat, insertedRSS, updatedRSS)

	_ = seedCatIDToDBCatID // kalau mau dipakai debugging/logging
	return nil
}

func findCategoryIDByName(ctx context.Context, tx *sql.Tx, name string) (int64, bool, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM app_rss_categories
		WHERE lower(name) = lower($1)
		LIMIT 1
	`, name).Scan(&id)
	if err == nil {
		return id, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	return 0, false, fmt.Errorf("find category by name: %w", err)
}

func insertCategory(ctx context.Context, tx *sql.Tx, name string, deletedAt *time.Time) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
		INSERT INTO app_rss_categories (name, deleted_at)
		VALUES ($1, $2)
		RETURNING id
	`, name, deletedAt).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert category: %w", err)
	}
	return id, nil
}

func updateCategory(ctx context.Context, tx *sql.Tx, id int64, name string, deletedAt *time.Time) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE app_rss_categories
		SET name = $2,
		    deleted_at = $3
		WHERE id = $1
	`, id, name, deletedAt)
	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}
	return nil
}

func findRssIDByURL(ctx context.Context, tx *sql.Tx, url string) (int64, bool, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM app_rss_feeds
		WHERE url = $1
		LIMIT 1
	`, url).Scan(&id)
	if err == nil {
		return id, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	return 0, false, fmt.Errorf("find rss by url: %w", err)
}

func insertRss(ctx context.Context, tx *sql.Tx, title, url, publisher string, categoryID int64, deletedAt *time.Time) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
		INSERT INTO app_rss_feeds (title, url, publisher, app_rss_category_id, deleted_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, title, url, publisher, categoryID, deletedAt).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert rss: %w", err)
	}
	return id, nil
}

func updateRssByID(ctx context.Context, tx *sql.Tx, id int64, title, url, publisher string, categoryID int64, deletedAt *time.Time) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE app_rss_feeds
		SET title = $2,
		    url = $3,
		    publisher = $4,
		    app_rss_category_id = $5,
		    deleted_at = $6
		WHERE id = $1
	`, id, title, url, publisher, categoryID, deletedAt)
	if err != nil {
		return fmt.Errorf("update rss: %w", err)
	}
	return nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "duplicate key") || strings.Contains(s, "unique constraint") || strings.Contains(s, "23505")
}
