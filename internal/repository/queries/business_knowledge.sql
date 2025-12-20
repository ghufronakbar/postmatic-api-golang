-- name: CreateBusinessKnowledge :one
INSERT INTO business_knowledges (
    name,
    primary_logo_url,
    category,
    description,
    unique_selling_point,
    website_url,
    vision_mission,
    location,
    color_tone,
    business_root_id
)
VALUES (
    sqlc.arg(name),
    sqlc.arg(primary_logo_url),
    sqlc.arg(category),
    sqlc.arg(description),
    sqlc.arg(unique_selling_point),
    sqlc.arg(website_url),
    sqlc.arg(vision_mission),
    sqlc.arg(location),
    sqlc.arg(color_tone),
    sqlc.arg(business_root_id)
)
RETURNING *;

-- name: GetBusinessKnowledgeByBusinessRootID :one
SELECT root.id AS business_root_id, kn.name, kn.primary_logo_url, kn.category, kn.description, kn.color_tone, root.created_at, root.updated_at, kn.unique_selling_point, kn.website_url, kn.vision_mission, kn.location
FROM business_knowledges kn
JOIN business_roots root ON kn.business_root_id = root.id
WHERE root.id = sqlc.arg(business_root_id) AND root.deleted_at IS NULL AND kn.deleted_at IS NULL;


-- name: SoftDeleteBusinessKnowledgeByBusinessRootID :one
UPDATE business_knowledges
SET deleted_at = NOW()
WHERE business_root_id = sqlc.arg(business_root_id)
RETURNING id;


-- name: UpsertBusinessKnowledgeByBusinessRootID :one
INSERT INTO business_knowledges (
  name,
  primary_logo_url,
  category,
  description,
  unique_selling_point,
  website_url,
  vision_mission,
  location,
  color_tone,
  business_root_id
) VALUES (
  sqlc.arg(name),
  sqlc.narg(primary_logo_url),
  sqlc.arg(category),
  sqlc.narg(description),
  sqlc.narg(unique_selling_point),
  sqlc.narg(website_url),
  sqlc.narg(vision_mission),
  sqlc.narg(location),
  sqlc.narg(color_tone),
  sqlc.arg(business_root_id)
)
ON CONFLICT (business_root_id) DO UPDATE SET
  name                = EXCLUDED.name,
  primary_logo_url     = EXCLUDED.primary_logo_url,
  category            = EXCLUDED.category,
  description         = EXCLUDED.description,
  unique_selling_point = EXCLUDED.unique_selling_point,
  website_url         = EXCLUDED.website_url,
  vision_mission      = EXCLUDED.vision_mission,
  location            = EXCLUDED.location,
  color_tone          = EXCLUDED.color_tone,
  deleted_at          = NULL
RETURNING *;
