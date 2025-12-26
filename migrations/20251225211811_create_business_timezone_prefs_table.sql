-- +goose Up
-- +goose StatementBegin
-- 1 to 1 with business_roots(id)
CREATE TABLE IF NOT EXISTS business_timezone_prefs (
	id SERIAL PRIMARY KEY,
	business_root_id UUID NOT NULL UNIQUE,
    FOREIGN KEY (business_root_id) REFERENCES business_roots (id) ON DELETE CASCADE,
	timezone VARCHAR(255) NOT NULL DEFAULT 'Asia/Jakarta',
	created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE TRIGGER trg_business_timezone_prefs_set_updated_at
BEFORE UPDATE ON business_timezone_prefs
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_touch_root_from_timezone_prefs
AFTER INSERT OR UPDATE OR DELETE ON business_timezone_prefs
FOR EACH ROW
EXECUTE FUNCTION touch_business_root_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_touch_root_from_timezone_prefs ON business_timezone_prefs;
DROP TRIGGER IF EXISTS trg_business_timezone_prefs_set_updated_at ON business_timezone_prefs;
DROP TABLE IF EXISTS business_timezone_prefs;
-- +goose StatementEnd
