-- +goose Up
-- +goose StatementBegin
CREATE TABLE app_rss_category (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	name VARCHAR(255) NOT NULL,
	deleted_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trigger_app_rss_category_updated_at
BEFORE UPDATE ON app_rss_category
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE app_rss (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	title VARCHAR(255) NOT NULL,
	url VARCHAR(255) NOT NULL,
	publisher VARCHAR(255) NOT NULL,
	master_rss_category_id UUID NOT NULL,
	FOREIGN KEY (master_rss_category_id) REFERENCES app_rss_category(id),
	deleted_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trigger_app_rss_updated_at
BEFORE UPDATE ON app_rss
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trigger_app_rss_updated_at ON app_rss;
DROP TABLE IF EXISTS app_rss;

DROP TRIGGER IF EXISTS trigger_app_rss_category_updated_at ON app_rss_category;
DROP TABLE IF EXISTS app_rss_category;
-- +goose StatementEnd
