-- Apply to existing DB: refresh image URLs only (no duplicate sanatorium rows).
-- "Санаторий Рассвет" is seeded in 01_booking_catalog.sql.

UPDATE deal.sanatoriums
SET image_urls = '["/images/sanatorium-lazurny-1.jpg","/images/sanatorium-lazurny-2.jpg","/images/sanatorium-lazurny-3.jpg"]'::jsonb,
    updated_at = NOW()
WHERE name = 'Санаторий Лазурный Берег';

UPDATE deal.sanatoriums
SET image_urls = '["/images/mountain-valley-1.jpg","/images/mountain-valley-2.jpg","/images/mountain-valley-3.jpg","/images/mountain-valley-4.jpg"]'::jsonb,
    updated_at = NOW()
WHERE name = 'Mountain Valley Sanatorium';

UPDATE deal.sanatoriums
SET image_urls = '["/images/sanatorium-rassvet-1.jpg","/images/sanatorium-rassvet-2.jpg","/images/sanatorium-rassvet-3.jpg"]'::jsonb,
    updated_at = NOW()
WHERE name = 'Санаторий Рассвет';
