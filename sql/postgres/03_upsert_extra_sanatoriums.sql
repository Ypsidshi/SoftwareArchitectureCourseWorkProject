-- Idempotent upsert-like script for existing DB without schema reset.
-- Uses name-based checks because there is no UNIQUE(name) constraint.

UPDATE deal.sanatoriums
SET image_urls = '["/images/sea-breeze-1.jpg","/images/sea-breeze-2.jpg","/images/sea-breeze-3.jpg","/images/sea-breeze-4.jpg"]'::jsonb,
    updated_at = NOW()
WHERE name = 'Sea Breeze Health Resort';

UPDATE deal.sanatoriums
SET image_urls = '["/images/mountain-valley-1.jpg","/images/mountain-valley-2.jpg","/images/mountain-valley-3.jpg","/images/mountain-valley-4.jpg"]'::jsonb,
    updated_at = NOW()
WHERE name = 'Mountain Valley Sanatorium';

UPDATE deal.sanatoriums
SET image_urls = '["/images/sanatorium-lazurny-1.jpg","/images/sanatorium-lazurny-2.jpg","/images/sanatorium-lazurny-3.jpg"]'::jsonb,
    updated_at = NOW()
WHERE name = 'Санаторий Лазурный Берег';

INSERT INTO deal.sanatoriums (
    name, description, city, address, distance_to_sea_km, amenities, image_urls,
    price_per_night, total_places, latitude, longitude
)
SELECT
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
WHERE NOT EXISTS (SELECT 1 FROM deal.sanatoriums WHERE name = 'Санаторий Лазурный Берег');

INSERT INTO deal.sanatoriums (
    name, description, city, address, distance_to_sea_km, amenities, image_urls,
    price_per_night, total_places, latitude, longitude
)
SELECT
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
WHERE NOT EXISTS (SELECT 1 FROM deal.sanatoriums WHERE name = 'Санаторий Сосновый Бор');

INSERT INTO deal.sanatoriums (
    name, description, city, address, distance_to_sea_km, amenities, image_urls,
    price_per_night, total_places, latitude, longitude
)
SELECT
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
WHERE NOT EXISTS (SELECT 1 FROM deal.sanatoriums WHERE name = 'Санаторий Морская Звезда');

INSERT INTO deal.sanatoriums (
    name, description, city, address, distance_to_sea_km, amenities, image_urls,
    price_per_night, total_places, latitude, longitude
)
SELECT
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
WHERE NOT EXISTS (SELECT 1 FROM deal.sanatoriums WHERE name = 'Санаторий Горный Источник');

INSERT INTO deal.sanatoriums (
    name, description, city, address, distance_to_sea_km, amenities, image_urls,
    price_per_night, total_places, latitude, longitude
)
SELECT
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
WHERE NOT EXISTS (SELECT 1 FROM deal.sanatoriums WHERE name = 'Санаторий Южные Террасы');

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id
FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('cardiology', 'pulmonology')
WHERE s.name = 'Санаторий Лазурный Берег'
ON CONFLICT DO NOTHING;

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id
FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('cardiology', 'pulmonology')
WHERE s.name = 'Санаторий Рассвет'
ON CONFLICT DO NOTHING;

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id
FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('cardiology', 'musculoskeletal')
WHERE s.name = 'Санаторий Сосновый Бор'
ON CONFLICT DO NOTHING;

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id
FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('pulmonology')
WHERE s.name = 'Санаторий Морская Звезда'
ON CONFLICT DO NOTHING;

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id
FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('musculoskeletal')
WHERE s.name = 'Санаторий Горный Источник'
ON CONFLICT DO NOTHING;

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id
FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('cardiology', 'musculoskeletal')
WHERE s.name = 'Санаторий Южные Террасы'
ON CONFLICT DO NOTHING;
