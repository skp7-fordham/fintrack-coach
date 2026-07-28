CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_user_id_lower_name
ON categories (user_id, LOWER(name));
