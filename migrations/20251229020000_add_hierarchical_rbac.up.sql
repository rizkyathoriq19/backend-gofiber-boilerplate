-- Add parent_role_id column to roles table for hierarchical RBAC
ALTER TABLE roles ADD COLUMN IF NOT EXISTS parent_role_id UUID REFERENCES roles(id) ON DELETE SET NULL;

-- Create index for parent role lookups
CREATE INDEX IF NOT EXISTS idx_roles_parent_role_id ON roles(parent_role_id);

-- Add level column to track role hierarchy depth
ALTER TABLE roles ADD COLUMN IF NOT EXISTS level INTEGER DEFAULT 0;

-- Function to get all ancestor roles (for permission inheritance)
CREATE OR REPLACE FUNCTION get_role_ancestors(role_id UUID)
RETURNS TABLE(ancestor_id UUID, ancestor_level INTEGER) AS $$
WITH RECURSIVE role_hierarchy AS (
    -- Base case: start with the given role
    SELECT id, parent_role_id, level, 0 as depth
    FROM roles
    WHERE id = role_id
    
    UNION ALL
    
    -- Recursive case: get parent roles
    SELECT r.id, r.parent_role_id, r.level, rh.depth + 1
    FROM roles r
    INNER JOIN role_hierarchy rh ON r.id = rh.parent_role_id
    WHERE rh.depth < 10 -- Prevent infinite loops, max 10 levels
)
SELECT id as ancestor_id, level as ancestor_level
FROM role_hierarchy
WHERE id != role_id; -- Exclude the starting role
$$ LANGUAGE SQL;

-- Function to get all descendant roles
CREATE OR REPLACE FUNCTION get_role_descendants(role_id UUID)
RETURNS TABLE(descendant_id UUID, descendant_level INTEGER) AS $$
WITH RECURSIVE role_hierarchy AS (
    -- Base case: start with the given role
    SELECT id, level, 0 as depth
    FROM roles
    WHERE id = role_id
    
    UNION ALL
    
    -- Recursive case: get child roles
    SELECT r.id, r.level, rh.depth + 1
    FROM roles r
    INNER JOIN role_hierarchy rh ON r.parent_role_id = rh.id
    WHERE rh.depth < 10 -- Prevent infinite loops, max 10 levels
)
SELECT id as descendant_id, level as descendant_level
FROM role_hierarchy
WHERE id != role_id; -- Exclude the starting role
$$ LANGUAGE SQL;

-- View to get all permissions for a role including inherited permissions
CREATE OR REPLACE VIEW role_permissions_with_inheritance AS
SELECT DISTINCT
    r.id as role_id,
    r.name as role_name,
    p.id as permission_id,
    p.name as permission_name,
    p.resource,
    p.action,
    CASE 
        WHEN rp.role_id = r.id THEN false 
        ELSE true 
    END as is_inherited
FROM roles r
LEFT JOIN LATERAL (
    SELECT * FROM get_role_ancestors(r.id)
    UNION ALL
    SELECT r.id, r.level -- Include the role itself
) ancestors(ancestor_id, ancestor_level) ON true
LEFT JOIN role_permissions rp ON rp.role_id = ancestors.ancestor_id
LEFT JOIN permissions p ON p.id = rp.permission_id
WHERE p.id IS NOT NULL;
