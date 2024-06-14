
ALTER TABLE IF EXISTS clubs
    ADD COLUMN social_links TEXT[] DEFAULT '{}' not null ,
    ADD COLUMN location TEXT DEFAULT '' not null ;
