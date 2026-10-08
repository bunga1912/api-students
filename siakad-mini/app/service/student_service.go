package service

import (
	"context"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
)

type StudentService struct {
	studentRepo repository.StudentRepository
}

func NewStudentService(
	studentRepo repository.StudentRepository,
) *StudentService {
	return &StudentService{
		studentRepo: studentRepo,
	}
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

func (s *StudentService) GetByNIM(
	ctx context.Context,
	nim string,
) (*model.Student, error) {

	return s.studentRepo.FindByNIM(ctx, nim)
}

func (s *StudentService) List(
	ctx context.Context,
) ([]model.Student, error) {

	return s.studentRepo.List(ctx)
}

func (s *StudentService) Create(
	ctx context.Context,
	student *model.Student,
) (*model.Student, error) {

	return s.studentRepo.Create(ctx, student)
}

func (s *StudentService) Update(
	ctx context.Context,
	student *model.Student,
) (*model.Student, error) {

	return s.studentRepo.Update(ctx, student)
}

func (s *StudentService) Delete(
	ctx context.Context,
	id int64,
) error {

	return s.studentRepo.Delete(ctx, id)
}
