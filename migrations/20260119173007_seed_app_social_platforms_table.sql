-- +goose Up
-- +goose StatementBegin

INSERT INTO app_social_platforms (
    platform_code, 
    logo, 
    name, 
    hint, 
    is_active
) VALUES 
(
    'instagram_business',
    'https://res.cloudinary.com/dudo4q4je/image/upload/v1753541219/instagram-logo_1199-122_huif2f.avif',
    'INSTAGRAM BUSINESS',
    'Link your INSTAGRAM BUSINESS account to your business',
    true
),
(
    'whatsapp_business',
    'https://res.cloudinary.com/dudo4q4je/image/upload/v1753541220/X_logo_isf5az.jpg',
    'WHATSAPP BUSINESS',
    'Link your WHATSAPP BUSINESS account to your business',
    false
),
(
    'tiktok',
    'https://res.cloudinary.com/dudo4q4je/image/upload/v1753541220/tiktok-social-media-app-icon-tiktok-social-media-app-icon-square-shape-vector-illustration-269930887_t8wbfh.webp',
    'TIKTOK',
    'Link your TIKTOK account to your business',
    false
),
(
    'youtube',
    'https://res.cloudinary.com/dudo4q4je/image/upload/v1753541219/images_mcteoy.jpg',
    'YOUTUBE',
    'Link your YOUTUBE account to your business',
    false
),
(
    'pinterest',
    'https://res.cloudinary.com/dudo4q4je/image/upload/v1753541219/2496099_df03il.png',
    'PINTEREST',
    'Link your PINTEREST account to your business',
    false
),
(
    'linked_in',
    'https://res.cloudinary.com/dudo4q4je/image/upload/v1753541220/a347cd89-1662-4421-be90-58e5e8004eae_l0ogtn.svg',
    'LINKEDIN',
    'Link your LINKEDIN account to your business',
    true
),
(
    'facebook_page',
    'https://res.cloudinary.com/dudo4q4je/image/upload/v1753541219/Untitled_design_eu2lim.png',
    'FACEBOOK PAGE',
    'Link your FACEBOOK PAGE account to your business',
    true
)
ON CONFLICT (platform_code) WHERE deleted_at IS NULL 
DO UPDATE SET
    logo = EXCLUDED.logo,
    name = EXCLUDED.name,
    hint = EXCLUDED.hint,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Menghapus data berdasarkan platform_code yang di-seed
DELETE FROM app_social_platforms 
WHERE platform_code IN (
    'instagram_business',
    'whatsapp_business',
    'tiktok',
    'youtube',
    'pinterest',
    'linked_in',
    'facebook_page'
);

-- +goose StatementEnd