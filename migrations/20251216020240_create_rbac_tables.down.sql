-- Drop junction tables first (due to foreign key constraints)
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS user_roles;

-- Drop main tables
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
