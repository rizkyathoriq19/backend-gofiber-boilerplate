-- Remove seeded users and their role assignments

-- Remove role assignments for seeded users
DELETE FROM user_roles WHERE user_id IN (
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000002'
);

-- Remove seeded users
DELETE FROM users WHERE id IN (
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000002'
);

