-- Start transaction for migration down script
BEGIN;

-- Drop the indexes
DROP INDEX IF EXISTS idx_bans_user_id;
DROP INDEX IF EXISTS idx_bans_club_id;
DROP INDEX IF EXISTS idx_bans_admin_id;

-- Drop the bans table which will automatically remove any associated foreign key constraints
DROP TABLE IF EXISTS bans CASCADE;

-- Commit the transaction to make sure all operations are executed atomically
COMMIT;