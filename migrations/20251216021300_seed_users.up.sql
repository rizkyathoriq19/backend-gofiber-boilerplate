-- Seed default users
-- Password: password123 (bcrypt hash with cost 12)
-- Hash generated from: password123

INSERT INTO users (id, name, email, password, role) VALUES
    -- Super Admin user
    ('00000000-0000-0000-0000-000000000001', 
     'Super Admin', 
     'superadmin@example.com', 
     '$2a$12$Q9IdH6PWbwol9aZYgHslM.VfkVMBqEL3HyceYr9Pa8JYuCpHTIXym', 
     'admin'),
    -- Regular user
    ('00000000-0000-0000-0000-000000000002', 
     'Test User', 
     'user@example.com', 
     '$2a$12$Q9IdH6PWbwol9aZYgHslM.VfkVMBqEL3HyceYr9Pa8JYuCpHTIXym', 
     'user')
ON CONFLICT (email) DO NOTHING;

-- Assign super_admin role to superadmin user
INSERT INTO user_roles (user_id, role_id)
SELECT '00000000-0000-0000-0000-000000000001', r.id 
FROM roles r WHERE r.name = 'super_admin'
ON CONFLICT DO NOTHING;

-- Assign user role to regular user
INSERT INTO user_roles (user_id, role_id)
SELECT '00000000-0000-0000-0000-000000000002', r.id 
FROM roles r WHERE r.name = 'user'
ON CONFLICT DO NOTHING;

