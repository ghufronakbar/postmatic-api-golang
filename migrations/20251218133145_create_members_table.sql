-- +goose Up
-- +goose StatementBegin

CREATE TYPE business_member_status AS ENUM ('pending', 'accepted', 'rejected', 'left', 'kicked');
CREATE TYPE business_member_role   AS ENUM ('owner', 'admin', 'member');

CREATE TABLE IF NOT EXISTS business_members (
    id BIGSERIAL PRIMARY KEY,

    status business_member_status NOT NULL DEFAULT 'pending',
    role   business_member_role   NOT NULL DEFAULT 'member',

    answered_at TIMESTAMPTZ,

    business_root_id BIGINT NOT NULL,
    profile_id       UUID NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT business_members_root_profile_unique UNIQUE (business_root_id, profile_id),

    FOREIGN KEY (business_root_id) REFERENCES business_roots(id) ON DELETE CASCADE,
    FOREIGN KEY (profile_id)       REFERENCES profiles(id)       ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_business_members_business_root_id ON business_members(business_root_id);
CREATE INDEX IF NOT EXISTS idx_business_members_profile_id       ON business_members(profile_id);

CREATE TRIGGER trigger_business_members_updated_at
BEFORE UPDATE ON business_members
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS trigger_business_members_updated_at ON business_members;
DROP TABLE IF EXISTS business_members;

DROP TYPE IF EXISTS business_member_role;
DROP TYPE IF EXISTS business_member_status;

-- +goose StatementEnd
