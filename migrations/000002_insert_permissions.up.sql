BEGIN;

INSERT INTO permissions (name, description) VALUES
    -- Club Management
    ('CREATE_CLUB', 'Create a new club.'),
    ('EDIT_CLUB', 'Edit club details (name, description, logo, etc.).'),
    ('DEACTIVATE_CLUB', 'Deactivate a club.'),
    -- Membership Management
    ('ADD_MEMBERS', 'Add members to a club.'),
    ('REMOVE_MEMBERS', 'Remove members from a club.'),
    ('APPROVE_MEMBERSHIP', 'Approve membership requests.'),
    -- Roles and Permissions Management
    ('ASSIGN_ROLES', 'Assign roles to club members.'),
    ('EDIT_ROLES', 'Modify roles (change permissions, rename, etc.).'),
    ('DELETE_ROLES', 'Delete roles.'),
    -- Event Management
    ('CREATE_EVENTS', 'Create events for a club.'),
    ('EDIT_EVENTS', 'Edit details of club events.'),
    ('DELETE_EVENTS', 'Delete club events.'),
    -- Post and Content Management
    ('CREATE_POSTS', 'Create posts (announcements, polls, etc.) for a club.'),
    ('EDIT_POSTS', 'Edit posts made in a club.'),
    ('DELETE_POSTS', 'Delete posts in a club.');

COMMIT;