package service

import (
	"context"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
)

type CourseService struct {
	courseRepo repository.CourseRepository
}

func NewCourseService(
	courseRepo repository.CourseRepository,
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
