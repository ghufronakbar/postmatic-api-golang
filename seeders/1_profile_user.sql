-- pastikan extension
CREATE EXTENSION IF NOT EXISTS pgcrypto;

WITH p AS (
  INSERT INTO profiles (name, email, image_url, country_code, phone, description)
  VALUES ('Lans The Prodigy', 'lanstheprodigy@gmail.com', NULL, '62', '8123456789', NULL)
  RETURNING id
)
INSERT INTO users (profile_id, password, provider, verified_at)
SELECT p.id, crypt('12345678', gen_salt('bf')), 'credential', now()
FROM p
RETURNING *;
