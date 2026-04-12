-- market is the main table storing all food items.
-- id is managed entirely by Postgres and never exposed through the API.
-- name is used as the public identifier across all routes and CLI commands.
CREATE TABLE IF NOT EXISTS market (
    id          SERIAL PRIMARY KEY,                  -- auto-incrementing internal ID
    name        VARCHAR(255) NOT NULL,               -- public identifier, used in all queries
    price       DECIMAL(10,2) NOT NULL,              -- 2 decimal places, e.g. 1.99
    calories    INT NOT NULL,
    sugar       DECIMAL(10,2) NOT NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- set once on insert, never touched again
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP  -- kept current by the trigger below
);

-- this function is called by the trigger on every UPDATE.
-- NEW refers to the row with the incoming updated values.
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW; -- RETURN NEW is required — it tells Postgres to proceed with the update
END;
$$ LANGUAGE plpgsql;

-- the trigger fires BEFORE each row UPDATE so updated_at is set before the row is written.
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON market
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();