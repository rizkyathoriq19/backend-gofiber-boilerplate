-- Rollback: Remove MEDIPROMPT roles and permissions

-- Remove role permissions for MEDIPROMPT resources
DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE resource IN ('rooms', 'devices', 'staff', 'patients', 'alerts', 'messages')
);

-- Remove MEDIPROMPT permissions
DELETE FROM permissions WHERE resource IN ('rooms', 'devices', 'staff', 'patients', 'alerts', 'messages');

-- Remove MEDIPROMPT roles (keep admin as it was there before)
DELETE FROM roles WHERE name IN ('nurse', 'doctor', 'patient', 'manager');
