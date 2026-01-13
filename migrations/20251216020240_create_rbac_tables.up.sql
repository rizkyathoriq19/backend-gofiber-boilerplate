-- Roles table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Permissions table  
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User-Roles junction table
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id)
);

-- Role-Permissions junction table
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, permission_id)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions(permission_id);
CREATE INDEX IF NOT EXISTS idx_permissions_resource ON permissions(resource);

-- Seed default roles
INSERT INTO roles (name, description) VALUES
    ('super_admin', 'Full system access with all permissions'),
    ('head_nurse', 'Head nurse - can view all rooms and manage nursing staff')
ON CONFLICT (name) DO NOTHING;

-- Seed default permissions
INSERT INTO permissions (name, resource, action, description) VALUES
    -- User permissions
    ('users:read', 'users', 'read', 'View user list and details'),
    ('users:write', 'users', 'write', 'Create and update users'),
    ('users:delete', 'users', 'delete', 'Delete users'),
    -- Role permissions
    ('roles:read', 'roles', 'read', 'View roles'),
    ('roles:write', 'roles', 'write', 'Create and update roles'),
    ('roles:delete', 'roles', 'delete', 'Delete roles'),
    -- Permission management
    ('permissions:read', 'permissions', 'read', 'View permissions'),
    ('permissions:assign', 'permissions', 'assign', 'Assign permissions to roles'),
    -- Profile permissions
    ('profile:read', 'profile', 'read', 'View own profile'),
    ('profile:write', 'profile', 'write', 'Update own profile')
ON CONFLICT (name) DO NOTHING;

-- Assign all permissions to super_admin
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'super_admin'
ON CONFLICT DO NOTHING;

-- Assign permissions to head_nurse
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'head_nurse' AND p.name IN ('users:read', 'roles:read', 'profile:read', 'profile:write')
ON CONFLICT DO NOTHING;
