package service

import (
	"errors"
	"testing"

	"siakad-mini/app/model"
)

func ptr(f float64) *float64 { return &f }

func validInput() CreateStudentInput {
	return CreateStudentInput{
		NIM:         "187221000001",
		Nama:        "Rina Putri",
		Email:       "rina@siakad.test",
		Prodi:       "Sistem Informasi",
		Angkatan:    2022,
		IPKTerakhir: ptr(3.45),
	}
}

func TestValidateCreateStudent(t *testing.T) {
	const year = 2026

	tests := []struct {
		name       string
		mutate     func(in *CreateStudentInput)
		wantFields []string
	}{
		{"valid", func(in *CreateStudentInput) {}, nil},
		{"ipk opsional", func(in *CreateStudentInput) { in.IPKTerakhir = nil }, nil},
		{"angkatan tahun berjalan", func(in *CreateStudentInput) { in.Angkatan = 2026 }, nil},
		{"nim kosong", func(in *CreateStudentInput) { in.NIM = "" }, []string{"nim"}},
		{"nim 11 digit", func(in *CreateStudentInput) { in.NIM = "18722100000" }, []string{"nim"}},
		{"nim ada huruf", func(in *CreateStudentInput) { in.NIM = "18722100000A" }, []string{"nim"}},
		{"email kosong", func(in *CreateStudentInput) { in.Email = "" }, []string{"email"}},
		{"email salah format", func(in *CreateStudentInput) { in.Email = "bukan-email" }, []string{"email"}},
		{"nama kosong", func(in *CreateStudentInput) { in.Nama = "" }, []string{"nama"}},
		{"prodi kosong", func(in *CreateStudentInput) { in.Prodi = "" }, []string{"prodi"}},
		{"angkatan kosong", func(in *CreateStudentInput) { in.Angkatan = 0 }, []string{"angkatan"}},
		{"angkatan 3 digit", func(in *CreateStudentInput) { in.Angkatan = 999 }, []string{"angkatan"}},
		{"angkatan masa depan", func(in *CreateStudentInput) { in.Angkatan = 2027 }, []string{"angkatan"}},
		{"ipk lebih dari 4", func(in *CreateStudentInput) { in.IPKTerakhir = ptr(4.01) }, []string{"ipk_terakhir"}},
		{"ipk negatif", func(in *CreateStudentInput) { in.IPKTerakhir = ptr(-0.1) }, []string{"ipk_terakhir"}},
		{"ipk batas 0 dan 4", func(in *CreateStudentInput) { in.IPKTerakhir = ptr(4) }, nil},
		{
			"banyak field salah",
			func(in *CreateStudentInput) { in.NIM = "1"; in.Email = "x"; in.Nama = "" },
			[]string{"nim", "email", "nama"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validInput()
			tt.mutate(&in)

			err := validateCreateStudent(in, year)

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
				t.Errorf("jumlah field error %d, mau %d (%v)", len(verr.Fields), len(tt.wantFields), verr.Fields)
			}
		})
	}
}

func TestLastPage(t *testing.T) {
	tests := []struct {
		total, perPage, want int
	}{
		{0, 10, 1},
		{1, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{20, 10, 2},
		{21, 10, 3},
		{20, 50, 1},
	}

	for _, tt := range tests {
		if got := lastPage(tt.total, tt.perPage); got != tt.want {
			t.Errorf("lastPage(%d, %d) = %d, mau %d", tt.total, tt.perPage, got, tt.want)
		}
	}
}