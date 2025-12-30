-- Create file_uploads table for file management
CREATE TABLE IF NOT EXISTS file_uploads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    original_name VARCHAR(255) NOT NULL,
    stored_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    extension VARCHAR(20),
    storage_type VARCHAR(20) NOT NULL DEFAULT 'local', -- 'local', 's3', 'minio'
    bucket_name VARCHAR(100),
    is_public BOOLEAN DEFAULT false,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes
CREATE INDEX idx_file_uploads_user_id ON file_uploads(user_id);
CREATE INDEX idx_file_uploads_mime_type ON file_uploads(mime_type);
CREATE INDEX idx_file_uploads_storage_type ON file_uploads(storage_type);
CREATE INDEX idx_file_uploads_created_at ON file_uploads(created_at DESC);
CREATE INDEX idx_file_uploads_deleted_at ON file_uploads(deleted_at) WHERE deleted_at IS NULL;

-- Update trigger
CREATE TRIGGER update_file_uploads_updated_at
    BEFORE UPDATE ON file_uploads
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
