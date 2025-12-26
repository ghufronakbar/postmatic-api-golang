-- +goose Up
-- +goose StatementBegin

-- ex: announcement, entertainment, marketing, etc
CREATE TABLE IF NOT EXISTS app_creator_image_type_categories (
	id BIGSERIAL PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
	created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trigger_app_creator_image_type_categories_updated_at
BEFORE UPDATE ON app_creator_image_type_categories
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- ex: fashion & lifestyle, makanan & minuman, etc
CREATE TABLE IF NOT EXISTS app_creator_image_product_categories (
	id BIGSERIAL PRIMARY KEY,
    indonesian_name VARCHAR(255) NOT NULL,
    english_name VARCHAR(255) NOT NULL,
	created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trigger_app_creator_image_product_categories_updated_at
BEFORE UPDATE ON app_creator_image_product_categories
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();


-- base table (tanpa FK category)
CREATE TABLE IF NOT EXISTS creator_images (
	id BIGSERIAL PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
	image_url TEXT NOT NULL,
    is_published BOOLEAN NOT NULL DEFAULT FALSE,
    is_banned BOOLEAN NOT NULL DEFAULT FALSE,
    banned_reason TEXT,
    price BIGINT NOT NULL,

	-- creator profile (if null -> template from app/postmatic)
	profile_id UUID,
    CONSTRAINT fk_creator_images_profile
		FOREIGN KEY (profile_id) REFERENCES profiles(id),

	created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
	deleted_at TIMESTAMPTZ
);

CREATE TRIGGER trigger_creator_images_updated_at
BEFORE UPDATE ON creator_images
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();


-- pivot: creator_images <-> product categories
CREATE TABLE IF NOT EXISTS creator_image_product_categories (
	creator_image_id BIGINT NOT NULL,
	product_category_id BIGINT NOT NULL,
	created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT pk_creator_image_product_categories
		PRIMARY KEY (creator_image_id, product_category_id),

	CONSTRAINT fk_cipc_creator_image
		FOREIGN KEY (creator_image_id) REFERENCES creator_images(id) ON DELETE CASCADE,

	-- biasanya NO ACTION/RESTRICT biar category tidak bisa dihapus kalau masih dipakai
	CONSTRAINT fk_cipc_product_category
		FOREIGN KEY (product_category_id) REFERENCES app_creator_image_product_categories(id)
);

CREATE INDEX IF NOT EXISTS idx_cipc_creator_image_id
	ON creator_image_product_categories(creator_image_id);

CREATE INDEX IF NOT EXISTS idx_cipc_product_category_id
	ON creator_image_product_categories(product_category_id);


-- pivot: creator_images <-> type categories
CREATE TABLE IF NOT EXISTS creator_image_type_categories (
	creator_image_id BIGINT NOT NULL,
	type_category_id BIGINT NOT NULL,
	created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT pk_creator_image_type_categories
		PRIMARY KEY (creator_image_id, type_category_id),

	CONSTRAINT fk_citc_creator_image
		FOREIGN KEY (creator_image_id) REFERENCES creator_images(id) ON DELETE CASCADE,

	CONSTRAINT fk_citc_type_category
		FOREIGN KEY (type_category_id) REFERENCES app_creator_image_type_categories(id)
);

CREATE INDEX IF NOT EXISTS idx_citc_creator_image_id
	ON creator_image_type_categories(creator_image_id);

CREATE INDEX IF NOT EXISTS idx_citc_type_category_id
	ON creator_image_type_categories(type_category_id);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

-- drop triggers explicitly (meskipun drop table juga akan otomatis buang trigger)
DROP TRIGGER IF EXISTS trigger_creator_images_updated_at ON creator_images;
DROP TRIGGER IF EXISTS trigger_app_creator_image_product_categories_updated_at ON app_creator_image_product_categories;
DROP TRIGGER IF EXISTS trigger_app_creator_image_type_categories_updated_at ON app_creator_image_type_categories;

-- drop pivot tables first (depend on creator_images & category tables)
DROP TABLE IF EXISTS creator_image_type_categories;
DROP TABLE IF EXISTS creator_image_product_categories;

-- drop base + master tables
DROP TABLE IF EXISTS creator_images;
DROP TABLE IF EXISTS app_creator_image_product_categories;
DROP TABLE IF EXISTS app_creator_image_type_categories;

-- DO NOT drop set_updated_at() here unless this migration created it
-- DROP FUNCTION IF EXISTS set_updated_at();

-- +goose StatementEnd
