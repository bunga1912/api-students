CREATE TABLE IF NOT EXISTS students (
    id         BIGSERIAL PRIMARY KEY,
    nim        VARCHAR(20) NOT NULL UNIQUE,
    name       VARCHAR(100) NOT NULL,
    grade      NUMERIC(5,2) NOT NULL DEFAULT 0 CHECK (grade >= 0 AND grade <= 100),
    is_active  BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_students_name ON students (name);