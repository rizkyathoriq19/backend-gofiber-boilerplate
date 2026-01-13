-- Migration: create mediprompt tables
-- Created at: 2026-01-07

-- =============================================
-- ROOMS TABLE
-- =============================================
CREATE TYPE room_type AS ENUM ('patient_room', 'nurse_station', 'icu', 'emergency', 'operating_room');

CREATE TABLE rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    type room_type NOT NULL DEFAULT 'patient_room',
    floor VARCHAR(20) NOT NULL,
    building VARCHAR(100),
    capacity INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rooms_type ON rooms(type);
CREATE INDEX idx_rooms_floor ON rooms(floor);
CREATE INDEX idx_rooms_is_active ON rooms(is_active);

CREATE TRIGGER update_rooms_updated_at 
    BEFORE UPDATE ON rooms 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- =============================================
-- DEVICES TABLE
-- =============================================
CREATE TYPE device_type AS ENUM ('microphone', 'teleprompter', 'button', 'sensor');
CREATE TYPE device_status AS ENUM ('online', 'offline', 'maintenance', 'error');

CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID REFERENCES rooms(id) ON DELETE SET NULL,
    type device_type NOT NULL,
    serial_number VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100),
    status device_status NOT NULL DEFAULT 'offline',
    api_key_hash VARCHAR(255),
    config JSONB DEFAULT '{}',
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_devices_room_id ON devices(room_id);
CREATE INDEX idx_devices_type ON devices(type);
CREATE INDEX idx_devices_status ON devices(status);
CREATE INDEX idx_devices_serial_number ON devices(serial_number);

CREATE TRIGGER update_devices_updated_at 
    BEFORE UPDATE ON devices 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- =============================================
-- STAFF TABLE (extends users)
-- =============================================
CREATE TYPE staff_type AS ENUM ('nurse', 'doctor', 'manager', 'admin');
CREATE TYPE shift_type AS ENUM ('morning', 'afternoon', 'night', 'on_call');

CREATE TABLE staff (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    employee_id VARCHAR(50) UNIQUE NOT NULL,
    type staff_type NOT NULL,
    department VARCHAR(100),
    shift shift_type DEFAULT 'morning',
    on_duty BOOLEAN DEFAULT false,
    phone VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

CREATE INDEX idx_staff_user_id ON staff(user_id);
CREATE INDEX idx_staff_type ON staff(type);
CREATE INDEX idx_staff_on_duty ON staff(on_duty);
CREATE INDEX idx_staff_shift ON staff(shift);

CREATE TRIGGER update_staff_updated_at 
    BEFORE UPDATE ON staff 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- =============================================
-- STAFF ROOM ASSIGNMENTS (for room-based permissions)
-- =============================================
CREATE TABLE staff_room_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    staff_id UUID NOT NULL REFERENCES staff(id) ON DELETE CASCADE,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    is_primary BOOLEAN DEFAULT false,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(staff_id, room_id)
);

CREATE INDEX idx_staff_room_assignments_staff_id ON staff_room_assignments(staff_id);
CREATE INDEX idx_staff_room_assignments_room_id ON staff_room_assignments(room_id);

-- =============================================
-- PATIENTS TABLE
-- =============================================
CREATE TYPE condition_level AS ENUM ('critical', 'serious', 'moderate', 'stable', 'good');

CREATE TABLE patients (
    id UUID PRIMARY KEY ,
    room_id UUID REFERENCES rooms(id) ON DELETE SET NULL,
    medical_record_number VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    date_of_birth DATE,
    gender VARCHAR(10),
    condition_level condition_level NOT NULL DEFAULT 'stable',
    diagnosis TEXT,
    notes TEXT,
    admission_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    discharge_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_patients_room_id ON patients(room_id);
CREATE INDEX idx_patients_medical_record_number ON patients(medical_record_number);
CREATE INDEX idx_patients_condition_level ON patients(condition_level);
CREATE INDEX idx_patients_admission_date ON patients(admission_date);

CREATE TRIGGER update_patients_updated_at 
    BEFORE UPDATE ON patients 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- =============================================
-- ALERTS TABLE (Core MEDIPROMPT)
-- =============================================
CREATE TYPE alert_type AS ENUM ('voice_call', 'button_press', 'emergency', 'system', 'scheduled');
CREATE TYPE alert_priority AS ENUM ('critical', 'high', 'medium', 'low');
CREATE TYPE alert_status AS ENUM ('pending', 'acknowledged', 'in_progress', 'resolved', 'escalated', 'cancelled');

CREATE TABLE alerts (
    id UUID PRIMARY KEY ,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    patient_id UUID REFERENCES patients(id) ON DELETE SET NULL,
    device_id UUID REFERENCES devices(id) ON DELETE SET NULL,
    assigned_staff_id UUID REFERENCES staff(id) ON DELETE SET NULL,
    resolved_by_staff_id UUID REFERENCES staff(id) ON DELETE SET NULL,
    
    type alert_type NOT NULL,
    priority alert_priority NOT NULL DEFAULT 'medium',
    status alert_status NOT NULL DEFAULT 'pending',
    
    message TEXT,
    detected_keywords TEXT[],
    audio_reference VARCHAR(255),
    
    escalation_count INTEGER DEFAULT 0,
    escalation_timeout_minutes INTEGER DEFAULT 5,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_alerts_room_id ON alerts(room_id);
CREATE INDEX idx_alerts_patient_id ON alerts(patient_id);
CREATE INDEX idx_alerts_assigned_staff_id ON alerts(assigned_staff_id);
CREATE INDEX idx_alerts_type ON alerts(type);
CREATE INDEX idx_alerts_priority ON alerts(priority);
CREATE INDEX idx_alerts_status ON alerts(status);
CREATE INDEX idx_alerts_created_at ON alerts(created_at);
CREATE INDEX idx_alerts_status_priority ON alerts(status, priority);

CREATE TRIGGER update_alerts_updated_at 
    BEFORE UPDATE ON alerts 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- =============================================
-- ALERT HISTORY TABLE (for audit trail)
-- =============================================
CREATE TABLE alert_history (
    id UUID PRIMARY KEY ,
    alert_id UUID NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    staff_id UUID REFERENCES staff(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL,
    previous_status alert_status,
    new_status alert_status,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_alert_history_alert_id ON alert_history(alert_id);
CREATE INDEX idx_alert_history_staff_id ON alert_history(staff_id);
CREATE INDEX idx_alert_history_created_at ON alert_history(created_at);

-- =============================================
-- MESSAGES TABLE (two-way communication)
-- =============================================
CREATE TYPE message_direction AS ENUM ('to_patient', 'from_patient', 'staff_to_staff');

CREATE TABLE messages (
    id UUID PRIMARY KEY ,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    patient_id UUID REFERENCES patients(id) ON DELETE SET NULL,
    sender_staff_id UUID REFERENCES staff(id) ON DELETE SET NULL,
    receiver_staff_id UUID REFERENCES staff(id) ON DELETE SET NULL,
    
    direction message_direction NOT NULL,
    content TEXT NOT NULL,
    is_read BOOLEAN DEFAULT false,
    is_urgent BOOLEAN DEFAULT false,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    read_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_messages_room_id ON messages(room_id);
CREATE INDEX idx_messages_patient_id ON messages(patient_id);
CREATE INDEX idx_messages_sender_staff_id ON messages(sender_staff_id);
CREATE INDEX idx_messages_is_read ON messages(is_read);
CREATE INDEX idx_messages_created_at ON messages(created_at);
