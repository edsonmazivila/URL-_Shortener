-- Drop indexes
DROP INDEX IF EXISTS idx_urls_access_count;
DROP INDEX IF EXISTS idx_urls_expires_at;
DROP INDEX IF EXISTS idx_urls_created_at;
DROP INDEX IF EXISTS idx_urls_short_code;

-- Drop table
DROP TABLE IF EXISTS urls;
