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
SELECT root.id AS business_root_id, kn.name, kn.primary_logo_url, kn.category, kn.description, kn.color_tone, root.created_at, root.updated_at
FROM business_knowledges kn
JOIN business_roots root ON kn.business_root_id = root.id
WHERE root.id = sqlc.arg(business_root_id) AND root.deleted_at IS NULL AND kn.deleted_at IS NULL;
