-- Migration: add vital_signs table
-- Created at: 2026-01-13

-- =============================================
-- VITAL SIGNS TABLE
-- =============================================
CREATE TABLE vital_signs (
    id UUID PRIMARY KEY,
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    recorded_by_staff_id UUID REFERENCES staff(id) ON DELETE SET NULL,
    
    -- Vital measurements
    heart_rate INT,                    -- BPM (beats per minute)
    blood_pressure_sys INT,            -- Systolic mmHg
    blood_pressure_dia INT,            -- Diastolic mmHg
    temperature DECIMAL(4,1),          -- Celsius
    oxygen_saturation INT,             -- SpO2 percentage
    respiratory_rate INT,              -- Breaths per minute
    pain_level INT,                    -- 0-10 scale
    
    notes TEXT,
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_vital_signs_patient_id ON vital_signs(patient_id);
CREATE INDEX idx_vital_signs_recorded_by ON vital_signs(recorded_by_staff_id);
CREATE INDEX idx_vital_signs_recorded_at ON vital_signs(recorded_at);
