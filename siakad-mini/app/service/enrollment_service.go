package service

import (
	"context"
	"errors"
	"fmt"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
)

type EnrollmentService struct {
	enrollmentRepo repository.EnrollmentRepository
	studentRepo    repository.StudentRepository
	courseRepo     repository.CourseRepository
}

func NewEnrollmentService(
	enrollmentRepo repository.EnrollmentRepository,
	studentRepo repository.StudentRepository,
	courseRepo repository.CourseRepository,
) *EnrollmentService {

	return &EnrollmentService{
		enrollmentRepo: enrollmentRepo,
		studentRepo:    studentRepo,
		courseRepo:     courseRepo,
	}
}

func (s *EnrollmentService) Create(
	ctx context.Context,
	studentID int64,
	courseID int64,
	tahunAkademik string,
) (*model.Enrollment, error) {

	if tahunAkademik == "" {
		return nil, errors.New("tahun akademik wajib diisi")
	}

	student, err := s.studentRepo.FindByID(ctx, studentID)
	if err != nil {
		return nil, err
	}

	course, err := s.courseRepo.FindByID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	// Cek apakah mahasiswa sudah mengambil mata kuliah ini.
	existing, err := s.enrollmentRepo.FindByStudentCourseYear(
		ctx,
		studentID,
		courseID,
		tahunAkademik,
	)

	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	if existing != nil {
		return nil, errors.New(
			"mata kuliah sudah diambil pada tahun akademik tersebut",
		)
	}

	// Cek kuota.
	terisi, err := s.enrollmentRepo.CountByCourse(
		ctx,
		courseID,
		tahunAkademik,
	)

	if err != nil {
		return nil, err
	}

	if terisi >= course.Kuota {
		return nil, fmt.Errorf(
			"kuota mata kuliah %s sudah penuh",
			course.KodeMK,
		)
	}

	// Hitung total SKS yang sudah diambil.
	currentSKS, err := s.enrollmentRepo.GetTotalSKS(
		ctx,
		studentID,
		tahunAkademik,
	)

	if err != nil {
		return nil, err
	}

	// Cek batas SKS berdasarkan IPK.
	if !CanTakeSKS(
		student.IPKTerakhir,
		currentSKS,
		course.SKS,
	) {
		maxSKS := MaxSKSByIPK(student.IPKTerakhir)

		return nil, fmt.Errorf(
			"pengambilan mata kuliah melebihi batas maksimal %d SKS",
			maxSKS,
		)
	}

	enrollment := &model.Enrollment{
		StudentID:     studentID,
		CourseID:      courseID,
		TahunAkademik: tahunAkademik,
	}

	return s.enrollmentRepo.Create(
		ctx,
		enrollment,
	)
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

func (s *EnrollmentService) Delete(
	ctx context.Context,
	id int64,
) error {

	return s.enrollmentRepo.Delete(ctx, id)
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
