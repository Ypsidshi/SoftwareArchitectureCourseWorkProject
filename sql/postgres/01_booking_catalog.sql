CREATE EXTENSION IF NOT EXISTS btree_gist;

UPDATE auth.users SET role = 'admin' WHERE role IN ('manager', 'accountant');

ALTER TABLE auth.users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE auth.users
    ADD CONSTRAINT users_role_check
    CHECK (role IN ('admin', 'client'));

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

CREATE UNIQUE INDEX IF NOT EXISTS uq_deal_sanatoriums_name_city ON deal.sanatoriums (name, city);

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
    ('musculoskeletal', 'Musculoskeletal rehabilitation profile'),
    ('neurology', 'Neurology and nervous system rehabilitation'),
    ('gastroenterology', 'Digestive system treatment programs'),
    ('endocrinology', 'Endocrine and metabolic disorders'),
    ('dermatology', 'Skin and allergic conditions'),
    ('urology', 'Urological health programs'),
    ('pediatrics', 'Family and children wellness programs'),
    ('balneology', 'Mineral water and spa therapy'),
    ('rehabilitation', 'General medical rehabilitation'),
    ('general_therapy', 'General therapeutic and preventive care')
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
        '["/images/sea-breeze-1.jpg","/images/sea-breeze-2.jpg","/images/sea-breeze-3.jpg","/images/sea-breeze-4.jpg"]'::jsonb,
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
        '["/images/mountain-valley-1.jpg","/images/mountain-valley-2.jpg","/images/mountain-valley-3.jpg","/images/mountain-valley-4.jpg"]'::jsonb,
        4800,
        1,
        43.905225,
        42.716964
    ),
    (
        'Санаторий Лазурный Берег',
        'Современный санаторий на побережье для отдыха и оздоровления.',
        'Геленджик',
        'Набережная, 21',
        0.3,
        '["spa","pool","wifi","medical_center"]'::jsonb,
        '["/images/sanatorium-lazurny-1.jpg","/images/sanatorium-lazurny-2.jpg","/images/sanatorium-lazurny-3.jpg"]'::jsonb,
        7300,
        2,
        44.562200,
        38.080000
    ),
    (
        'Санаторий Рассвет',
        'Просторный приморский санаторий с программами оздоровления и реабилитации.',
        'Светлогорск',
        'Морская, 8',
        0.6,
        '["spa","pool","wifi","medical_center"]'::jsonb,
        '["/images/sanatorium-rassvet-1.jpg","/images/sanatorium-rassvet-2.jpg","/images/sanatorium-rassvet-3.jpg"]'::jsonb,
        7100,
        2,
        54.943900,
        20.151500
    ),
    (
        'Санаторий Сосновый Бор',
        'Санаторий в лесной зоне с программами реабилитации и кардиопрофилем.',
        'Белокуриха',
        'Лесная, 5',
        700.0,
        '["mineral_water","wifi","gym"]'::jsonb,
        '["/images/pine-forest-1.jpg","/images/pine-forest-2.jpg","/images/pine-forest-3.jpg","/images/pine-forest-4.jpg"]'::jsonb,
        5400,
        2,
        51.996300,
        84.985000
    ),
    (
        'Санаторий Морская Звезда',
        'Курортный комплекс для семейного отдыха и профилактики дыхательной системы.',
        'Анапа',
        'Пионерский пр-т, 46',
        0.7,
        '["pool","wifi","medical_center"]'::jsonb,
        '["/images/sea-star-1.jpg","/images/sea-star-2.jpg","/images/sea-star-3.jpg","/images/sea-star-4.jpg"]'::jsonb,
        6800,
        3,
        44.922100,
        37.316600
    ),
    (
        'Санаторий Горный Источник',
        'Горный санаторий с термальными процедурами и спокойной атмосферой.',
        'Ессентуки',
        'Курортная, 14',
        420.0,
        '["spa","mineral_water","wifi"]'::jsonb,
        '["/images/mountain-spring-1.jpg","/images/mountain-spring-2.jpg","/images/mountain-spring-3.jpg","/images/mountain-spring-4.jpg"]'::jsonb,
        5900,
        2,
        44.044400,
        42.858900
    ),
    (
        'Санаторий Южные Террасы',
        'Южный пансионат санаторного типа с программами опорно-двигательной терапии.',
        'Ялта',
        'Приморская, 32',
        1.2,
        '["spa","pool","wifi","gym"]'::jsonb,
        '["/images/south-terraces-1.jpg","/images/south-terraces-2.jpg","/images/south-terraces-3.jpg","/images/south-terraces-4.jpg"]'::jsonb,
        7600,
        2,
        44.495200,
        34.166300
    )
ON CONFLICT DO NOTHING;

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id
FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('cardiology', 'pulmonology', 'rehabilitation')
WHERE s.name = 'Sea Breeze Health Resort'
ON CONFLICT DO NOTHING;

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id
FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('pulmonology', 'balneology', 'gastroenterology')
WHERE s.name = 'Mountain Valley Sanatorium'
ON CONFLICT DO NOTHING;

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id
FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('dermatology', 'general_therapy', 'cardiology')
WHERE s.name = 'Санаторий Лазурный Берег'
ON CONFLICT DO NOTHING;

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id
FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('neurology', 'pediatrics', 'rehabilitation')
WHERE s.name = 'Санаторий Рассвет'
ON CONFLICT DO NOTHING;

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id
FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('musculoskeletal', 'balneology', 'cardiology')
WHERE s.name = 'Санаторий Сосновый Бор'
ON CONFLICT DO NOTHING;

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id
FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('pulmonology', 'pediatrics', 'dermatology')
WHERE s.name = 'Санаторий Морская Звезда'
ON CONFLICT DO NOTHING;

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id
FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('gastroenterology', 'balneology', 'endocrinology')
WHERE s.name = 'Санаторий Горный Источник'
ON CONFLICT DO NOTHING;

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id
FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('musculoskeletal', 'neurology', 'urology')
WHERE s.name = 'Санаторий Южные Террасы'
ON CONFLICT DO NOTHING;
