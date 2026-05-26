-- =======================================================================
-- Создание таблицы users для хранения учётных записей пользователей
-- =======================================================================

-- Таблица для хранения данных о пользователях системы
CREATE TABLE users (
    -- Уникальный идентификатор пользователя (UUID v4)
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Электронная почта пользователя (уникальна в рамках системы)
    email               VARCHAR(255) UNIQUE NOT NULL,

    -- Хеш пароля (Argon2id)
    password_hash       VARCHAR(255) NOT NULL,

    -- Отображаемое имя пользователя
    full_name           VARCHAR(255) NOT NULL DEFAULT '',

    -- Подтверждён ли email:
    -- FALSE — email не подтверждён, требуется верификация
    -- TRUE — email подтверждён
    email_verified      BOOLEAN NOT NULL DEFAULT FALSE,

    -- Статус пользователя:
    -- FALSE — активен, вход разрешён
    -- TRUE — заблокирован, вход запрещён, существующие токены инвалидированы
    disabled            BOOLEAN NOT NULL DEFAULT FALSE,

    -- Время создания учётной записи
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Время последнего изменения учётной записи
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Время последнего успешного входа
    -- Может быть NULL, если пользователь ни разу не входил
    last_login_at       TIMESTAMPTZ
);