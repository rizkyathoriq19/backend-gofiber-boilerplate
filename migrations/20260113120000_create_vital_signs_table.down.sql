-- Down migration: drop vital_signs table

DROP INDEX IF EXISTS idx_vital_signs_recorded_at;
DROP INDEX IF EXISTS idx_vital_signs_recorded_by;
DROP INDEX IF EXISTS idx_vital_signs_patient_id;
DROP TABLE IF EXISTS vital_signs;
