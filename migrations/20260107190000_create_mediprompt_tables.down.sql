-- Migration: drop mediprompt tables
-- Created at: 2026-01-07

-- Drop tables in reverse order (respect foreign keys)
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS alert_history;
DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS patients;
DROP TABLE IF EXISTS staff_room_assignments;
DROP TABLE IF EXISTS staff;
DROP TABLE IF EXISTS devices;
DROP TABLE IF EXISTS rooms;

-- Drop types
DROP TYPE IF EXISTS message_direction;
DROP TYPE IF EXISTS alert_status;
DROP TYPE IF EXISTS alert_priority;
DROP TYPE IF EXISTS alert_type;
DROP TYPE IF EXISTS condition_level;
DROP TYPE IF EXISTS shift_type;
DROP TYPE IF EXISTS staff_type;
DROP TYPE IF EXISTS device_status;
DROP TYPE IF EXISTS device_type;
DROP TYPE IF EXISTS room_type;
