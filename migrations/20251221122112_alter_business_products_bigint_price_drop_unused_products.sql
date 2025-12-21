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

DROP TRIGGER IF EXISTS trigger_products_updated_at ON products;
DROP TABLE IF EXISTS products;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE business_products
    ALTER COLUMN price TYPE DECIMAL(15,2)
    USING price::DECIMAL(15,2);

CREATE TABLE IF NOT EXISTS products ( 
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    price INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trigger_products_updated_at
BEFORE UPDATE ON products
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- +goose StatementEnd
