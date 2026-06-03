-- Replace demo image URLs with real sanatorium photos (public/images/*.jpg).
-- Apply to existing DB:
--   docker cp sql/postgres/06_real_sanatorium_images.sql coursework-postgres:/tmp/06.sql
--   docker exec coursework-postgres psql -U postgres -d sanatorium -f /tmp/06.sql

UPDATE deal.sanatoriums
SET image_urls = '["/images/sea-breeze-1.jpg","/images/sea-breeze-2.jpg","/images/sea-breeze-3.jpg","/images/sea-breeze-4.jpg"]'::jsonb,
    updated_at = NOW()
WHERE name = 'Sea Breeze Health Resort';

UPDATE deal.sanatoriums
SET image_urls = '["/images/mountain-valley-1.jpg","/images/mountain-valley-2.jpg","/images/mountain-valley-3.jpg","/images/mountain-valley-4.jpg"]'::jsonb,
    updated_at = NOW()
WHERE name = 'Mountain Valley Sanatorium';

UPDATE deal.sanatoriums SET image_urls = '["/images/sanatorium-lazurny-1.jpg","/images/sanatorium-lazurny-2.jpg","/images/sanatorium-lazurny-3.jpg"]'::jsonb, updated_at = NOW() WHERE city IN ('Gelendzhik', 'Геленджик');
UPDATE deal.sanatoriums SET image_urls = '["/images/sanatorium-rassvet-1.jpg","/images/sanatorium-rassvet-2.jpg","/images/sanatorium-rassvet-3.jpg"]'::jsonb, updated_at = NOW() WHERE city IN ('Svetlogorsk', 'Светлогорск');
UPDATE deal.sanatoriums SET image_urls = '["/images/pine-forest-1.jpg","/images/pine-forest-2.jpg","/images/pine-forest-3.jpg","/images/pine-forest-4.jpg"]'::jsonb, updated_at = NOW() WHERE city IN ('Belokurikha', 'Белокуриха');
UPDATE deal.sanatoriums SET image_urls = '["/images/sea-star-1.jpg","/images/sea-star-2.jpg","/images/sea-star-3.jpg","/images/sea-star-4.jpg"]'::jsonb, updated_at = NOW() WHERE city IN ('Anapa', 'Анапа');
UPDATE deal.sanatoriums SET image_urls = '["/images/mountain-spring-1.jpg","/images/mountain-spring-2.jpg","/images/mountain-spring-3.jpg","/images/mountain-spring-4.jpg"]'::jsonb, updated_at = NOW() WHERE city IN ('Yessentuki', 'Ессентуки');
UPDATE deal.sanatoriums SET image_urls = '["/images/south-terraces-1.jpg","/images/south-terraces-2.jpg","/images/south-terraces-3.jpg","/images/south-terraces-4.jpg"]'::jsonb, updated_at = NOW() WHERE city IN ('Yalta', 'Ялта');
