ALTER TABLE market ADD CONSTRAINT market_name_unique UNIQUE (name);
-- to enforce uniqueness at database level.