CREATE TABLE IF NOT EXISTS users (
    id         BIGSERIAL PRIMARY KEY,
    email      VARCHAR(100) NOT NULL,
    password   TEXT NOT NULL,
    role       VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'mahasiswa')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Email unik tanpa membedakan huruf besar dan kecil.
CREATE UNIQUE INDEX IF NOT EXISTS users_email_lower_key ON users (LOWER(email));