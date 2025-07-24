-- Таблица сессий (Refresh-токены)
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(512) NOT NULL,  -- SHA-256 токена
    user_ip VARCHAR(45),                           -- IPv4/IPv6
    user_agent TEXT,                          -- Браузер/устройство
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);