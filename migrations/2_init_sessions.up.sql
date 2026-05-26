-- =======================================================================
-- Создание таблицы sessions для хранения активных сессий пользователей
-- =======================================================================

-- Таблица для хранения данных о сессиях пользователей
CREATE TABLE sessions (
    -- Уникальный идентификатор сессии (UUID v4)
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Идентификатор пользователя, которому принадлежит сессия
    user_id                 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Идентификатор приложения, из которого выполнен вход:
    -- "web", "mobile", "admin"
    app_id                  VARCHAR(50) NOT NULL DEFAULT 'web',

    -- IP-адрес, с которого был выполнен вход
    user_ip                 VARCHAR(45),

    -- User-Agent устройства, с которого был выполнен вход
    user_agent              TEXT,

    -- SHA-256 хеш opaque refresh-токена
    -- Используется для поиска сессии при обновлении токенов
    refresh_token_hash      VARCHAR(64) NOT NULL UNIQUE,

    -- Время создания сессии
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Время истечения сессии
    -- После этой даты refresh-токен становится недействительным
    expires_at              TIMESTAMPTZ NOT NULL,

    -- Время последней активности в сессии
    -- Обновляется при каждом использовании refresh-токена
    last_activity_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);