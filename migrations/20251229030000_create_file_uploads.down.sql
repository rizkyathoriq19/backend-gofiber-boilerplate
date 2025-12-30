-- Drop trigger
DROP TRIGGER IF EXISTS update_file_uploads_updated_at ON file_uploads;

-- Drop indexes
DROP INDEX IF EXISTS idx_file_uploads_deleted_at;
DROP INDEX IF EXISTS idx_file_uploads_created_at;
DROP INDEX IF EXISTS idx_file_uploads_storage_type;
DROP INDEX IF EXISTS idx_file_uploads_mime_type;
DROP INDEX IF EXISTS idx_file_uploads_user_id;

-- Drop table
DROP TABLE IF EXISTS file_uploads;
