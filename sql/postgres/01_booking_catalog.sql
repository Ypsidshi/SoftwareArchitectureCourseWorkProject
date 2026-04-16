CREATE EXTENSION IF NOT EXISTS btree_gist;

ALTER TABLE auth.users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE auth.users
    ADD CONSTRAINT users_role_check
    CHECK (role IN ('admin', 'manager', 'accountant', 'client'));

CREATE TABLE IF NOT EXISTS deal.medical_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS deal.sanatoriums (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    city TEXT NOT NULL,
    address TEXT NOT NULL,
    distance_to_sea_km NUMERIC(8, 2) NOT NULL DEFAULT 0 CHECK (distance_to_sea_km >= 0),
    amenities JSONB NOT NULL DEFAULT '[]'::jsonb,
    image_urls JSONB NOT NULL DEFAULT '[]'::jsonb,
    price_per_night NUMERIC(12, 2) NOT NULL CHECK (price_per_night > 0),
    total_places INT NOT NULL DEFAULT 1 CHECK (total_places > 0),
    latitude NUMERIC(9, 6) NULL,
    longitude NUMERIC(9, 6) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sanatoriums_city ON deal.sanatoriums (LOWER(city));
CREATE INDEX IF NOT EXISTS idx_sanatoriums_price ON deal.sanatoriums (price_per_night);
CREATE INDEX IF NOT EXISTS idx_sanatoriums_distance ON deal.sanatoriums (distance_to_sea_km);

CREATE TABLE IF NOT EXISTS deal.sanatorium_medical_profiles (
    sanatorium_id UUID NOT NULL REFERENCES deal.sanatoriums(id) ON DELETE CASCADE,
    profile_id UUID NOT NULL REFERENCES deal.medical_profiles(id) ON DELETE CASCADE,
    PRIMARY KEY (sanatorium_id, profile_id)
);

CREATE TABLE IF NOT EXISTS deal.bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL,
    sanatorium_id UUID NOT NULL REFERENCES deal.sanatoriums(id) ON DELETE RESTRICT,
    check_in DATE NOT NULL,
    check_out DATE NOT NULL,
    guests INT NOT NULL CHECK (guests > 0),
    status TEXT NOT NULL DEFAULT 'confirmed' CHECK (status IN ('created', 'confirmed', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    cancelled_at TIMESTAMPTZ NULL,
    CONSTRAINT ck_booking_dates CHECK (check_in < check_out)
);

CREATE INDEX IF NOT EXISTS idx_bookings_client ON deal.bookings (client_id);
CREATE INDEX IF NOT EXISTS idx_bookings_sanatorium ON deal.bookings (sanatorium_id);
CREATE INDEX IF NOT EXISTS idx_bookings_dates ON deal.bookings (check_in, check_out);
CREATE INDEX IF NOT EXISTS idx_bookings_status ON deal.bookings (status);

-- Optional sample data for quick manual testing in local environment.
INSERT INTO deal.medical_profiles (name, description)
VALUES
    ('cardiology', 'Heart and cardiovascular treatment profile'),
    ('pulmonology', 'Lung and respiratory treatment profile'),
    ('musculoskeletal', 'Musculoskeletal rehabilitation profile')
ON CONFLICT (name) DO NOTHING;

INSERT INTO deal.sanatoriums (
    name,
    description,
    city,
    address,
    distance_to_sea_km,
    amenities,
    image_urls,
    price_per_night,
    total_places,
    latitude,
    longitude
)
VALUES
    (
        'Sea Breeze Health Resort',
        'Modern sanatorium for family and therapeutic rest.',
        'Sochi',
        'Kurortny Ave, 10',
        0.5,
        '["spa","pool","wifi","medical_center"]'::jsonb,
        '["https://example.com/images/sea-breeze-1.jpg","https://example.com/images/sea-breeze-2.jpg"]'::jsonb,
        6200,
        1,
        43.585472,
        39.723098
    ),
    (
        'Mountain Valley Sanatorium',
        'Quiet mountain complex for respiratory and rehabilitation programs.',
        'Kislovodsk',
        'Park Lane, 7',
        999.0,
        '["mineral_water","wifi","gym"]'::jsonb,
        '["https://example.com/images/mountain-valley-1.jpg"]'::jsonb,
        4800,
        1,
        43.905225,
        42.716964
    )
ON CONFLICT DO NOTHING;

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id
FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('cardiology', 'musculoskeletal')
WHERE s.name = 'Sea Breeze Health Resort'
ON CONFLICT DO NOTHING;

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id
FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('pulmonology')
WHERE s.name = 'Mountain Valley Sanatorium'
ON CONFLICT DO NOTHING;
