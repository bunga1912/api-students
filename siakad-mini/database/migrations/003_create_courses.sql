CREATE TABLE IF NOT EXISTS courses (
    id       BIGSERIAL PRIMARY KEY,
    kode_mk  VARCHAR(20) NOT NULL UNIQUE,
    nama_mk  VARCHAR(150) NOT NULL,
    sks      SMALLINT NOT NULL CHECK (sks > 0),
    semester SMALLINT NOT NULL CHECK (semester > 0),
    kuota    INTEGER NOT NULL CHECK (kuota >= 0)
);

CREATE INDEX IF NOT EXISTS courses_semester_idx ON courses (semester);