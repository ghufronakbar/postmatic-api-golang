-- +goose Up
-- +goose StatementBegin

CREATE TYPE auth_provider AS ENUM ('google', 'credential');
CREATE TABLE IF NOT EXISTS profiles (
    -- Tambahkan DEFAULT gen_random_uuid()
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), 
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,

    -- Gunakan sql.NullString di Go nantinya
    image_url VARCHAR(255),
    country_code VARCHAR(4),
    phone VARCHAR(20),
    description TEXT,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP

);

CREATE TRIGGER trigger_profiles_updated_at
BEFORE UPDATE ON profiles
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    password VARCHAR(255),
    provider auth_provider NOT NULL,
    verified_at TIMESTAMPTZ,

    -- Tambahkan NOT NULL agar di Go tipenya uuid.UUID (bukan NullUUID)
    profile_id UUID NOT NULL, 
    FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP

);

CREATE TRIGGER trigger_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS trigger_users_updated_at ON users;
DROP TRIGGER IF EXISTS trigger_profiles_updated_at ON profiles;

DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS profiles;

DROP TYPE IF EXISTS auth_provider;

-- +goose StatementEnd
