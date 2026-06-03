-- Expand medical profile catalog and reassign demo sanatorium links.
-- docker cp sql/postgres/09_expand_medical_profiles.sql coursework-postgres:/tmp/09.sql
-- docker exec coursework-postgres psql -U postgres -d sanatorium -f /tmp/09.sql

INSERT INTO deal.medical_profiles (name, description)
VALUES
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

DELETE FROM deal.sanatorium_medical_profiles smp
WHERE smp.sanatorium_id IN (
    SELECT id FROM deal.sanatoriums
    WHERE name IN (
        'Sea Breeze Health Resort',
        'Mountain Valley Sanatorium',
        'Санаторий Лазурный Берег',
        'Санаторий Рассвет',
        'Санаторий Сосновый Бор',
        'Санаторий Морская Звезда',
        'Санаторий Горный Источник',
        'Санаторий Южные Террасы'
    )
);

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('cardiology', 'pulmonology', 'rehabilitation')
WHERE s.name = 'Sea Breeze Health Resort' ON CONFLICT DO NOTHING;

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('pulmonology', 'balneology', 'gastroenterology')
WHERE s.name = 'Mountain Valley Sanatorium' ON CONFLICT DO NOTHING;

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('dermatology', 'general_therapy', 'cardiology')
WHERE s.name = 'Санаторий Лазурный Берег' ON CONFLICT DO NOTHING;

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('neurology', 'pediatrics', 'rehabilitation')
WHERE s.name = 'Санаторий Рассвет' ON CONFLICT DO NOTHING;

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('musculoskeletal', 'balneology', 'cardiology')
WHERE s.name = 'Санаторий Сосновый Бор' ON CONFLICT DO NOTHING;

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('pulmonology', 'pediatrics', 'dermatology')
WHERE s.name = 'Санаторий Морская Звезда' ON CONFLICT DO NOTHING;

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('gastroenterology', 'balneology', 'endocrinology')
WHERE s.name = 'Санаторий Горный Источник' ON CONFLICT DO NOTHING;

INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
SELECT s.id, p.id FROM deal.sanatoriums s
JOIN deal.medical_profiles p ON p.name IN ('musculoskeletal', 'neurology', 'urology')
WHERE s.name = 'Санаторий Южные Террасы' ON CONFLICT DO NOTHING;
