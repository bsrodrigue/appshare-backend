-- +goose Up
INSERT INTO users (
    email,
    username,
    phone_number,
    password_hash,
    first_name,
    last_name
) VALUES 
('john.doe@example.com', 'johndoe', '+1234567890', '$2a$10$8K1p/a0dxvV/XmXzC/u0XO7E.xR.Q.X.X.X.X.X.X.X.X.X.X.X.', 'John', 'Doe'),
('jane.smith@example.com', 'janesmith', '+1987654321', '$2a$10$8K1p/a0dxvV/XmXzC/u0XO7E.xR.Q.X.X.X.X.X.X.X.X.X.X.X.', 'Jane', 'Smith'),
('bob.wilson@example.com', 'bobwilson', '+1122334455', '$2a$10$8K1p/a0dxvV/XmXzC/u0XO7E.xR.Q.X.X.X.X.X.X.X.X.X.X.X.', 'Bob', 'Wilson')
ON CONFLICT (email) DO NOTHING;

-- +goose Down
DELETE FROM users WHERE email IN ('john.doe@example.com', 'jane.smith@example.com', 'bob.wilson@example.com');
