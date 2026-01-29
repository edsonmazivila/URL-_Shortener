-- Create urls table
CREATE TABLE IF NOT EXISTS urls (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(20) NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    access_count BIGINT NOT NULL DEFAULT 0,
    last_accessed TIMESTAMP WITH TIME ZONE,
    CONSTRAINT short_code_length_check CHECK (LENGTH(short_code) >= 3)
);

-- Create indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls(short_code);
CREATE INDEX IF NOT EXISTS idx_urls_created_at ON urls(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_urls_expires_at ON urls(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_urls_access_count ON urls(access_count DESC);

-- Add comments for documentation
COMMENT ON TABLE urls IS 'Stores shortened URLs and their metadata';
COMMENT ON COLUMN urls.id IS 'Primary key, auto-incrementing identifier';
COMMENT ON COLUMN urls.short_code IS 'Unique short code for the URL';
COMMENT ON COLUMN urls.original_url IS 'Original long URL';
COMMENT ON COLUMN urls.created_at IS 'Timestamp when the URL was created';
COMMENT ON COLUMN urls.expires_at IS 'Optional expiration timestamp';
COMMENT ON COLUMN urls.access_count IS 'Number of times the URL has been accessed';
COMMENT ON COLUMN urls.last_accessed IS 'Timestamp of the last access';
