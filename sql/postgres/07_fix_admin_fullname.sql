-- Fix demo admin display name (UTF-8). Apply via docker cp if full_name shows as ??? in UI.
UPDATE auth.users
SET full_name = 'Администратор',
    updated_at = NOW()
WHERE email = 'admin@sanatorium.local';
