-- Permission pengelolaan portofolio mahasiswa
--   psql -U postgres -d api_students -f database/migrations/portfolio_permissions.sql
--
-- Portofolionya disimpan di MongoDB, tetapi hak aksesnya tetap dikelola di
-- PostgreSQL bersama permission lain. Membaca portofolio cukup dengan login;
-- membuat, mengganti, dan mengubahnya membutuhkan permission ini.
INSERT INTO permissions (name, description) VALUES
    ('portfolio:manage', 'Membuat, mengganti, dan mengubah portofolio mahasiswa')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'portfolio:manage'),
    ('staff', 'portfolio:manage')
ON CONFLICT DO NOTHING;