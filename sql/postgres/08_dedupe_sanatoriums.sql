-- Remove duplicate sanatoriums (same name + city), keep the oldest row.
-- Reassign bookings to the kept row, then delete extras. Add unique index last.

WITH ranked AS (
    SELECT
        id,
        name,
        city,
        ROW_NUMBER() OVER (PARTITION BY name, city ORDER BY created_at ASC, id ASC) AS rn
    FROM deal.sanatoriums
),
dupes AS (
    SELECT d.id AS dupe_id, k.id AS keep_id
    FROM ranked d
    JOIN ranked k ON k.name = d.name AND k.city = d.city AND k.rn = 1
    WHERE d.rn > 1
)
UPDATE deal.bookings b
SET sanatorium_id = d.keep_id,
    updated_at = NOW()
FROM dupes d
WHERE b.sanatorium_id = d.dupe_id;

WITH ranked AS (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY name, city ORDER BY created_at ASC, id ASC) AS rn
    FROM deal.sanatoriums
)
DELETE FROM deal.sanatoriums s
WHERE s.id IN (SELECT id FROM ranked WHERE rn > 1);

CREATE UNIQUE INDEX IF NOT EXISTS uq_deal_sanatoriums_name_city ON deal.sanatoriums (name, city);
