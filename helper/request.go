package helper

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
)

// ParamID membaca parameter :id dari jalur dan memastikan bentuknya benar.
func ParamID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

// Daftar putih field yang boleh dipakai untuk mengurutkan.
var allowedSort = map[string]bool{
	"id": true, "nim": true, "name": true, "grade": true,
}

// ParseListQuery membaca query string dan memberi nilai bawaan yang aman.
// Batas atas limit ditetapkan 50: data mahasiswa realistis tidak akan
// mencapai ratusan ribu baris dalam satu daftar kelas/angkatan, sehingga
// batas ini cukup longgar untuk pemakaian wajar namun tetap mencegah
// klien meminta seluruh data sekaligus dan membebani server.
func ParseListQuery(c *fiber.Ctx) model.ListQuery {
	q := model.ListQuery{
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", 10),
		Search: strings.TrimSpace(c.Query("search")),
		Sort:   c.Query("sort", "id"),
		Order:  strings.ToLower(c.Query("order", "asc")),
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 10
	}
	if q.Limit > 50 { // batas atas wajib ada
		q.Limit = 50
	}
	if !allowedSort[q.Sort] { // daftar putih, bukan daftar hitam
		q.Sort = "id"
	}
	if q.Order != "desc" {
		q.Order = "asc"
	}
	if raw := c.Query("is_active"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil {
			q.IsActive = &v
		}
	}
	return q
}