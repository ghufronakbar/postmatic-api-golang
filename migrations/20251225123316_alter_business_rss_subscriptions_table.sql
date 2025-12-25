-- +goose Up
-- +goose StatementBegin
-- Relasi: business_rss_subscriptions (*) -> app_rss_feeds (1)

ALTER TABLE business_rss_subscriptions
ADD COLUMN IF NOT EXISTS app_rss_feed_id UUID NOT NULL;

ALTER TABLE business_rss_subscriptions
ADD CONSTRAINT fk_business_rss_subscriptions_app_rss_feed_id
FOREIGN KEY (app_rss_feed_id)
REFERENCES app_rss_feeds(id)
ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_business_rss_subscriptions_app_rss_feed_id
ON business_rss_subscriptions(app_rss_feed_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_business_rss_subscriptions_app_rss_feed_id;

ALTER TABLE business_rss_subscriptions
DROP CONSTRAINT IF EXISTS fk_business_rss_subscriptions_app_rss_feed_id;

ALTER TABLE business_rss_subscriptions
DROP COLUMN IF EXISTS app_rss_feed_id;
-- +goose StatementEnd
