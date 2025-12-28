-- URL shortener for compact link display in agent responses
CREATE TABLE IF NOT EXISTS "URL_Shortener" (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(10) NOT NULL UNIQUE,
    full_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_url_shortener_code ON "URL_Shortener"(code);
CREATE INDEX IF NOT EXISTS idx_url_shortener_expires ON "URL_Shortener"(expires_at);
