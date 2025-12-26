// cmd/seed/creator_image_seed/main.go
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
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
)

//go:embed creator_image.json
var creatorImageJSON []byte

type CreatorImageSeedItem struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	ImageURL    string    `json:"imageUrl"`
	IsPublished bool      `json:"isPublished"`
	Price       int       `json:"price"`
	PublisherID *string   `json:"publisherId"` // -> profile_id (UUID)
	DeletedAt   *string   `json:"deletedAt"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	TypeCats    []TypeCat `json:"templateImageCategories"`
	ProductCats []ProdCat `json:"templateProductCategories"`
}

type TypeCat struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProdCat struct {
	ID             string `json:"id"`
	IndonesianName string `json:"indonesianName"`
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
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	if *runMigrate {
		if err := runGooseUp(db, *migrationsDir); err != nil {
			log.Fatalf("goose up failed: %v", err)
		}
	}

	var items []CreatorImageSeedItem
	if err := json.Unmarshal(creatorImageJSON, &items); err != nil {
		log.Fatalf("unmarshal creator_image.json: %v", err)
	}

	if err := seedCreatorImages(ctx, db, items); err != nil {
		log.Fatalf("seed failed: %v", err)
	}

	log.Println("✅ creator_image seed done")
}

func runGooseUp(db *sql.DB, migrationsDir string) error {
	goose.SetDialect("postgres")
	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("goose.Up(%s): %w", migrationsDir, err)
	}
	return nil
}

func seedCreatorImages(ctx context.Context, db *sql.DB, items []CreatorImageSeedItem) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// cache supaya gak query berulang
	typeCatCache := map[string]int64{}    // key: lower(name)
	productCatCache := map[string]int64{} // key: lower(indonesian_name)

	var insertedImg, updatedImg int

	for _, it := range items {
		name := strings.TrimSpace(it.Name)
		imageURL := strings.TrimSpace(it.ImageURL)
		if name == "" || imageURL == "" {
			return fmt.Errorf("creator image invalid (id=%q name=%q imageUrl=%q)", it.ID, it.Name, it.ImageURL)
		}

		// publisherId -> profile_id (nullable)
		var profileID *uuid.UUID
		if it.PublisherID != nil && strings.TrimSpace(*it.PublisherID) != "" {
			u, err := uuid.Parse(strings.TrimSpace(*it.PublisherID))
			if err != nil {
				return fmt.Errorf("invalid publisherId=%q for image id=%q: %w", *it.PublisherID, it.ID, err)
			}
			// cek FK existence biar seed gak gagal
			ok, err := profileExists(ctx, tx, u)
			if err != nil {
				return err
			}
			if !ok {
				log.Printf("⚠️  profile not found id=%s (image id=%s). Setting profile_id=NULL.\n", u.String(), it.ID)
			} else {
				profileID = &u
			}
		}

		deletedAt, err := parseNullableTime(it.DeletedAt)
		if err != nil {
			return fmt.Errorf("parse deletedAt for image id=%q: %w", it.ID, err)
		}

		imgID, existed, err := upsertCreatorImage(ctx, tx, it, name, imageURL, profileID, deletedAt)
		if err != nil {
			return err
		}
		if existed {
			updatedImg++
		} else {
			insertedImg++
		}

		// sync pivots deterministik: delete existing then insert current
		if err := syncPivots(ctx, tx, imgID, it, typeCatCache, productCatCache); err != nil {
			return err
		}
	}

	// penting kalau kamu insert manual id ke BIGSERIAL
	if err := fixSequences(ctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	log.Printf("✅ creator_images: inserted=%d updated=%d\n", insertedImg, updatedImg)
	return nil
}

func profileExists(ctx context.Context, tx *sql.Tx, id uuid.UUID) (bool, error) {
	var one int
	err := tx.QueryRowContext(ctx, `SELECT 1 FROM profiles WHERE id = $1 LIMIT 1`, id).Scan(&one)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return false, fmt.Errorf("profileExists: %w", err)
}

func upsertCreatorImage(
	ctx context.Context,
	tx *sql.Tx,
	it CreatorImageSeedItem,
	name, imageURL string,
	profileID *uuid.UUID,
	deletedAt *time.Time,
) (dbID int64, existed bool, err error) {

	// 1) kalau seed id parseable dan sudah ada -> update by id
	if seedID, ok := parseSeedBigInt(it.ID); ok {
		if gotID, found, err := findCreatorImageIDByID(ctx, tx, seedID); err != nil {
			return 0, false, err
		} else if found {
			if err := updateCreatorImageByID(ctx, tx, gotID, name, imageURL, it.IsPublished, int64(it.Price), profileID, deletedAt); err != nil {
				return 0, false, err
			}
			return gotID, true, nil
		}
	}

	// 2) kalau image_url sudah ada -> update by image_url
	if gotID, found, err := findCreatorImageIDByImageURL(ctx, tx, imageURL); err != nil {
		return 0, false, err
	} else if found {
		if err := updateCreatorImageByID(ctx, tx, gotID, name, imageURL, it.IsPublished, int64(it.Price), profileID, deletedAt); err != nil {
			return 0, false, err
		}
		return gotID, true, nil
	}

	// 3) insert baru (pakai seed id kalau bisa)
	if seedID, ok := parseSeedBigInt(it.ID); ok {
		newID, err := insertCreatorImageWithID(ctx, tx, seedID, name, imageURL, it.IsPublished, int64(it.Price), profileID, deletedAt, it.CreatedAt, it.UpdatedAt)
		if err == nil {
			return newID, false, nil
		}
		// kalau duplicate id (misal pernah seed), fallback update by id
		if isUniqueViolation(err) {
			if err2 := updateCreatorImageByID(ctx, tx, seedID, name, imageURL, it.IsPublished, int64(it.Price), profileID, deletedAt); err2 != nil {
				return 0, false, err2
			}
			return seedID, true, nil
		}
		return 0, false, err
	}

	newID, err := insertCreatorImage(ctx, tx, name, imageURL, it.IsPublished, int64(it.Price), profileID, deletedAt, it.CreatedAt, it.UpdatedAt)
	if err != nil {
		return 0, false, err
	}
	return newID, false, nil
}

func findCreatorImageIDByID(ctx context.Context, tx *sql.Tx, id int64) (int64, bool, error) {
	var got int64
	err := tx.QueryRowContext(ctx, `SELECT id FROM creator_images WHERE id = $1 LIMIT 1`, id).Scan(&got)
	if err == nil {
		return got, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	return 0, false, fmt.Errorf("findCreatorImageIDByID: %w", err)
}

func findCreatorImageIDByImageURL(ctx context.Context, tx *sql.Tx, imageURL string) (int64, bool, error) {
	var got int64
	err := tx.QueryRowContext(ctx, `SELECT id FROM creator_images WHERE image_url = $1 LIMIT 1`, imageURL).Scan(&got)
	if err == nil {
		return got, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	return 0, false, fmt.Errorf("findCreatorImageIDByImageURL: %w", err)
}

func insertCreatorImageWithID(
	ctx context.Context,
	tx *sql.Tx,
	id int64,
	name, imageURL string,
	isPublished bool,
	price int64,
	profileID *uuid.UUID,
	deletedAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) (int64, error) {
	var pid any = nil
	if profileID != nil {
		pid = *profileID
	}
	var newID int64
	err := tx.QueryRowContext(ctx, `
		INSERT INTO creator_images (
			id, name, image_url, is_published, is_banned, banned_reason, price, profile_id, deleted_at, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4, FALSE, NULL, $5, $6, $7, $8, $9)
		RETURNING id
	`, id, name, imageURL, isPublished, price, pid, deletedAt, createdAt, updatedAt).Scan(&newID)
	if err != nil {
		return 0, fmt.Errorf("insertCreatorImageWithID: %w", err)
	}
	return newID, nil
}

func insertCreatorImage(
	ctx context.Context,
	tx *sql.Tx,
	name, imageURL string,
	isPublished bool,
	price int64,
	profileID *uuid.UUID,
	deletedAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) (int64, error) {
	var pid any = nil
	if profileID != nil {
		pid = *profileID
	}
	var newID int64
	err := tx.QueryRowContext(ctx, `
		INSERT INTO creator_images (
			name, image_url, is_published, is_banned, banned_reason, price, profile_id, deleted_at, created_at, updated_at
		)
		VALUES ($1,$2,$3, FALSE, NULL, $4, $5, $6, $7, $8)
		RETURNING id
	`, name, imageURL, isPublished, price, pid, deletedAt, createdAt, updatedAt).Scan(&newID)
	if err != nil {
		return 0, fmt.Errorf("insertCreatorImage: %w", err)
	}
	return newID, nil
}

func updateCreatorImageByID(
	ctx context.Context,
	tx *sql.Tx,
	id int64,
	name, imageURL string,
	isPublished bool,
	price int64,
	profileID *uuid.UUID,
	deletedAt *time.Time,
) error {
	var pid any = nil
	if profileID != nil {
		pid = *profileID
	}
	_, err := tx.ExecContext(ctx, `
		UPDATE creator_images
		SET name = $2,
		    image_url = $3,
		    is_published = $4,
		    -- schema baru:
		    is_banned = FALSE,
		    banned_reason = NULL,
		    price = $5,
		    profile_id = $6,
		    deleted_at = $7
		WHERE id = $1
	`, id, name, imageURL, isPublished, price, pid, deletedAt)
	if err != nil {
		return fmt.Errorf("updateCreatorImageByID: %w", err)
	}
	return nil
}

func syncPivots(
	ctx context.Context,
	tx *sql.Tx,
	creatorImageID int64,
	it CreatorImageSeedItem,
	typeCatCache map[string]int64,
	productCatCache map[string]int64,
) error {
	// clear current pivots
	if _, err := tx.ExecContext(ctx, `DELETE FROM creator_image_type_categories WHERE creator_image_id = $1`, creatorImageID); err != nil {
		return fmt.Errorf("delete type pivot: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM creator_image_product_categories WHERE creator_image_id = $1`, creatorImageID); err != nil {
		return fmt.Errorf("delete product pivot: %w", err)
	}

	// type categories
	seenType := map[int64]struct{}{}
	for _, c := range it.TypeCats {
		nm := strings.TrimSpace(c.Name)
		if nm == "" {
			continue
		}
		catID, err := ensureTypeCategory(ctx, tx, c, typeCatCache)
		if err != nil {
			return err
		}
		if _, ok := seenType[catID]; ok {
			continue
		}
		seenType[catID] = struct{}{}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO creator_image_type_categories (creator_image_id, type_category_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, creatorImageID, catID); err != nil {
			return fmt.Errorf("insert type pivot: %w", err)
		}
	}

	// product categories
	seenProd := map[int64]struct{}{}
	for _, c := range it.ProductCats {
		in := strings.TrimSpace(c.IndonesianName)
		if in == "" {
			continue
		}
		catID, err := ensureProductCategory(ctx, tx, c, productCatCache)
		if err != nil {
			return err
		}
		if _, ok := seenProd[catID]; ok {
			continue
		}
		seenProd[catID] = struct{}{}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO creator_image_product_categories (creator_image_id, product_category_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, creatorImageID, catID); err != nil {
			return fmt.Errorf("insert product pivot: %w", err)
		}
	}

	return nil
}

func ensureTypeCategory(ctx context.Context, tx *sql.Tx, seed TypeCat, cache map[string]int64) (int64, error) {
	name := strings.TrimSpace(seed.Name)
	key := strings.ToLower(name)
	if key == "" {
		return 0, fmt.Errorf("type category name empty")
	}
	if v, ok := cache[key]; ok {
		return v, nil
	}

	// 1) dedupe by name (case-insensitive)
	if id, found, err := findTypeCategoryIDByName(ctx, tx, name); err != nil {
		return 0, err
	} else if found {
		cache[key] = id
		return id, nil
	}

	// 2) insert by seed id if possible
	if seedID, ok := parseSeedBigInt(seed.ID); ok {
		id, err := insertTypeCategoryWithID(ctx, tx, seedID, name)
		if err == nil {
			cache[key] = id
			return id, nil
		}
		if isUniqueViolation(err) {
			// id exists -> update name
			if err2 := updateTypeCategoryByID(ctx, tx, seedID, name); err2 != nil {
				return 0, err2
			}
			cache[key] = seedID
			return seedID, nil
		}
		return 0, err
	}

	// 3) insert without id
	id, err := insertTypeCategory(ctx, tx, name)
	if err != nil {
		return 0, err
	}
	cache[key] = id
	return id, nil
}

func ensureProductCategory(ctx context.Context, tx *sql.Tx, seed ProdCat, cache map[string]int64) (int64, error) {
	indo := strings.TrimSpace(seed.IndonesianName)
	key := strings.ToLower(indo)
	if key == "" {
		return 0, fmt.Errorf("product category indonesianName empty")
	}
	if v, ok := cache[key]; ok {
		return v, nil
	}

	english := indo // sesuai permintaan: english_name = indonesianName

	// 1) dedupe by indonesian_name (case-insensitive)
	if id, found, err := findProductCategoryIDByIndoName(ctx, tx, indo); err != nil {
		return 0, err
	} else if found {
		// optional: update english_name biar konsisten
		if err := updateProductCategoryByID(ctx, tx, id, indo, english); err != nil {
			return 0, err
		}
		cache[key] = id
		return id, nil
	}

	// 2) insert by seed id if possible
	if seedID, ok := parseSeedBigInt(seed.ID); ok {
		id, err := insertProductCategoryWithID(ctx, tx, seedID, indo, english)
		if err == nil {
			cache[key] = id
			return id, nil
		}
		if isUniqueViolation(err) {
			if err2 := updateProductCategoryByID(ctx, tx, seedID, indo, english); err2 != nil {
				return 0, err2
			}
			cache[key] = seedID
			return seedID, nil
		}
		return 0, err
	}

	// 3) insert without id
	id, err := insertProductCategory(ctx, tx, indo, english)
	if err != nil {
		return 0, err
	}
	cache[key] = id
	return id, nil
}

func findTypeCategoryIDByName(ctx context.Context, tx *sql.Tx, name string) (int64, bool, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM app_creator_image_type_categories
		WHERE lower(name) = lower($1)
		LIMIT 1
	`, name).Scan(&id)
	if err == nil {
		return id, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	return 0, false, fmt.Errorf("findTypeCategoryIDByName: %w", err)
}

func insertTypeCategoryWithID(ctx context.Context, tx *sql.Tx, id int64, name string) (int64, error) {
	var got int64
	err := tx.QueryRowContext(ctx, `
		INSERT INTO app_creator_image_type_categories (id, name)
		VALUES ($1, $2)
		RETURNING id
	`, id, name).Scan(&got)
	if err != nil {
		return 0, fmt.Errorf("insertTypeCategoryWithID: %w", err)
	}
	return got, nil
}

func insertTypeCategory(ctx context.Context, tx *sql.Tx, name string) (int64, error) {
	var got int64
	err := tx.QueryRowContext(ctx, `
		INSERT INTO app_creator_image_type_categories (name)
		VALUES ($1)
		RETURNING id
	`, name).Scan(&got)
	if err != nil {
		return 0, fmt.Errorf("insertTypeCategory: %w", err)
	}
	return got, nil
}

func updateTypeCategoryByID(ctx context.Context, tx *sql.Tx, id int64, name string) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE app_creator_image_type_categories
		SET name = $2
		WHERE id = $1
	`, id, name)
	if err != nil {
		return fmt.Errorf("updateTypeCategoryByID: %w", err)
	}
	return nil
}

func findProductCategoryIDByIndoName(ctx context.Context, tx *sql.Tx, indo string) (int64, bool, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM app_creator_image_product_categories
		WHERE lower(indonesian_name) = lower($1)
		LIMIT 1
	`, indo).Scan(&id)
	if err == nil {
		return id, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	return 0, false, fmt.Errorf("findProductCategoryIDByIndoName: %w", err)
}

func insertProductCategoryWithID(ctx context.Context, tx *sql.Tx, id int64, indo, english string) (int64, error) {
	var got int64
	err := tx.QueryRowContext(ctx, `
		INSERT INTO app_creator_image_product_categories (id, indonesian_name, english_name)
		VALUES ($1, $2, $3)
		RETURNING id
	`, id, indo, english).Scan(&got)
	if err != nil {
		return 0, fmt.Errorf("insertProductCategoryWithID: %w", err)
	}
	return got, nil
}

func insertProductCategory(ctx context.Context, tx *sql.Tx, indo, english string) (int64, error) {
	var got int64
	err := tx.QueryRowContext(ctx, `
		INSERT INTO app_creator_image_product_categories (indonesian_name, english_name)
		VALUES ($1, $2)
		RETURNING id
	`, indo, english).Scan(&got)
	if err != nil {
		return 0, fmt.Errorf("insertProductCategory: %w", err)
	}
	return got, nil
}

func updateProductCategoryByID(ctx context.Context, tx *sql.Tx, id int64, indo, english string) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE app_creator_image_product_categories
		SET indonesian_name = $2,
		    english_name = $3
		WHERE id = $1
	`, id, indo, english)
	if err != nil {
		return fmt.Errorf("updateProductCategoryByID: %w", err)
	}
	return nil
}

// Kalau kamu insert manual id ke BIGSERIAL, ini penting biar sequence maju.
func fixSequences(ctx context.Context, tx *sql.Tx) error {
	stmts := []string{
		`SELECT setval(pg_get_serial_sequence('app_creator_image_type_categories','id'), COALESCE((SELECT MAX(id) FROM app_creator_image_type_categories), 1), true);`,
		`SELECT setval(pg_get_serial_sequence('app_creator_image_product_categories','id'), COALESCE((SELECT MAX(id) FROM app_creator_image_product_categories), 1), true);`,
		`SELECT setval(pg_get_serial_sequence('creator_images','id'), COALESCE((SELECT MAX(id) FROM creator_images), 1), true);`,
	}
	for _, q := range stmts {
		if _, err := tx.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("fixSequences: %w", err)
		}
	}
	return nil
}

func parseSeedBigInt(s string) (int64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func parseNullableTime(s *string) (*time.Time, error) {
	if s == nil {
		return nil, nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil, nil
	}
	// coba beberapa format umum (RFC3339 paling utama)
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02 15:04:05",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, v); err == nil {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("unsupported time format: %q", v)
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "duplicate key") ||
		strings.Contains(s, "unique constraint") ||
		strings.Contains(s, "23505")
}
