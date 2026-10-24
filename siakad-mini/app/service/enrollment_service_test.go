package service

import (
	"errors"
	"testing"

	"siakad-mini/app/model"
)

func TestValidateEnrollment(t *testing.T) {
	tests := []struct {
		name       string
		courseID   int64
		tahun      string
		wantFields []string
	}{
		{"valid ganjil", 1, "2026/2027-Ganjil", nil},
		{"valid genap", 1, "2026/2027-Genap", nil},
		{"course_id kosong", 0, "2026/2027-Ganjil", []string{"course_id"}},
		{"tahun kosong", 1, "", []string{"tahun_akademik"}},
		{"format salah", 1, "2026-Ganjil", []string{"tahun_akademik"}},
		{"semester salah", 1, "2026/2027-Pendek", []string{"tahun_akademik"}},
		{"semua salah", 0, "", []string{"course_id", "tahun_akademik"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEnrollment(tt.courseID, tt.tahun)

			if len(tt.wantFields) == 0 {
				if err != nil {
					t.Fatalf("seharusnya valid, dapat error: %v", err)
				}

				return
			}

			var verr *model.ValidationError
			if !errors.As(err, &verr) {
				t.Fatalf("seharusnya ValidationError, dapat: %v", err)
			}

			for _, f := range tt.wantFields {
				if _, ok := verr.Fields[f]; !ok {
					t.Errorf("field %q seharusnya error", f)
				}
			}

			if len(verr.Fields) != len(tt.wantFields) {
				t.Errorf("jumlah field error %d, mau %d", len(verr.Fields), len(tt.wantFields))
			}
		})
	}
}