package service

import (
	"testing"

	"api-students/app/model"
)

// Pengujian ini tidak menyalakan server, tidak menyentuh database,
// dan tidak membuat fiber.Ctx — murni menguji fungsi Go biasa.

func TestApplyPatch(t *testing.T) {
	initial := model.Student{
		ID: 1, NIM: "2201001", Name: "Sari", Grade: 88.5, IsActive: true,
	}

	t.Run("ubah is_active saja", func(t *testing.T) {
		inactive := false
		result := ApplyPatch(initial, model.PatchStudentRequest{IsActive: &inactive})

		if result.IsActive {
			t.Error("is_active seharusnya berubah menjadi false")
		}
		if result.Name != "Sari" {
			t.Error("field yang tidak dikirim seharusnya tidak berubah")
		}
	})

	t.Run("ubah beberapa field sekaligus", func(t *testing.T) {
		nim := "2201099"
		name := "Budi"
		result := ApplyPatch(initial, model.PatchStudentRequest{NIM: &nim, Name: &name})

		if result.NIM != "2201099" {
			t.Error("nim seharusnya berubah")
		}
		if result.Name != "Budi" {
			t.Error("name seharusnya berubah")
		}
		if result.Grade != 88.5 {
			t.Error("grade yang tidak dikirim seharusnya tidak berubah")
		}
	})
}

func TestIsEmptyPatch(t *testing.T) {
	if !IsEmptyPatch(model.PatchStudentRequest{}) {
		t.Error("request kosong harap dianggap empty patch")
	}

	name := "Sari"
	if IsEmptyPatch(model.PatchStudentRequest{Name: &name}) {
		t.Error("request dengan satu field terisi tidak boleh dianggap empty patch")
	}
}

func TestCountTotalPages(t *testing.T) {
	cases := []struct{ total, limit, want int }{
		{0, 10, 0},
		{1, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{137, 20, 7},
	}
	for _, tc := range cases {
		if got := CountTotalPages(tc.total, tc.limit); got != tc.want {
			t.Errorf("total=%d limit=%d: harap %d, dapat %d",
				tc.total, tc.limit, tc.want, got)
		}
	}
}