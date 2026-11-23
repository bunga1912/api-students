-- Migration 009: satu event outbox boleh menghasilkan banyak notifikasi
-- psql -U postgres -d <nama_database> -f migrations/009_notifications_per_recipient.sql
--
-- Sebelumnya: UNIQUE (outbox_id) -> satu event hanya boleh punya SATU notifikasi.
-- Sekarang  : UNIQUE (outbox_id, user_id) -> satu event boleh punya satu
--             notifikasi UNTUK SETIAP penerima, tetapi tidak boleh dua untuk
--             penerima yang sama.
--
-- Idempotency key-nya kini berupa pasangan (event, penerima), bukan lagi event saja.

BEGIN;

ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_outbox_unique;
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_outbox_user_unique;
ALTER TABLE notifications ADD CONSTRAINT notifications_outbox_user_unique UNIQUE (outbox_id, user_id);

COMMIT;