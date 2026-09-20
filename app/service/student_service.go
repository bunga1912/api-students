package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

// StudentService memegang dua peran sekaligus (controller + orkestrasi),
// sama seperti UserService pada modul.
type StudentService struct {
	repo  repository.StudentRepository
	perms *helper.PermissionSet
}

func NewStudentService(repo repository.StudentRepository, perms *helper.PermissionSet) *StudentService {
	return &StudentService{repo: repo, perms: perms}
}

// toModel menerjemahkan representasi repository ke representasi keluar.
func toModel(s repository.Student) model.Student {
	return model.Student{
		ID: s.ID, NIM: s.NIM, Name: s.Name, Grade: s.Grade, IsActive: s.IsActive,
	}
}

func (s *StudentService) List(c *fiber.Ctx) error {
	q := helper.ParseListQuery(c)

	filter := repository.StudentListFilter{
		Page: q.Page, Limit: q.Limit, Search: q.Search,
		Sort: q.Sort, Order: q.Order, IsActive: q.IsActive,
	}

	hasil, total, err := s.repo.FindAll(c.Context(), filter)
	if err != nil {
		return helper.Fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dijangkau")
	}

	data := make([]model.Student, 0, len(hasil))
	for _, item := range hasil {
		data = append(data, toModel(item))
	}

	return helper.SuccessList(c, "daftar mahasiswa berhasil diambil", data, &model.Meta{
		Page: q.Page, Limit: q.Limit, Total: total,
		TotalPages: CountTotalPages(total, q.Limit),
	})
}

func (s *StudentService) Get(c *fiber.Ctx) error {
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	item, err := s.repo.FindByID(c.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return helper.Fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}
	if err != nil {
		return helper.Fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dijangkau")
	}
	return helper.Success(c, fiber.StatusOK, "mahasiswa ditemukan", toModel(item))
}

func (s *StudentService) Create(c *fiber.Ctx) error {
	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if errs := ValidateCreate(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	baru, err := s.repo.Create(c.Context(), repository.Student{
		NIM: req.NIM, Name: req.Name, Grade: req.Grade,
	})
	if errors.Is(err, repository.ErrDuplicate) {
		return helper.Fail(c, fiber.StatusConflict, "NIM sudah terdaftar")
	}
	if err != nil {
		return helper.Fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dijangkau")
	}

	return helper.Created(c, "mahasiswa berhasil ditambahkan", toModel(baru),
		"/api/v1/students/"+strconv.Itoa(baru.ID))
}

func (s *StudentService) Replace(c *fiber.Ctx) error {
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if errs := ValidateReplace(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	hasil, err := s.repo.Update(c.Context(), repository.Student{
		ID: id, NIM: req.NIM, Name: req.Name, Grade: req.Grade, IsActive: req.IsActive,
	})
	if errors.Is(err, repository.ErrNotFound) {
		return helper.Fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}
	if errors.Is(err, repository.ErrDuplicate) {
		return helper.Fail(c, fiber.StatusConflict, "NIM sudah dipakai mahasiswa lain")
	}
	if err != nil {
		return helper.Fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dijangkau")
	}

	return helper.Success(c, fiber.StatusOK, "mahasiswa berhasil diganti seluruhnya", toModel(hasil))
}

func (s *StudentService) Patch(c *fiber.Ctx) error {
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	if IsEmptyPatch(req) {
		return helper.Fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	existingRepo, err := s.repo.FindByID(c.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return helper.Fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}
	if err != nil {
		return helper.Fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dijangkau")
	}

	updated, errs := ApplyPatch(toModel(existingRepo), req)
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	hasil, err := s.repo.Update(c.Context(), repository.Student{
		ID: updated.ID, NIM: updated.NIM, Name: updated.Name,
		Grade: updated.Grade, IsActive: updated.IsActive,
	})
	if errors.Is(err, repository.ErrDuplicate) {
		return helper.Fail(c, fiber.StatusConflict, "NIM sudah dipakai mahasiswa lain")
	}
	if err != nil {
		return helper.Fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dijangkau")
	}

	return helper.Success(c, fiber.StatusOK, "mahasiswa berhasil diperbarui sebagian", toModel(hasil))
}

func (s *StudentService) Delete(c *fiber.Ctx) error {
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	err := s.repo.Delete(c.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return helper.Fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}
	if err != nil {
		return helper.Fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dijangkau")
	}
	return helper.NoContent(c)
}