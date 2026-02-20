-- Add collection_enabled to devices table (default true).
-- Run once against existing databases. New deployments may include this column in schema.
ALTER TABLE devices ADD COLUMN IF NOT EXISTS collection_enabled BOOLEAN DEFAULT true;
