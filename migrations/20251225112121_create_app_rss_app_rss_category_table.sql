-- +goose Up
-- +goose StatementBegin
CREATE TABLE app_rss_categories (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	name VARCHAR(255) NOT NULL,
	deleted_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trigger_app_rss_categories_updated_at
BEFORE UPDATE ON app_rss_categories
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE app_rss_feeds (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	title VARCHAR(255) NOT NULL,
	url VARCHAR(255) NOT NULL,
	publisher VARCHAR(255) NOT NULL,
	app_rss_category_id UUID NOT NULL,
	FOREIGN KEY (app_rss_category_id) REFERENCES app_rss_categories(id),
	deleted_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trigger_app_rss_feeds_updated_at
BEFORE UPDATE ON app_rss_feeds
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trigger_app_rss_feeds_updated_at ON app_rss_feeds;
DROP TABLE IF EXISTS app_rss_feeds;

DROP TRIGGER IF EXISTS trigger_app_rss_categories_updated_at ON app_rss_categories;
DROP TABLE IF EXISTS app_rss_categories;
-- +goose StatementEnd
