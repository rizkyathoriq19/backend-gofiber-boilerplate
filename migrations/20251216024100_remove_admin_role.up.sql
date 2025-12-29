-- Migration to remove admin role and update default roles
-- Run this after the initial RBAC tables are created

-- Remove admin role from role_permissions first
DELETE FROM role_permissions 
WHERE role_id = (SELECT id FROM roles WHERE name = 'admin');

-- Remove admin role from user_roles
DELETE FROM user_roles 
WHERE role_id = (SELECT id FROM roles WHERE name = 'admin');

-- Delete admin role
DELETE FROM roles WHERE name = 'admin';

-- Update existing admin users to super_admin role
UPDATE user_roles 
SET role_id = (SELECT id FROM roles WHERE name = 'super_admin')
WHERE role_id NOT IN (SELECT id FROM roles);
