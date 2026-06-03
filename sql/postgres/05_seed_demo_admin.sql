-- Demo administrator (idempotent). Password: Admin1234!
-- Regenerate hash: go run in auth-service with golang.org/x/crypto/bcrypt

INSERT INTO auth.users (email, full_name, role, password_hash, is_active)
VALUES (
    'admin@sanatorium.local',
    'Администратор',
    'admin',
    '$2a$10$nm8ZD4T9jLY91Lk7As2SlurKrM29GxBmFuIHLYcw2eoTaa1ke4tRy',
    TRUE
)
ON CONFLICT (email) DO UPDATE
SET role = EXCLUDED.role,
    full_name = EXCLUDED.full_name,
    password_hash = EXCLUDED.password_hash,
    is_active = TRUE,
    updated_at = NOW();
