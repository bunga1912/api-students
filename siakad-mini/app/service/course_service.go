package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"siakad-mini/app/model"
)

type CourseService struct {
	courseRepo model.CourseRepository
}

func NewCourseService(
	courseRepo model.CourseRepository,
) *CourseService {
	return &CourseService{
		courseRepo: courseRepo,
	}
}

func (s *CourseService) GetByID(
	ctx context.Context,
	id int64,
) (*model.Course, error) {

	return s.courseRepo.FindByID(ctx, id)
}

func (s *CourseService) GetByCode(
	ctx context.Context,
	code string,
) (*model.Course, error) {

	return s.courseRepo.FindByCode(ctx, code)
}

// ListWithQuota mengembalikan mata kuliah beserta terisi dan sisa_kuota,
// dengan filter semester, search, dan hanya yang masih tersedia.
func (s *CourseService) ListWithQuota(
	ctx context.Context,
	f model.CourseFilter,
) ([]model.CourseWithQuota, error) {

	f.Search = strings.TrimSpace(f.Search)
	f.TahunAkademik = strings.TrimSpace(f.TahunAkademik)

	verr := model.NewValidationError()

	if f.Semester < 0 {
		verr.Add("semester", "minimal 1")
	}

	if utf8.RuneCountInString(f.Search) > 100 {
		verr.Add("search", "maksimal 100 karakter")
	}

	if f.TahunAkademik != "" && !tahunAkademikPattern.MatchString(f.TahunAkademik) {
		verr.Add("tahun_akademik", "format harus seperti 2026/2027-Ganjil")
	}

	if verr.HasErrors() {
		return nil, verr
	}

	return s.courseRepo.ListWithQuota(ctx, f)
}

func (s *CourseService) List(
	ctx context.Context,
) ([]model.Course, error) {

	return s.courseRepo.List(ctx)
}

func (s *CourseService) ListBySemester(
	ctx context.Context,
	semester int,
) ([]model.Course, error) {

	return s.courseRepo.ListBySemester(ctx, semester)
}

func (s *CourseService) Create(
	ctx context.Context,
	course *model.Course,
) (*model.Course, error) {

	return s.courseRepo.Create(ctx, course)
}

func (s *CourseService) Update(
	ctx context.Context,
	course *model.Course,
) (*model.Course, error) {

	return s.courseRepo.Update(ctx, course)
}

func (s *CourseService) Delete(
	ctx context.Context,
	id int64,
) error {

	return s.courseRepo.Delete(ctx, id)
}