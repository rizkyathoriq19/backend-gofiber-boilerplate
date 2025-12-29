-- Seed default users
-- Password: password123 (bcrypt hash with cost 10)
-- Hash generated from: password123

INSERT INTO users (id, name, email, password, role) VALUES
    -- Super Admin user
    ('00000000-0000-0000-0000-000000000001', 
     'Super Admin', 
     'superadmin@example.com', 
     '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 
     'admin'),
    -- Admin user
    ('00000000-0000-0000-0000-000000000002', 
     'Admin User', 
     'admin@example.com', 
     '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 
     'admin'),
    -- Regular user
    ('00000000-0000-0000-0000-000000000003', 
     'Test User', 
     'user@example.com', 
     '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 
     'user')
ON CONFLICT (email) DO NOTHING;

-- Assign super_admin role to superadmin user
INSERT INTO user_roles (user_id, role_id)
SELECT '00000000-0000-0000-0000-000000000001', r.id 
FROM roles r WHERE r.name = 'super_admin'
ON CONFLICT DO NOTHING;

-- Assign admin role to admin user
INSERT INTO user_roles (user_id, role_id)
SELECT '00000000-0000-0000-0000-000000000002', r.id 
FROM roles r WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;

-- Assign user role to regular user
INSERT INTO user_roles (user_id, role_id)
SELECT '00000000-0000-0000-0000-000000000003', r.id 
FROM roles r WHERE r.name = 'user'
ON CONFLICT DO NOTHING;
