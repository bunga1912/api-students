package service

// CanAccessStudent menjawab: apakah pemanggil (currentUserID) boleh
// mengakses data mahasiswa yang dimiliki ownerID.
//
// Aturannya:
//   - Pemilik data SELALU boleh mengakses miliknya sendiri, tanpa
//     permission tambahan apa pun.
//   - Selain pemilik, hanya boleh mengakses bila punya izin ":any".
//   - Data tanpa pemilik tercatat (ownerID == nil) hanya boleh diakses
//     lewat izin ":any" — tidak ada yang otomatis memilikinya.
func CanAccessStudent(currentUserID int, ownerID *int, hasAnyPermission bool) bool {
	if ownerID != nil && *ownerID == currentUserID {
		return true
	}
	return hasAnyPermission
}