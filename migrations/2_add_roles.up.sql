-- Таблица пермишенов (все возможные действия в системе)
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(100) UNIQUE NOT NULL,
    description TEXT
);

-- Таблица ролей
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,  -- "admin", "manager", "user"
    description TEXT
);

-- Связь ролей и пермишенов (Many-to-Many)
CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- Связь пользователей и ролей (Many-to-Many)
CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- Добавляем пермишены
INSERT INTO permissions (code, description) VALUES
-- Пользователи
('user:create:own', 'Создать свой аккаунт'),
('user:read:own', 'Получить информацию о своём профиле'),
('user:update:own', 'Обновить свой профиль'),
('user:create:other', 'Создать любого пользователя'),
('user:read:other', 'Получить информацию о профиле любого пользователя'),
('user:update:other', 'Обновить любого пользователя'),
-- Платежи
('payment:create:own', 'Создать свой платёж'),
('payment:read:own', 'Получить информацию о своих платежах'),
('payment:update:own', 'Обновить свой платёж'),
('payment:cancel:own', 'Отменить свой платёж'),
('payment:create:other', 'Создать любой платёж'),
('payment:read:other', 'Получать информацию о любых платах'),
('payment:update:other', 'Обновить любой платёж'),
('payment:cancel:other', 'Отменить любой платёж'),
-- Роли
('role:read:own', 'Получить информацию о своих правах доступа'),
('role:create:other', 'Создать роль'),
('role:read:other', 'Получить информацию о правах доступа любого пользователя'),
('role:assign:other', 'Назначить роль любому пользователю'),
('role:revoke:other', 'Отозвать роль у любого пользователя');

-- Стандартные роли
INSERT INTO roles (name, description) VALUES
('admin', 'Полный доступ ко всем ресурсам(кроме создания собственных платежей)'),
('user', 'Обычный пользователь'),
('support', 'Может читать всё, но не изменять');

-- Админ
INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM roles WHERE name = 'admin'), 
    (SELECT id FROM permissions WHERE code LIKE '%:other' OR code IN ('user:read:own', 'role:read:own'));

-- Пользователь
INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM roles WHERE name = 'user'), 
    (SELECT id FROM permissions WHERE code NOT LIKE '%:other'); 

-- Поддержка
INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM roles WHERE name = 'support'), 
    (SELECT id FROM permissions WHERE code LIKE '%:read:other' OR code IN ('user:read:own', 'role:read:own'));

WITH new_user AS (
    INSERT INTO users (email, password_hash, full_name)
    VALUES (
        'admin@example.com',
        '$2a$10$xD6vWXJhC5r7Z4jzU4nYjeLq9v9k6QY8X0uVdJWcJk5p8rK1l2D3G',
        'Admin User'
    )
    RETURNING id
)

INSERT INTO user_roles (user_id, role_id)
SELECT new_user.id, roles.id
FROM new_user, roles
WHERE roles.name = 'admin';