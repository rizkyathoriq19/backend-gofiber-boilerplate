-- Seeder: MEDIPROMPT sample data
-- Created at: 2026-01-07
-- Password: password123 (same bcrypt hash as existing seeds)

-- =============================================
-- SEED SAMPLE USERS FOR MEDIPROMPT
-- =============================================
INSERT INTO users (id, name, email, password, role) VALUES
    -- Manager
    ('00000000-0000-0000-0000-000000000010', 
     'Ward Manager', 
     'manager@hospital.com', 
     '$2a$12$Q9IdH6PWbwol9aZYgHslM.VfkVMBqEL3HyceYr9Pa8JYuCpHTIXym', 
     'manager'),
    -- Doctor
    ('00000000-0000-0000-0000-000000000011', 
     'Dr. John Smith', 
     'doctor@hospital.com', 
     '$2a$12$Q9IdH6PWbwol9aZYgHslM.VfkVMBqEL3HyceYr9Pa8JYuCpHTIXym', 
     'doctor'),
    -- Nurses
    ('00000000-0000-0000-0000-000000000012', 
     'Nurse Alice', 
     'nurse1@hospital.com', 
     '$2a$12$Q9IdH6PWbwol9aZYgHslM.VfkVMBqEL3HyceYr9Pa8JYuCpHTIXym', 
     'nurse'),
    ('00000000-0000-0000-0000-000000000013', 
     'Nurse Bob', 
     'nurse2@hospital.com', 
     '$2a$12$Q9IdH6PWbwol9aZYgHslM.VfkVMBqEL3HyceYr9Pa8JYuCpHTIXym', 
     'nurse')
ON CONFLICT (email) DO NOTHING;

-- Assign roles to users
INSERT INTO user_roles (user_id, role_id)
SELECT '00000000-0000-0000-0000-000000000010', r.id FROM roles r WHERE r.name = 'manager'
ON CONFLICT DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
SELECT '00000000-0000-0000-0000-000000000011', r.id FROM roles r WHERE r.name = 'doctor'
ON CONFLICT DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
SELECT '00000000-0000-0000-0000-000000000012', r.id FROM roles r WHERE r.name = 'nurse'
ON CONFLICT DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
SELECT '00000000-0000-0000-0000-000000000013', r.id FROM roles r WHERE r.name = 'nurse'
ON CONFLICT DO NOTHING;

-- =============================================
-- SEED SAMPLE ROOMS
-- =============================================
INSERT INTO rooms (id, name, type, floor, building, capacity, is_active) VALUES
    ('00000000-0000-0000-0001-000000000001', 'Room 101', 'patient_room', '1', 'Main', 2, true),
    ('00000000-0000-0000-0001-000000000002', 'Room 102', 'patient_room', '1', 'Main', 2, true),
    ('00000000-0000-0000-0001-000000000003', 'Room 103', 'patient_room', '1', 'Main', 1, true),
    ('00000000-0000-0000-0001-000000000004', 'ICU 1', 'icu', '2', 'Main', 1, true),
    ('00000000-0000-0000-0001-000000000005', 'ICU 2', 'icu', '2', 'Main', 1, true),
    ('00000000-0000-0000-0001-000000000006', 'Nurse Station 1F', 'nurse_station', '1', 'Main', 5, true),
    ('00000000-0000-0000-0001-000000000007', 'Emergency Room', 'emergency', '1', 'Main', 10, true)
ON CONFLICT (id) DO NOTHING;

-- =============================================
-- SEED STAFF RECORDS
-- =============================================
INSERT INTO staff (id, user_id, employee_id, type, department, shift, on_duty, phone) VALUES
    ('00000000-0000-0000-0002-000000000001', '00000000-0000-0000-0000-000000000010', 'MGR-001', 'manager', 'Administration', 'morning', true, '+62811111111'),
    ('00000000-0000-0000-0002-000000000002', '00000000-0000-0000-0000-000000000011', 'DOC-001', 'doctor', 'General Medicine', 'morning', true, '+62822222222'),
    ('00000000-0000-0000-0002-000000000003', '00000000-0000-0000-0000-000000000012', 'NRS-001', 'nurse', 'General Ward', 'morning', true, '+62833333333'),
    ('00000000-0000-0000-0002-000000000004', '00000000-0000-0000-0000-000000000013', 'NRS-002', 'nurse', 'General Ward', 'afternoon', false, '+62844444444')
ON CONFLICT (user_id) DO NOTHING;

-- =============================================
-- ASSIGN STAFF TO ROOMS
-- =============================================
INSERT INTO staff_room_assignments (staff_id, room_id, is_primary) VALUES
    -- Nurse Alice assigned to Room 101, 102, 103
    ('00000000-0000-0000-0002-000000000003', '00000000-0000-0000-0001-000000000001', true),
    ('00000000-0000-0000-0002-000000000003', '00000000-0000-0000-0001-000000000002', false),
    ('00000000-0000-0000-0002-000000000003', '00000000-0000-0000-0001-000000000003', false),
    -- Nurse Bob assigned to ICU 1, ICU 2
    ('00000000-0000-0000-0002-000000000004', '00000000-0000-0000-0001-000000000004', true),
    ('00000000-0000-0000-0002-000000000004', '00000000-0000-0000-0001-000000000005', false),
    -- Doctor assigned to all patient rooms
    ('00000000-0000-0000-0002-000000000002', '00000000-0000-0000-0001-000000000001', false),
    ('00000000-0000-0000-0002-000000000002', '00000000-0000-0000-0001-000000000002', false),
    ('00000000-0000-0000-0002-000000000002', '00000000-0000-0000-0001-000000000004', true),
    ('00000000-0000-0000-0002-000000000002', '00000000-0000-0000-0001-000000000005', false)
ON CONFLICT (staff_id, room_id) DO NOTHING;

-- =============================================
-- SEED SAMPLE PATIENTS
-- =============================================
INSERT INTO patients (id, room_id, medical_record_number, name, date_of_birth, gender, condition_level, diagnosis, notes) VALUES
    ('00000000-0000-0000-0003-000000000001', '00000000-0000-0000-0001-000000000001', 'MRN-2026-0001', 'Patient Ahmad', '1985-05-15', 'male', 'stable', 'Post-surgery recovery', 'Day 3 post-op, recovering well'),
    ('00000000-0000-0000-0003-000000000002', '00000000-0000-0000-0001-000000000001', 'MRN-2026-0002', 'Patient Siti', '1990-08-22', 'female', 'moderate', 'Pneumonia', 'On antibiotics, monitor oxygen levels'),
    ('00000000-0000-0000-0003-000000000003', '00000000-0000-0000-0001-000000000004', 'MRN-2026-0003', 'Patient Budi', '1975-12-01', 'male', 'critical', 'Heart failure', 'ICU monitoring, critical care required'),
    ('00000000-0000-0000-0003-000000000004', '00000000-0000-0000-0001-000000000002', 'MRN-2026-0004', 'Patient Dewi', '2000-03-10', 'female', 'good', 'Observation', 'Minor injury, observation only')
ON CONFLICT (medical_record_number) DO NOTHING;

-- =============================================
-- SEED SAMPLE DEVICES
-- =============================================
INSERT INTO devices (id, room_id, type, serial_number, name, status, config) VALUES
    ('00000000-0000-0000-0004-000000000001', '00000000-0000-0000-0001-000000000001', 'microphone', 'MIC-101-001', 'Room 101 Microphone', 'online', '{"sensitivity": "high", "language": "id"}'),
    ('00000000-0000-0000-0004-000000000002', '00000000-0000-0000-0001-000000000001', 'teleprompter', 'TEL-101-001', 'Room 101 Display', 'online', '{"mode": "nurse_call", "brightness": 80}'),
    ('00000000-0000-0000-0004-000000000003', '00000000-0000-0000-0001-000000000004', 'microphone', 'MIC-ICU1-001', 'ICU 1 Microphone', 'online', '{"sensitivity": "very_high", "language": "id"}'),
    ('00000000-0000-0000-0004-000000000004', '00000000-0000-0000-0001-000000000004', 'button', 'BTN-ICU1-001', 'ICU 1 Emergency Button', 'online', '{"type": "emergency"}'),
    ('00000000-0000-0000-0004-000000000005', '00000000-0000-0000-0001-000000000006', 'teleprompter', 'TEL-NS1-001', 'Nurse Station Display', 'online', '{"mode": "dashboard", "brightness": 100}')
ON CONFLICT (serial_number) DO NOTHING;
