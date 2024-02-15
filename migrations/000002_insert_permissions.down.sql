BEGIN;

DELETE FROM permissions WHERE name IN (
    'CREATE_CLUB',
    'EDIT_CLUB',
    'DEACTIVATE_CLUB',
    'ADD_MEMBERS',
    'REMOVE_MEMBERS',
    'APPROVE_MEMBERSHIP',
    'ASSIGN_ROLES',
    'EDIT_ROLES',
    'DELETE_ROLES',
    'CREATE_EVENTS',
    'EDIT_EVENTS',
    'DELETE_EVENTS',
    'CREATE_POSTS',
    'EDIT_POSTS',
    'DELETE_POSTS'
);

COMMIT;
