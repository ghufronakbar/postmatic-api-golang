-- +goose Up
-- +goose StatementBegin

-- payment_type nantinya ditambah jika ada product lainnya
CREATE TYPE payment_product_type AS ENUM ('image_token', 'video_token', 'livestream_token');

CREATE TYPE payment_status AS ENUM ('pending', 'success', 'failed', 'canceled', 'refunded', 'expired', 'denied');

CREATE TABLE IF NOT EXISTS payment_histories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- profile that make the payment
    profile_id UUID NOT NULL,
    FOREIGN KEY (profile_id) REFERENCES profiles (id),

    -- business root that make the payment
    business_root_id BIGINT NOT NULL,
    FOREIGN KEY (business_root_id) REFERENCES business_roots (id),

    -- common
    product_amount BIGINT NOT NULL,
    status payment_status NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'IDR',
    -- ex: OVO, Gopay, etc
    payment_method VARCHAR(255) NOT NULL,
    -- ex: bank, ewallet, etc
    payment_method_type VARCHAR(255) NOT NULL,

    -- informative (denormalized)
    -- product (app token product)
    record_product_name VARCHAR(255) NOT NULL,
    record_product_type payment_product_type NOT NULL,
    record_product_price BIGINT NOT NULL,
    record_product_image_url VARCHAR(255) NOT NULL,
    -- tidak perlu foreign key karena tidak perlu diakses (hanya untuk audit log)
    -- ex ambil dari app_token_products 
    reference_product_id UUID NOT NULL,

    -- payment details
    -- item
    -- ex: Rp 10.000
    subtotal_item_amount BIGINT NOT NULL,

    -- discount
    -- ex: Rp 10.000
    discount_amount BIGINT NOT NULL,
    -- jika discount fixed maka discount_percentage akan null
    discount_percentage INT,
    discount_type discount_type NOT NULL,

    -- admin fee
    admin_fee_amount BIGINT NOT NULL,
    admin_fee_percentage INT,
    -- pakai discount type karena sama dengan discount typenya
    admin_fee_type discount_type NOT NULL,

    -- tax
    -- ex: Rp 10.000 (tax pasti dari percentage)
    tax_amount BIGINT NOT NULL,
    tax_percentage INT NOT NULL,

    -- relation into referral_records for tracking referral (1 to 1 not mandatory)
    referral_record_id BIGINT,
    FOREIGN KEY (referral_record_id) REFERENCES referral_records (id),

    -- midtrans
    midtrans_transaction_id VARCHAR(255),
    midtrans_expired_at TIMESTAMPTZ,

    -- payment lifecycle timestamps
    payment_pending_at TIMESTAMPTZ,      -- saat pertama kali checkout
    payment_success_at TIMESTAMPTZ,      -- saat pembayaran berhasil
    payment_failed_at TIMESTAMPTZ,       -- saat pembayaran gagal
    payment_canceled_at TIMESTAMPTZ,     -- saat user cancel
    payment_expired_at TIMESTAMPTZ,      -- saat expired
    payment_refunded_at TIMESTAMPTZ,     -- saat refund

    -- grand total (setelah semua kalkulasi)
    total_amount BIGINT NOT NULL,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);
CREATE TRIGGER trigger_payment_histories_updated_at
BEFORE UPDATE ON payment_histories
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trigger_payment_histories_updated_at ON payment_histories;
DROP TABLE IF EXISTS payment_histories;
-- +goose StatementEnd
