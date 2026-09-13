package helper

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

// bcryptCost menentukan berapa kali proses hashing diulang secara internal.
// Semakin besar, semakin lambat penyerang menebak password satu per satu.
// Nilai 12 dipilih sebagai kompromi: cukup lambat untuk menghambat brute
// force, masih cukup cepat agar login sekali tidak terasa berat.
const bcryptCost = 12

// dummyHash dipakai saat username tidak ditemukan, agar waktu tanggap
// login tetap mirip dengan kasus password salah. Tanpa ini, selisih
// waktu respons dapat membocorkan username mana yang terdaftar.
var dummyHash = []byte("$2a$12$abcdefghijklmnopqrstuuLKa3Bt1TCmU/6zvhZ8x4nq1yBiuGvS")

// HashPassword mengubah password menjadi hash yang tidak dapat dikembalikan.
// bcrypt menyisipkan salt acak ke dalam hasilnya, sehingga dua pengguna
// dengan password sama tetap menghasilkan hash yang berbeda.
func HashPassword(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// VerifyPassword membandingkan password dengan hash-nya.
func VerifyPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// VerifyDummyPassword sengaja membuang waktu seperti VerifyPassword,
// dipakai ketika username tidak ditemukan saat login.
func VerifyDummyPassword(plain string) {
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(plain))
}

// RandomToken menghasilkan string acak yang aman secara kriptografis.
// math/rand TIDAK dipakai di sini karena hasilnya dapat diperkirakan
// bila seed-nya diketahui — tidak layak untuk keperluan keamanan.
func RandomToken(numBytes int) (string, error) {
	buf := make([]byte, numBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// SHA256Hex dipakai untuk menyimpan refresh token dalam bentuk hash.
// bcrypt tidak dipakai di sini karena tokennya sudah acak dan panjang
// (32 byte), sehingga tidak perlu diperlambat seperti password buatan
// manusia yang rentan ditebak.
func SHA256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}