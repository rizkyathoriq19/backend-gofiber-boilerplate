-- Rollback migration: create users table
-- Created at: 2025-08-31 15:28:36

-- Add your DOWN migration SQL here (reverse of UP)
-- Drop triggers
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS users;

-- Drop enum
DROP TYPE IF EXISTS user_role;

-- Note: We keep the uuid-ossp extension as it might be used by other tables