-- +goose Up
INSERT INTO permissions (key, description) VALUES
    ('application.create', 'Create new applications under the project'),
    ('application.update', 'Edit application details'),
    ('application.delete', 'Delete applications'),
    ('notification.list', 'List project notifications'),
    ('notification.view', 'View notification details'),
    ('package.upload', 'Upload new application packages'),
    ('package.download', 'Download application packages'),
    ('member.invite', 'Invite new members to the project'),
    ('member.remove', 'Remove members from the project')
ON CONFLICT (key) DO NOTHING;

-- +goose Down
DELETE FROM permissions WHERE key IN (
    'application.create', 'application.update', 'application.delete',
    'notification.list', 'notification.view', 'package.upload',
    'package.download', 'member.invite', 'member.remove'
);
