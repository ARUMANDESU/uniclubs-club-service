-- bans table stores information about user bans in clubs
CREATE TABLE IF NOT EXISTS bans (
      id SERIAL PRIMARY KEY,
      user_id BIGINT NOT NULL,  -- ID of the user who is banned
      club_id BIGINT NOT NULL,  -- ID of the club where the user is banned
      admin_id BIGINT NOT NULL,  -- ID of the admin who banned the user
      reason TEXT NOT NULL,
      banned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
      FOREIGN KEY (user_id) REFERENCES users(id),
      FOREIGN KEY (club_id) REFERENCES clubs(id),
      FOREIGN KEY (admin_id) REFERENCES users(id),
      UNIQUE (user_id, club_id)  --Prevents duplicate bans
);

-- Indexes to speed up queries, but slow down writes
CREATE INDEX IF NOT EXISTS idx_bans_user_id ON bans(user_id);
CREATE INDEX IF NOT EXISTS idx_bans_club_id ON bans(club_id);
CREATE INDEX IF NOT EXISTS idx_bans_admin_id ON bans(admin_id);