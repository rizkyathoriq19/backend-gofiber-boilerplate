-- Drop the view
DROP VIEW IF EXISTS role_permissions_with_inheritance;

-- Drop the functions
DROP FUNCTION IF EXISTS get_role_descendants(UUID);
DROP FUNCTION IF EXISTS get_role_ancestors(UUID);

-- Drop the index
DROP INDEX IF EXISTS idx_roles_parent_role_id;

-- Remove the columns
ALTER TABLE roles DROP COLUMN IF EXISTS level;
ALTER TABLE roles DROP COLUMN IF EXISTS parent_role_id;
