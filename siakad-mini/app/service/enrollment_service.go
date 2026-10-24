package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"siakad-mini/app/model"
)

// Format tahun akademik, contoh: 2026/2027-Ganjil
var tahunAkademikPattern = regexp.MustCompile(`^\d{4}/\d{4}-(Ganjil|Genap)$`)

type EnrollmentService struct {
	enrollmentRepo model.EnrollmentRepository
	tx             model.TxManager
}

func NewEnrollmentService(
	enrollmentRepo model.EnrollmentRepository,
	tx model.TxManager,
) *EnrollmentService {

	return &EnrollmentService{
		enrollmentRepo: enrollmentRepo,
		tx:             tx,
	}
}

func validateEnrollment(courseID int64, tahunAkademik string) error {
	verr := model.NewValidationError()

	if courseID <= 0 {
		verr.Add("course_id", "wajib diisi")
	}

	if tahunAkademik == "" {
		verr.Add("tahun_akademik", "wajib diisi")
	} else if !tahunAkademikPattern.MatchString(tahunAkademik) {
		verr.Add("tahun_akademik", "format harus seperti 2026/2027-Ganjil")
	}

	if verr.HasErrors() {
		return verr
	}

	return nil
}

// Create mengambil mata kuliah untuk mahasiswa. Seluruh pengecekan dan insert
// berjalan dalam satu transaction.
//
// Urutan penguncian SELALU mahasiswa -> mata kuliah:
//   - kunci mahasiswa  : dua request dari mahasiswa yang sama tidak bisa
//     sama-sama lolos cek batas SKS.
//   - kunci mata kuliah: dua request ke mata kuliah yang sama tidak bisa
//     sama-sama lolos cek kuota.
//
// Urutan yang konsisten mencegah deadlock.
func (s *EnrollmentService) Create(
	ctx context.Context,
	studentID int64,
	courseID int64,
	tahunAkademik string,
) (*model.Enrollment, error) {

	if err := validateEnrollment(courseID, tahunAkademik); err != nil {
		return nil, err
	}

	var result *model.Enrollment

	err := s.tx.WithinTx(ctx, func(repos model.TxRepositories) error {

		student, err := repos.Students.FindByIDForUpdate(ctx, studentID)
		if err != nil {
			return err
		}

		course, err := repos.Courses.FindByIDForUpdate(ctx, courseID)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return model.FieldError("course_id", "mata kuliah tidak ditemukan")
			}

			return err
		}

		// Cek apakah mahasiswa sudah mengambil mata kuliah ini.
		existing, err := repos.Enrollments.FindByStudentCourseYear(
			ctx,
			studentID,
			courseID,
			tahunAkademik,
		)

		if err != nil && !errors.Is(err, model.ErrNotFound) {
			return err
		}

		if existing != nil {
			return model.ErrSudahDiambil
		}

		// Cek kuota.
		terisi, err := repos.Enrollments.CountByCourse(
			ctx,
			courseID,
			tahunAkademik,
		)

		if err != nil {
			return err
		}

		if terisi >= course.Kuota {
			return fmt.Errorf(
				"%w: %s",
				model.ErrKuotaPenuh,
				course.KodeMK,
			)
		}

		// Hitung total SKS yang sudah diambil.
		currentSKS, err := repos.Enrollments.GetTotalSKS(
			ctx,
			studentID,
			tahunAkademik,
		)

		if err != nil {
			return err
		}

		// Cek batas SKS berdasarkan IPK.
		if !CanTakeSKS(
			student.IPKTerakhir,
			currentSKS,
			course.SKS,
		) {
			maxSKS := MaxSKSByIPK(student.IPKTerakhir)

			sisa := maxSKS - currentSKS
			if sisa < 0 {
				sisa = 0
			}

			return fmt.Errorf(
				"%w: batas %d SKS, sisa %d SKS, mata kuliah ini %d SKS",
				model.ErrSKSMelebihi,
				maxSKS,
				sisa,
				course.SKS,
			)
		}

		created, err := repos.Enrollments.Create(
			ctx,
			&model.Enrollment{
				StudentID:     studentID,
				CourseID:      courseID,
				TahunAkademik: tahunAkademik,
			},
		)

		if err != nil {
			return err
		}

		result = created

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *EnrollmentService) ListByStudent(
	ctx context.Context,
	studentID int64,
	tahunAkademik string,
) ([]model.EnrollmentDetail, error) {

	return s.enrollmentRepo.ListByStudent(
		ctx,
		studentID,
		tahunAkademik,
	)
}

func (s *EnrollmentService) GetTotalSKS(
	ctx context.Context,
	studentID int64,
	tahunAkademik string,
) (int, error) {

	return s.enrollmentRepo.GetTotalSKS(
		ctx,
		studentID,
		tahunAkademik,
	)
}

// Delete membatalkan enrollment milik mahasiswa studentID.
// ErrNotFound jika enrollment tidak ada, ErrForbidden jika milik mahasiswa lain.
// Sisa kuota dihitung dari tabel enrollments, jadi otomatis bertambah.
func (s *EnrollmentService) Delete(
	ctx context.Context,
	studentID int64,
	id int64,
) error {

	enrollment, err := s.enrollmentRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if enrollment.StudentID != studentID {
		return model.ErrForbidden
	}

	return s.enrollmentRepo.Delete(ctx, id)
}