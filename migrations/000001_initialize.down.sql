-- Start transaction for migration down script
BEGIN;

-- Drop the tables which will automatically remove any associated foreign key constraints
DROP TABLE IF EXISTS join_club_requests CASCADE;
DROP TABLE IF EXISTS create_club_requests CASCADE;
DROP TABLE IF EXISTS users_roles CASCADE;
DROP TABLE IF EXISTS clubs_users CASCADE;
DROP TABLE IF EXISTS roles_permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS clubs CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Commit the transaction to make sure all operations are executed atomically
COMMIT;