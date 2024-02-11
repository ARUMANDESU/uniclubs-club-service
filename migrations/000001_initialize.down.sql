-- Remove foreign key constraints from 'join_club_requests' table
ALTER TABLE join_club_requests
    DROP CONSTRAINT IF EXISTS fk_join_club_requests_user,
    DROP CONSTRAINT IF EXISTS fk_join_club_requests_club;

-- Remove foreign key constraints from 'create_club_requests' table
ALTER TABLE create_club_requests
    DROP CONSTRAINT IF EXISTS fk_create_club_requests_user,
    DROP CONSTRAINT IF EXISTS fk_create_club_requests_club;

-- Remove foreign key constraints from 'clubs_users' table
ALTER TABLE clubs_users
    DROP CONSTRAINT IF EXISTS fk_clubs_users_role,
    DROP CONSTRAINT IF EXISTS fk_clubs_users_club,
    DROP CONSTRAINT IF EXISTS fk_clubs_users_user;

-- Remove foreign key constraints from 'roles_permissions' junction table
ALTER TABLE roles_permissions
    DROP CONSTRAINT IF EXISTS fk_roles_permissions_permission,
    DROP CONSTRAINT IF EXISTS fk_roles_permissions_role;

-- Remove foreign key constraint from 'roles' table
ALTER TABLE roles
    DROP CONSTRAINT IF EXISTS fk_roles_club;

-- Drop tables that were created
DROP TABLE IF EXISTS join_club_requests;
DROP TABLE IF EXISTS create_club_requests;
DROP TABLE IF EXISTS clubs_users;
DROP TABLE IF EXISTS roles_permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS clubs;
DROP TABLE IF EXISTS users;
