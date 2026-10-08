CREATE TABLE IF NOT EXISTS students (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL UNIQUE REFERENCES users(id),
    nim          VARCHAR(12) NOT NULL UNIQUE,
    nama         VARCHAR(100) NOT NULL,
    prodi        VARCHAR(100) NOT NULL,
    angkatan     INTEGER NOT NULL,
    ipk_terakhir NUMERIC(3,2) NOT NULL DEFAULT 0
                 CHECK (ipk_terakhir >= 0 AND ipk_terakhir <= 4),
    deleted_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS students_prodi_idx ON students (prodi);
CREATE INDEX IF NOT EXISTS students_angkatan_idx ON students (angkatan);