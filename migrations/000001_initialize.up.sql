CREATE TABLE users(
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    barcode TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    avatar_url TEXT NOT NULL DEFAULT ''
);

CREATE TABLE clubs(
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    approved BOOLEAN NOT NULL DEFAULT false,
    description TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL DEFAULT '',
    logo_url TEXT NOT NULL DEFAULT '',
    banner_url TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE permissions(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE roles(
    id BIGSERIAL PRIMARY KEY,
    club_id BIGINT,
    name TEXT NOT NULL
);

CREATE TABLE roles_permissions(
    role_id BIGINT,
    permission_id INT
);

CREATE TABLE clubs_users(
    user_id BIGINT NOT NULL,
    club_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE create_club_requests(
    id SERIAL PRIMARY KEY,
    club_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    request_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE join_club_requests(
     id SERIAL PRIMARY KEY,
     club_id BIGINT NOT NULL,
     user_id BIGINT NOT NULL,
     request_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add foreign key to 'roles' table referencing 'clubs'
ALTER TABLE roles
    ADD CONSTRAINT fk_roles_club
        FOREIGN KEY (club_id)
            REFERENCES clubs(id);

-- Add foreign key constraints to 'roles_permissions' junction table
ALTER TABLE roles_permissions
    ADD CONSTRAINT fk_roles_permissions_role
        FOREIGN KEY (role_id)
            REFERENCES roles(id),
    ADD CONSTRAINT fk_roles_permissions_permission
        FOREIGN KEY (permission_id)
            REFERENCES permissions(id);

-- Add foreign key constraints to 'clubs_users' table
ALTER TABLE clubs_users
    ADD CONSTRAINT fk_clubs_users_user
        FOREIGN KEY (user_id)
            REFERENCES users(id),
    ADD CONSTRAINT fk_clubs_users_club
        FOREIGN KEY (club_id)
            REFERENCES clubs(id),
    ADD CONSTRAINT fk_clubs_users_role
        FOREIGN KEY (role_id)
            REFERENCES roles(id);

-- Add foreign key constraints to 'create_club_requests' table
ALTER TABLE create_club_requests
    ADD CONSTRAINT fk_create_club_requests_club
        FOREIGN KEY (club_id)
            REFERENCES clubs(id),
    ADD CONSTRAINT fk_create_club_requests_user
        FOREIGN KEY (user_id)
            REFERENCES users(id);

-- Add foreign key constraints to 'join_club_requests' table
ALTER TABLE join_club_requests
    ADD CONSTRAINT fk_join_club_requests_club
        FOREIGN KEY (club_id)
            REFERENCES clubs(id),
    ADD CONSTRAINT fk_join_club_requests_user
        FOREIGN KEY (user_id)
            REFERENCES users(id);