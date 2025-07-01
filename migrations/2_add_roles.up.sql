-- Таблица ролей
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,  -- "admin", "user"
    permissions   TEXT[] NOT NULL DEFAULT '{}',
    description TEXT
);

-- Связь пользователей и ролей (Many-to-Many)
CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- Добавляем стандартные роли
INSERT INTO roles (name, description) VALUES 
('admin', 'Администратор системы'),
('user', 'Обычный пользователь');