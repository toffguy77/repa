ALTER TABLE groups ADD COLUMN categories text[] NOT NULL DEFAULT '{}'::text[];
