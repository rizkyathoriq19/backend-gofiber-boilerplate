-- Migration: Add MEDIPROMPT roles and permissions
-- Created at: 2026-01-07

-- =============================================
-- EXPAND USER_ROLE ENUM
-- =============================================
-- Add new role values to user_role enum (for users.role column)
ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'nurse';
ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'doctor';
ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'patient';
ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'manager';

-- =============================================
-- ADD MEDIPROMPT ROLES
-- =============================================
-- Roles: nurse, doctor, patient, manager (admin already exists)

INSERT INTO roles (name, description) VALUES
    ('nurse', 'Nursing staff - can respond to alerts and manage patients in assigned rooms'),
    ('doctor', 'Medical doctor - clinical access with patient management privileges'),
    ('patient', 'Patient user - limited self-service access'),
    ('manager', 'Hospital/ward manager - management access for staff and operations')
ON CONFLICT (name) DO NOTHING;

-- =============================================
-- ADD MEDIPROMPT PERMISSIONS
-- =============================================
INSERT INTO permissions (name, resource, action, description) VALUES
    -- Room permissions
    ('rooms:read', 'rooms', 'read', 'View rooms'),
    ('rooms:write', 'rooms', 'write', 'Create and update rooms'),
    ('rooms:delete', 'rooms', 'delete', 'Delete rooms'),
    
    -- Device permissions
    ('devices:read', 'devices', 'read', 'View devices'),
    ('devices:write', 'devices', 'write', 'Create and update devices'),
    ('devices:delete', 'devices', 'delete', 'Delete devices'),
    ('devices:manage', 'devices', 'manage', 'Manage device API keys'),
    
    -- Staff permissions
    ('staff:read', 'staff', 'read', 'View staff'),
    ('staff:write', 'staff', 'write', 'Create and update staff'),
    ('staff:delete', 'staff', 'delete', 'Delete staff'),
    ('staff:assign_rooms', 'staff', 'assign_rooms', 'Assign staff to rooms'),
    
    -- Patient permissions
    ('patients:read', 'patients', 'read', 'View patients'),
    ('patients:write', 'patients', 'write', 'Create and update patients'),
    ('patients:delete', 'patients', 'delete', 'Delete patients'),
    ('patients:admit', 'patients', 'admit', 'Admit patients'),
    ('patients:discharge', 'patients', 'discharge', 'Discharge patients'),
    
    -- Alert permissions
    ('alerts:read', 'alerts', 'read', 'View alerts'),
    ('alerts:create', 'alerts', 'create', 'Create alerts'),
    ('alerts:acknowledge', 'alerts', 'acknowledge', 'Acknowledge alerts'),
    ('alerts:resolve', 'alerts', 'resolve', 'Resolve alerts'),
    ('alerts:escalate', 'alerts', 'escalate', 'Escalate alerts'),
    
    -- Message permissions
    ('messages:read', 'messages', 'read', 'View messages'),
    ('messages:write', 'messages', 'write', 'Send messages')
ON CONFLICT (name) DO NOTHING;

-- =============================================
-- ASSIGN PERMISSIONS TO ROLES
-- =============================================

-- Admin gets all MEDIPROMPT permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.resource IN ('rooms', 'devices', 'staff', 'patients', 'alerts', 'messages')
ON CONFLICT DO NOTHING;

-- Manager: staff management, room management, view alerts
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'manager' AND p.name IN (
    'rooms:read', 'rooms:write',
    'staff:read', 'staff:write', 'staff:assign_rooms',
    'patients:read',
    'alerts:read',
    'messages:read', 'messages:write',
    'profile:read', 'profile:write'
)
ON CONFLICT DO NOTHING;

-- Doctor: patient management, alerts, messages
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'doctor' AND p.name IN (
    'rooms:read',
    'staff:read',
    'patients:read', 'patients:write', 'patients:admit', 'patients:discharge',
    'alerts:read', 'alerts:create', 'alerts:acknowledge', 'alerts:resolve',
    'messages:read', 'messages:write',
    'profile:read', 'profile:write'
)
ON CONFLICT DO NOTHING;

-- Nurse: respond to alerts, view/update patients
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'nurse' AND p.name IN (
    'rooms:read',
    'staff:read',
    'patients:read', 'patients:write',
    'alerts:read', 'alerts:acknowledge', 'alerts:resolve',
    'messages:read', 'messages:write',
    'profile:read', 'profile:write'
)
ON CONFLICT DO NOTHING;

-- Patient: view own info, create alerts (call nurse)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'patient' AND p.name IN (
    'alerts:create',
    'messages:read', 'messages:write',
    'profile:read'
)
ON CONFLICT DO NOTHING;
