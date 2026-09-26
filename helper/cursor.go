package helper

import (
	"encoding/base64"
	"strconv"
)

// EncodeCursor mengubah id menjadi satu string yang aman di URL.
//
// base64 dipakai agar bentuk internalnya dapat kita ubah tanpa mengubah
// kontrak API. Perlu ditegaskan: base64 adalah PENGKODEAN, bukan
// enkripsi. Siapa pun dapat membacanya, dan siapa pun dapat menyusun
// cursor palsu. Karena itu cursor tidak boleh memuat apa pun yang
// bersifat rahasia atau yang dipercaya sebagai dasar hak akses.
func EncodeCursor(id int) string {
	raw := strconv.Itoa(id)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// DecodeCursor membaca kembali id dari string.
//
// Seluruh jalur kegagalan mengembalikan error, tidak ada yang "diperbaiki
// diam-diam". Cursor rusak berarti permintaan client memang salah, dan
// client berhak tahu itu lewat 400 — bukan diam-diam dikembalikan ke
// halaman pertama seolah tidak terjadi apa-apa.
func DecodeCursor(encoded string) (int, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return 0, BadRequest("cursor tidak valid")
	}

	id, err := strconv.Atoi(string(decoded))
	if err != nil || id < 1 {
		return 0, BadRequest("cursor tidak valid")
	}

	return id, nil
}