-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM business_products
        WHERE price != FLOOR(price)
    ) THEN
        RAISE EXCEPTION 'price contains decimal values';
    END IF;
END $$;

ALTER TABLE business_products
    ALTER COLUMN price TYPE BIGINT
    USING price::BIGINT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE business_products
    ALTER COLUMN price TYPE DECIMAL(15,2)
    USING price::DECIMAL(15,2);

-- +goose StatementEnd
