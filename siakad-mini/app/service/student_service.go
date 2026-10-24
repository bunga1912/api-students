package service

import (
	"context"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"siakad-mini/app/model"
)

const (
	maxPerPage = 50
	maxPage    = 1_000_000 // mencegah overflow pada OFFSET
)

var nimPattern = regexp.MustCompile(`^\d{12}$`)

type StudentService struct {
	studentRepo    model.StudentRepository
	enrollmentRepo model.EnrollmentRepository
	tx             model.TxManager
}

func NewStudentService(
	studentRepo model.StudentRepository,
	enrollmentRepo model.EnrollmentRepository,
	tx model.TxManager,
) *StudentService {
	return &StudentService{
		studentRepo:    studentRepo,
		enrollmentRepo: enrollmentRepo,
		tx:             tx,
	}
}

// CreateStudentInput adalah data untuk menambah mahasiswa sekaligus akun user-nya.
type CreateStudentInput struct {
	NIM         string
	Nama        string
	Email       string
	Prodi       string
	Angkatan    int
	IPKTerakhir *float64 // opsional
}

// UpdateStudentInput adalah data untuk memperbarui mahasiswa. NIM tidak bisa diubah.
type UpdateStudentInput struct {
	Nama        string
	Prodi       string
	Angkatan    int
	IPKTerakhir *float64 // opsional: nil berarti IPK tidak diubah
}

// StudentDetail adalah mahasiswa beserta KRS-nya.
type StudentDetail struct {
	Student     *model.Student
	Enrollments []model.EnrollmentDetail
	TotalSKS    int
	BatasSKS    int
}

// ---------- validasi ----------

func requireText(verr *model.ValidationError, field, value string, max int) {
	switch {
	case value == "":
		verr.Add(field, "wajib diisi")
	case utf8.RuneCountInString(value) > max:
		verr.Add(field, fmt.Sprintf("maksimal %d karakter", max))
	}
}

// validateProfile memeriksa field yang sama pada tambah dan ubah mahasiswa.
func validateProfile(
	verr *model.ValidationError,
	nama string,
	prodi string,
	angkatan int,
	ipk *float64,
	currentYear int,
) {
	requireText(verr, "nama", nama, 100)
	requireText(verr, "prodi", prodi, 100)

	switch {
	case angkatan == 0:
		verr.Add("angkatan", "wajib diisi")
	case angkatan < 1000 || angkatan > 9999:
		verr.Add("angkatan", "harus 4 digit")
	case angkatan > currentYear:
		verr.Add("angkatan", "tidak boleh lebih dari tahun berjalan")
	}

	if ipk != nil && !(*ipk >= 0 && *ipk <= 4) {
		verr.Add("ipk_terakhir", "harus antara 0,00 dan 4,00")
	}
}

func validateCreateStudent(in CreateStudentInput, currentYear int) error {
	verr := model.NewValidationError()

	switch {
	case in.NIM == "":
		verr.Add("nim", "wajib diisi")
	case !nimPattern.MatchString(in.NIM):
		verr.Add("nim", "harus 12 digit angka")
	}

	switch {
	case in.Email == "":
		verr.Add("email", "wajib diisi")
	case utf8.RuneCountInString(in.Email) > 100:
		verr.Add("email", "maksimal 100 karakter")
	default:
		addr, err := mail.ParseAddress(in.Email)
		if err != nil || addr.Address != in.Email {
			verr.Add("email", "format email tidak valid")
		}
	}

	validateProfile(verr, in.Nama, in.Prodi, in.Angkatan, in.IPKTerakhir, currentYear)

	if verr.HasErrors() {
		return verr
	}

	return nil
}

func lastPage(total, perPage int) int {
	if total <= 0 || perPage <= 0 {
		return 1
	}

	return (total + perPage - 1) / perPage
}

// ---------- use case ----------

// List mengembalikan satu halaman mahasiswa sesuai filter.
func (s *StudentService) List(
	ctx context.Context,
	f model.StudentFilter,
) (*model.StudentPage, error) {

	f.Prodi = strings.TrimSpace(f.Prodi)
	f.Search = strings.TrimSpace(f.Search)
	f.Sort = strings.TrimSpace(f.Sort)

	verr := model.NewValidationError()

	switch {
	case f.Page < 1:
		verr.Add("page", "minimal 1")
	case f.Page > maxPage:
		verr.Add("page", "terlalu besar")
	}

	if f.PerPage < 1 {
		verr.Add("per_page", "minimal 1")
	}

	if f.Angkatan < 0 {
		verr.Add("angkatan", "tidak valid")
	}

	switch f.Sort {
	case "", model.StudentSortNama, model.StudentSortIPKDesc:
	default:
		verr.Add("sort", "harus 'nama' atau '-ipk_terakhir'")
	}

	if verr.HasErrors() {
		return nil, verr
	}

	if f.PerPage > maxPerPage {
		f.PerPage = maxPerPage
	}

	items, total, err := s.studentRepo.ListPaged(ctx, f)
	if err != nil {
		return nil, err
	}

	return &model.StudentPage{
		Items:    items,
		Total:    total,
		Page:     f.Page,
		PerPage:  f.PerPage,
		LastPage: lastPage(total, f.PerPage),
	}, nil
}

func (s *StudentService) GetByID(
	ctx context.Context,
	id int64,
) (*model.Student, error) {

	return s.studentRepo.FindByID(ctx, id)
}

func (s *StudentService) GetByUserID(
	ctx context.Context,
	userID int64,
) (*model.Student, error) {

	return s.studentRepo.FindByUserID(ctx, userID)
}

// GetDetail mengembalikan mahasiswa beserta mata kuliah yang diambil, total SKS,
// dan batas SKS berdasarkan IPK. tahunAkademik kosong berarti semua tahun.
func (s *StudentService) GetDetail(
	ctx context.Context,
	id int64,
	tahunAkademik string,
) (*StudentDetail, error) {

	tahunAkademik = strings.TrimSpace(tahunAkademik)

	if tahunAkademik != "" && !tahunAkademikPattern.MatchString(tahunAkademik) {
		return nil, model.FieldError(
			"tahun_akademik",
			"format harus seperti 2026/2027-Ganjil",
		)
	}

	student, err := s.studentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	enrollments, err := s.enrollmentRepo.ListDetailByStudent(ctx, id, tahunAkademik)
	if err != nil {
		return nil, err
	}

	total := 0
	for _, e := range enrollments {
		total += e.SKS
	}

	return &StudentDetail{
		Student:     student,
		Enrollments: enrollments,
		TotalSKS:    total,
		BatasSKS:    MaxSKSByIPK(student.IPKTerakhir),
	}, nil
}

// Create membuat akun user (role mahasiswa, password awal = NIM yang di-hash)
// dan data mahasiswa dalam satu transaction.
func (s *StudentService) Create(
	ctx context.Context,
	in CreateStudentInput,
) (*model.Student, error) {

	in.NIM = strings.TrimSpace(in.NIM)
	in.Nama = strings.TrimSpace(in.Nama)
	in.Email = strings.TrimSpace(in.Email)
	in.Prodi = strings.TrimSpace(in.Prodi)

	if err := validateCreateStudent(in, time.Now().Year()); err != nil {
		return nil, err
	}

	// Hashing sengaja di luar transaction supaya koneksi DB tidak tertahan.
	hash, err := bcrypt.GenerateFromPassword([]byte(in.NIM), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	ipk := 0.0
	if in.IPKTerakhir != nil {
		ipk = *in.IPKTerakhir
	}

	var created *model.Student

	err = s.tx.WithinTx(ctx, func(repos model.TxRepositories) error {

		user, err := repos.Users.Create(ctx, &model.User{
			Email:    in.Email,
			Password: string(hash),
			Role:     model.RoleMahasiswa,
		})

		if err != nil {
			return err
		}

		student, err := repos.Students.Create(ctx, &model.Student{
			UserID:      user.ID,
			NIM:         in.NIM,
			Nama:        in.Nama,
			Prodi:       in.Prodi,
			Angkatan:    in.Angkatan,
			IPKTerakhir: ipk,
		})

		if err != nil {
			return err
		}

		created = student

		return nil
	})

	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *StudentService) Update(
	ctx context.Context,
	id int64,
	in UpdateStudentInput,
) (*model.Student, error) {

	existing, err := s.studentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	in.Nama = strings.TrimSpace(in.Nama)
	in.Prodi = strings.TrimSpace(in.Prodi)

	verr := model.NewValidationError()

	validateProfile(verr, in.Nama, in.Prodi, in.Angkatan, in.IPKTerakhir, time.Now().Year())

	if verr.HasErrors() {
		return nil, verr
	}

	existing.Nama = in.Nama
	existing.Prodi = in.Prodi
	existing.Angkatan = in.Angkatan

	if in.IPKTerakhir != nil {
		existing.IPKTerakhir = *in.IPKTerakhir
	}

	return s.studentRepo.Update(ctx, existing)
}

// Delete melakukan soft delete.
func (s *StudentService) Delete(
	ctx context.Context,
	id int64,
) error {

	return s.studentRepo.Delete(ctx, id)
}