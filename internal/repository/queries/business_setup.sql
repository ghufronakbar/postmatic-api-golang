-- name: SetupBusinessRootFirstTime :one
WITH r AS (
  INSERT INTO business_roots DEFAULT VALUES
  RETURNING id
),
k AS (
  INSERT INTO business_knowledges (
    name, primary_logo_url, category, description,
    unique_selling_point, website_url, vision_mission, location, color_tone,
    business_root_id
  )
  SELECT
    $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    r.id
  FROM r
  RETURNING id, business_root_id
),
p AS (
  INSERT INTO business_products (
    name, category, description,
    currency, price, image_urls,
    business_root_id
  )
  SELECT
    $11, $12, $13,
    $14, $15, $16::varchar(255)[],
    r.id
  FROM r
  RETURNING id, business_root_id
),
ro AS (
  INSERT INTO business_roles (
    target_audience, tone, audience_persona, hashtags, call_to_action, goals,
    business_root_id
  )
  SELECT
    $17, $18, $19, $20::varchar(255)[], $21, $22,
    r.id
  FROM r
  RETURNING id, business_root_id
),
m AS (
  INSERT INTO business_members (
    status, role, answered_at,
    business_root_id, profile_id
  )
  SELECT
    'accepted'::business_member_status,
    'owner'::business_member_role,
    CURRENT_TIMESTAMP,
    r.id,
    $1
  FROM r
  RETURNING id, business_root_id, profile_id
)
SELECT
  r.id  AS business_root_id,
  k.id  AS business_knowledge_id,
  p.id  AS business_product_id,
  ro.id AS business_role_id,
  m.id  AS business_member_id
FROM r, k, p, ro, m;
