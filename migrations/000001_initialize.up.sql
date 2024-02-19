CREATE TABLE users (
   id BIGSERIAL PRIMARY KEY,
   email TEXT NOT NULL UNIQUE,
   barcode TEXT NOT NULL UNIQUE,
   first_name TEXT NOT NULL,
   last_name TEXT NOT NULL,
   avatar_url TEXT DEFAULT '' NOT NULL
);

CREATE TABLE clubs (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    owner_id BIGINT NOT NULL ,
    approved BOOLEAN DEFAULT false NOT NULL,
    description TEXT DEFAULT '' NOT NULL,
    type TEXT DEFAULT '' NOT NULL,
    logo_url TEXT DEFAULT '' NOT NULL,
    banner_url TEXT DEFAULT '' NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (owner_id) REFERENCES users(id)
);

CREATE TABLE roles (
    id BIGSERIAL PRIMARY KEY,
    club_id BIGINT REFERENCES clubs(id),
    name TEXT NOT NULL,
    permissions TEXT NOT NULL DEFAULT '0',
    position INT NOT NULL,
    color INT NOT NULL DEFAULT 8223868
);


CREATE TABLE clubs_users (
    user_id BIGINT NOT NULL,
    club_id BIGINT NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    PRIMARY KEY (user_id, club_id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (club_id) REFERENCES clubs(id)
);

CREATE TABLE users_roles (
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (role_id) REFERENCES roles(id)
);

CREATE TABLE create_club_requests (
    id SERIAL PRIMARY KEY,
    club_id BIGINT NOT NULL REFERENCES clubs(id),
    user_id BIGINT NOT NULL REFERENCES users(id),
    request_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE join_club_requests (
    id SERIAL PRIMARY KEY,
    club_id BIGINT NOT NULL REFERENCES clubs(id),
    user_id BIGINT NOT NULL REFERENCES users(id),
    request_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);