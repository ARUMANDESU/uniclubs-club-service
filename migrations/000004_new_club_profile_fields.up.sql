
ALTER TABLE IF EXISTS clubs
    ADD COLUMN social_links TEXT[] DEFAULT '{}',
    ADD COLUMN location TEXT;
