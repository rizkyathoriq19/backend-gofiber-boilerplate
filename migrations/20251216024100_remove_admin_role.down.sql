-- Rollback: Re-add admin role

-- Re-insert admin role
INSERT INTO roles (name, description) VALUES
    ('admin', 'Administrative access for user management')
ON CONFLICT (name) DO NOTHING;

-- Re-assign admin permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.name IN ('users:read', 'users:write', 'users:delete', 'roles:read', 'profile:read', 'profile:write')
ON CONFLICT DO NOTHING;
