# API Students

REST API sederhana untuk mengelola data mahasiswa (CRUD), dibangun dengan Go dan Fiber v2.
Dibuat untuk Tugas 2 mata kuliah Praktikum Pemrograman Backend Lanjut.

## Tools yang digunakan
- Go 1.26.5
- Fiber v2.52.15

## Cara menjalankan

Server berjalan di `http://localhost:8000`.

## Kontrak API

Semua response menggunakan format baku:
```json
{ "success": true, "message": "...", "data": { ... } }
```
Response daftar menambahkan `meta`, response gagal menggunakan `success: false` dan bisa menyertakan `errors`.

| Method | Endpoint | Parameter | Contoh Body Request | Status Mungkin | Contoh Response |
|---|---|---|---|---|---|
| GET | `/api/v1/students` | Query: `page` (default 1), `limit` (default 10, maks 50), `search` (nama, tidak case-sensitive), `sort` (`id`\|`nim`\|`name`\|`grade`, default `id`), `order` (`asc`\|`desc`, default `asc`), `is_active` (`true`\|`false`) | – | 200 | `{"success":true,"message":"daftar mahasiswa berhasil diambil","data":[{"id":1,"nim":"2201001","name":"Sari","grade":88.5,"is_active":true}],"meta":{"page":1,"limit":10,"total":1,"total_pages":1}}` |
| GET | `/api/v1/students/:id` | Path: `id` (angka positif) | – | 200, 400, 404 | `{"success":true,"message":"mahasiswa ditemukan","data":{"id":1,"nim":"2201001","name":"Sari","grade":88.5,"is_active":true}}` |
| POST | `/api/v1/students` | Header: `Content-Type: application/json` | `{"nim":"2201001","name":"Sari","grade":88.5}` | 201, 400, 409, 415, 422 | `{"success":true,"message":"mahasiswa berhasil ditambahkan","data":{"id":1,"nim":"2201001","name":"Sari","grade":88.5,"is_active":true}}` (header `Location: /api/v1/students/1`) |
| PUT | `/api/v1/students/:id` | Path: `id`. Header: `Content-Type: application/json`. Seluruh field wajib dikirim | `{"nim":"2201001","name":"Sari Wijaya","grade":95.0,"is_active":false}` | 200, 400, 404, 409, 415, 422 | `{"success":true,"message":"mahasiswa berhasil diganti seluruhnya","data":{"id":1,"nim":"2201001","name":"Sari Wijaya","grade":95.0,"is_active":false}}` |
| PATCH | `/api/v1/students/:id` | Path: `id`. Header: `Content-Type: application/json`. Hanya field yang ingin diubah yang perlu dikirim | `{"is_active":true}` | 200, 400, 404, 409, 415, 422 | `{"success":true,"message":"mahasiswa berhasil diperbarui sebagian","data":{"id":1,"nim":"2201001","name":"Sari Wijaya","grade":95.0,"is_active":true}}` |
| DELETE | `/api/v1/students/:id` | Path: `id` (angka positif) | – | 204, 400, 404 | *(tanpa body)* |

## Keterangan Status

| Status | Arti |
|---|---|
| 200 | Permintaan berhasil |
| 201 | Data baru berhasil dibuat |
| 204 | Berhasil dihapus, tanpa isi respons |
| 400 | Body bukan JSON valid, atau `id` bukan angka |
| 404 | Mahasiswa tidak ditemukan |
| 409 | NIM sudah dipakai mahasiswa lain |
| 415 | `Content-Type` bukan `application/json` |
| 422 | Validasi isi gagal (field wajib kosong, grade di luar 0–100, dsb) |

## Catatan Batas Limit

Batas atas `limit` ditetapkan **50** karena data mahasiswa dalam satu daftar kelas/angkatan secara realistis tidak akan mencapai ratusan atau ribuan baris sekaligus, sehingga batas ini cukup longgar untuk pemakaian wajar namun tetap mencegah klien meminta seluruh data sekaligus dan membebani server.