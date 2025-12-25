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

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed rss.json
var rssJSON []byte

// JSON shape (sesuai TS types kamu)
type SeedCategory struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	DeletedAt  *time.Time `json:"deletedAt"`
	CreatedAt  string     `json:"createdAt"` // diabaikan (DB handle)
	UpdatedAt  string     `json:"updatedAt"` // diabaikan (DB handle)
	MasterRsss []SeedRSS  `json:"masterRsses"`
}

type SeedRSS struct {
	ID                  string     `json:"id"`
	Title               string     `json:"title"`
	URL                 string     `json:"url"`
	Publisher           string     `json:"publisher"`
	MasterRssCategoryID string     `json:"masterRssCategoryId"`
	DeletedAt           *time.Time `json:"deletedAt"`
	CreatedAt           string     `json:"createdAt"` // diabaikan (DB handle)
	UpdatedAt           string     `json:"updatedAt"` // diabaikan (DB handle)
}

func main() {
	var (
		dsn           = flag.String("dsn", os.Getenv("DATABASE_URL"), "Postgres DSN (default: env DATABASE_URL)")
		runMigrate    = flag.Bool("migrate", true, "Run goose up migrations before seeding")
		migrationsDir = flag.String("migrations-dir", "./migrations", "Path to migrations directory")
	)
	flag.Parse()

	fmt.Println("dsn:", *dsn)

	if strings.TrimSpace(*dsn) == "" {
		log.Fatal("DATABASE_URL empty. Set env DATABASE_URL or pass -dsn.")
	}
	ctx := context.Background()

	db, err := sql.Open("pgx", *dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

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
	// goose.Up akan jalankan semua migration yang belum applied
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

	// map untuk memastikan RSS yang insert pakai category id yang benar (kalau nama sudah ada di DB)
	seedCatIDToDBCatID := make(map[string]uuid.UUID)

	var insertedCat, updatedCat, insertedRSS, updatedRSS int

	for _, c := range categories {
		catSeedID, err := uuid.Parse(c.ID)
		if err != nil {
			return fmt.Errorf("invalid category id %q: %w", c.ID, err)
		}
		catName := strings.TrimSpace(c.Name)
		if catName == "" {
			return fmt.Errorf("category name empty for id=%s", c.ID)
		}

		dbCatID, found, err := findCategoryIDByName(ctx, tx, catName)
		if err != nil {
			return err
		}

		switch {
		case found:
			// update existing category (avoid duplicate by name)
			if err := updateCategory(ctx, tx, dbCatID, catName, c.DeletedAt); err != nil {
				return err
			}
			updatedCat++
			seedCatIDToDBCatID[catSeedID.String()] = dbCatID

		default:
			// insert new category (use seed id)
			if err := insertCategory(ctx, tx, catSeedID, catName, c.DeletedAt); err != nil {
				// kalau ternyata id sudah ada (misalnya pernah seed), update aja
				if isUniqueViolation(err) {
					if err2 := updateCategory(ctx, tx, catSeedID, catName, c.DeletedAt); err2 != nil {
						return err2
					}
					updatedCat++
				} else {
					return err
				}
			} else {
				insertedCat++
			}
			seedCatIDToDBCatID[catSeedID.String()] = catSeedID
		}

		// seed RSS items under category
		for _, r := range c.MasterRsss {
			rssSeedID, err := uuid.Parse(r.ID)
			if err != nil {
				return fmt.Errorf("invalid rss id %q: %w", r.ID, err)
			}
			title := strings.TrimSpace(r.Title)
			url := strings.TrimSpace(r.URL)
			publisher := strings.TrimSpace(r.Publisher)
			if title == "" || url == "" || publisher == "" {
				return fmt.Errorf("rss has empty field (id=%s title=%q url=%q publisher=%q)", r.ID, title, url, publisher)
			}

			// pilih category id yang sudah “dinormalisasi” (by name dedupe)
			dbCatID := seedCatIDToDBCatID[catSeedID.String()]

			// 1) kalau id sudah ada -> update by id
			if existingID, ok, err := findRssIDByID(ctx, tx, rssSeedID); err != nil {
				return err
			} else if ok {
				if err := updateRssByID(ctx, tx, existingID, title, url, publisher, dbCatID, r.DeletedAt); err != nil {
					return err
				}
				updatedRSS++
				continue
			}

			// 2) kalau url sudah ada -> update by url (avoid duplicate by url)
			if existingID, ok, err := findRssIDByURL(ctx, tx, url); err != nil {
				return err
			} else if ok {
				if err := updateRssByID(ctx, tx, existingID, title, url, publisher, dbCatID, r.DeletedAt); err != nil {
					return err
				}
				updatedRSS++
				continue
			}

			// 3) else insert baru
			if err := insertRss(ctx, tx, rssSeedID, title, url, publisher, dbCatID, r.DeletedAt); err != nil {
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

	return nil
}

func findCategoryIDByName(ctx context.Context, tx *sql.Tx, name string) (uuid.UUID, bool, error) {
	var id uuid.UUID
	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM app_rss_categories
		WHERE deleted_at IS NULL AND lower(name) = lower($1)
		LIMIT 1
	`, name).Scan(&id)
	if err == nil {
		return id, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.UUID{}, false, nil
	}
	return uuid.UUID{}, false, fmt.Errorf("find category by name: %w", err)
}

func insertCategory(ctx context.Context, tx *sql.Tx, id uuid.UUID, name string, deletedAt *time.Time) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO app_rss_categories (id, name, deleted_at)
		VALUES ($1, $2, $3)
	`, id, name, deletedAt)
	if err != nil {
		return fmt.Errorf("insert category: %w", err)
	}
	return nil
}

func updateCategory(ctx context.Context, tx *sql.Tx, id uuid.UUID, name string, deletedAt *time.Time) error {
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

func findRssIDByID(ctx context.Context, tx *sql.Tx, id uuid.UUID) (uuid.UUID, bool, error) {
	var got uuid.UUID
	err := tx.QueryRowContext(ctx, `
		SELECT id FROM app_rss_feeds WHERE id = $1 LIMIT 1
	`, id).Scan(&got)
	if err == nil {
		return got, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.UUID{}, false, nil
	}
	return uuid.UUID{}, false, fmt.Errorf("find rss by id: %w", err)
}

func findRssIDByURL(ctx context.Context, tx *sql.Tx, url string) (uuid.UUID, bool, error) {
	var id uuid.UUID
	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM app_rss_feeds
		WHERE deleted_at IS NULL AND url = $1
		LIMIT 1
	`, url).Scan(&id)
	if err == nil {
		return id, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.UUID{}, false, nil
	}
	return uuid.UUID{}, false, fmt.Errorf("find rss by url: %w", err)
}

func insertRss(ctx context.Context, tx *sql.Tx, id uuid.UUID, title, url, publisher string, categoryID uuid.UUID, deletedAt *time.Time) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO app_rss_feeds (id, title, url, publisher, app_rss_category_id, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, title, url, publisher, categoryID, deletedAt)
	if err != nil {
		return fmt.Errorf("insert rss: %w", err)
	}
	return nil
}

func updateRssByID(ctx context.Context, tx *sql.Tx, id uuid.UUID, title, url, publisher string, categoryID uuid.UUID, deletedAt *time.Time) error {
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

// Sederhana: deteksi unique violation via string (driver bisa beda-beda).
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "duplicate key") || strings.Contains(s, "unique constraint") || strings.Contains(s, "23505")
}
