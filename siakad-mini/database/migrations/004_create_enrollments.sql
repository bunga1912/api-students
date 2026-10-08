CREATE TABLE IF NOT EXISTS enrollments (
    id             BIGSERIAL PRIMARY KEY,
    student_id     BIGINT NOT NULL REFERENCES students(id),
    course_id      BIGINT NOT NULL REFERENCES courses(id),
    tahun_akademik VARCHAR(20) NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT enrollments_unik UNIQUE (student_id, course_id, tahun_akademik)
);

-- Dipakai untuk menghitung kuota terisi per mata kuliah.
CREATE INDEX IF NOT EXISTS enrollments_course_idx
    ON enrollments (course_id, tahun_akademik);