package service

import (
	"testing"

	"api-students/app/model"
)

// Pengujian ini tidak menyalakan server, tidak menyentuh database,
// dan tidak membuat fiber.Ctx — murni menguji fungsi Go biasa.

func TestValidateCreate(t *testing.T) {
	cases := []struct {
		name    string
		req     model.CreateStudentRequest
		wantErr bool
		field   string // field yang diharapkan error, kalau wantErr true
	}{
		{
			name:    "valid",
			req:     model.CreateStudentRequest{NIM: "2201001", Name: "Sari", Grade: 88.5},
			wantErr: false,
		},
		{
			name:    "nim kosong",
			req:     model.CreateStudentRequest{NIM: "", Name: "Sari", Grade: 88.5},
			wantErr: true,
			field:   "nim",
		},
		{
			name:    "name kosong",
			req:     model.CreateStudentRequest{NIM: "2201001", Name: "  ", Grade: 88.5},
			wantErr: true,
			field:   "name",
		},
		{
			name:    "grade di luar rentang",
			req:     model.CreateStudentRequest{NIM: "2201001", Name: "Sari", Grade: 150},
			wantErr: true,
			field:   "grade",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateCreate(tc.req)
			if tc.wantErr && len(errs) == 0 {
				t.Fatalf("harap ada error, tapi lolos validasi")
			}
			if !tc.wantErr && len(errs) != 0 {
				t.Fatalf("tidak harap error, tapi dapat: %v", errs)
			}
			if tc.wantErr {
				if _, ok := errs[tc.field]; !ok {
					t.Errorf("harap error pada field %q, dapat: %v", tc.field, errs)
				}
			}
		})
	}
}

func TestValidateReplace(t *testing.T) {
	cases := []struct {
		name    string
		req     model.ReplaceStudentRequest
		wantErr bool
	}{
		{
			name:    "valid",
			req:     model.ReplaceStudentRequest{NIM: "2201001", Name: "Sari", Grade: 90, IsActive: true},
			wantErr: false,
		},
		{
			name:    "nim kosong",
			req:     model.ReplaceStudentRequest{NIM: "", Name: "Sari", Grade: 90},
			wantErr: true,
		},
		{
			name:    "grade negatif",
			req:     model.ReplaceStudentRequest{NIM: "2201001", Name: "Sari", Grade: -1},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateReplace(tc.req)
			if tc.wantErr && len(errs) == 0 {
				t.Fatalf("harap ada error, tapi lolos validasi")
			}
			if !tc.wantErr && len(errs) != 0 {
				t.Fatalf("tidak harap error, tapi dapat: %v", errs)
			}
		})
	}
}

func TestApplyPatch(t *testing.T) {
	initial := model.Student{
		ID: 1, NIM: "2201001", Name: "Sari", Grade: 88.5, IsActive: true,
	}

	t.Run("ubah is_active saja", func(t *testing.T) {
		inactive := false
		result, errs := ApplyPatch(initial, model.PatchStudentRequest{IsActive: &inactive})

		if len(errs) != 0 {
			t.Fatalf("tidak seharusnya ada error: %v", errs)
		}
		if result.IsActive {
			t.Error("is_active seharusnya berubah menjadi false")
		}
		if result.Name != "Sari" {
			t.Error("field yang tidak dikirim seharusnya tidak berubah")
		}
	})

	t.Run("grade di luar rentang ditolak", func(t *testing.T) {
		grade := 999.0
		_, errs := ApplyPatch(initial, model.PatchStudentRequest{Grade: &grade})

		if _, ok := errs["grade"]; !ok {
			t.Errorf("harap ada error pada field grade, dapat: %v", errs)
		}
	})

	t.Run("nim kosong ditolak", func(t *testing.T) {
		empty := "   "
		_, errs := ApplyPatch(initial, model.PatchStudentRequest{NIM: &empty})

		if _, ok := errs["nim"]; !ok {
			t.Errorf("harap ada error pada field nim, dapat: %v", errs)
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