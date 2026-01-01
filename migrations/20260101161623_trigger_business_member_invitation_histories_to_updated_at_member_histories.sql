-- +goose Up
-- +goose StatementBegin

CREATE OR REPLACE FUNCTION fn_touch_business_member_updated_at_from_status_history()
RETURNS TRIGGER AS $$
BEGIN
  UPDATE business_members AS bm
  SET updated_at = CURRENT_TIMESTAMP
  WHERE bm.id = NEW.member_id
    AND bm.deleted_at IS NULL;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_touch_business_member_updated_at_from_status_history
ON business_member_status_histories;

CREATE TRIGGER trg_touch_business_member_updated_at_from_status_history
AFTER INSERT ON business_member_status_histories
FOR EACH ROW
EXECUTE FUNCTION fn_touch_business_member_updated_at_from_status_history();

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS trg_touch_business_member_updated_at_from_status_history
ON business_member_status_histories;

DROP FUNCTION IF EXISTS fn_touch_business_member_updated_at_from_status_history();

-- +goose StatementEnd
