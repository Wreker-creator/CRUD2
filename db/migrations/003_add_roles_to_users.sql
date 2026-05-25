-- we are trying to add role based access to the current database,
-- so we need to add a new column to the users table
-- the default will be user and we can set it to admin later

ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'user';