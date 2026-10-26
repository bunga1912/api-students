-- Migration 007: pembimbing tugas akhir (D.2)
-- psql -U postgres -d api_students -f migrations/007_supervisors.sql

CREATE TABLE IF NOT EXISTS supervisors (
    id               SERIAL PRIMARY KEY,
    name             VARCHAR(120) NOT NULL,
    quota            INTEGER NOT NULL CHECK (quota > 0),
    supervised_count INTEGER NOT NULL DEFAULT 0,
    version          INTEGER NOT NULL DEFAULT 1,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT supervisors_name_unique UNIQUE (name),

    -- Jaring pengaman terakhir, sama seperti courses_enrolled_within_quota.
    -- Bila kode aplikasi benar, constraint ini TIDAK PERNAH terpicu.
    CONSTRAINT supervisors_count_within_quota
        CHECK (supervised_count >= 0 AND supervised_count <= quota)
);

CREATE TABLE IF NOT EXISTS supervisions (
    -- PRIMARY KEY pada student_id: satu mahasiswa hanya boleh muncul SATU
    -- kali di tabel ini, sehingga hanya boleh memiliki SATU pembimbing.
    -- Aturan ini dijaga database, bukan hanya kode: dua request ganda yang
    -- datang bersamaan tidak mungkin sama-sama berhasil.
    student_id    BIGINT PRIMARY KEY REFERENCES students(id) ON DELETE CASCADE,
    supervisor_id INTEGER NOT NULL REFERENCES supervisors(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS supervisions_supervisor_idx ON supervisions (supervisor_id);

INSERT INTO supervisors (name, quota) VALUES
    ('Dr. Andi Pratama', 10),
    ('Dr. Budi Santoso', 10),
    ('Dr. Citra Lestari', 10),
    ('Dr. Rebutan (uji beban)', 3)
ON CONFLICT (name) DO NOTHING;

INSERT INTO permissions (name, description) VALUES
    ('supervision:manage', 'Menetapkan, mengganti, dan mengubah kuota pembimbing tugas akhir')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'supervision:manage')
ON CONFLICT DO NOTHING;