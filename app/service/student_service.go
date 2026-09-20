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

type StudentService struct {
	repo  repository.StudentRepository
	perms *helper.PermissionSet
}

func NewStudentService(repo repository.StudentRepository, perms *helper.PermissionSet) *StudentService {
	return &StudentService{repo: repo, perms: perms}
}

func toModel(s repository.Student) model.Student {
	return model.Student{
		ID: s.ID, NIM: s.NIM, Name: s.Name, Grade: s.Grade,
		IsActive: s.IsActive, OwnerID: s.OwnerID,
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

// Get memeriksa kepemilikan SETELAH data diambil, karena baru pada titik
// itu owner_id-nya diketahui. Ini sebabnya pemeriksaan seperti ini tidak
// bisa dipindah ke middleware: middleware belum pernah menyentuh database.
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

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	hasAny := s.perms.Can(authUser.Role, "student:read:any")
	if !CanAccessStudent(authUser.UserID, item.OwnerID, hasAny) {
		return helper.Fail(c, fiber.StatusForbidden,
			"anda tidak memiliki hak untuk mengakses data ini")
	}

	return helper.Success(c, fiber.StatusOK, "mahasiswa ditemukan", toModel(item))
}

// Create mengisi owner_id dari identitas pemanggil, BUKAN dari body
// request. Bila klien mengirim "owner_id" di body, nilainya diabaikan
// begitu saja karena CreateStudentRequest tidak memiliki field itu —
// pola yang sama seperti pencegahan mass assignment pada role di Modul 5.
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

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}
	ownerID := authUser.UserID

	baru, err := s.repo.Create(c.Context(), repository.Student{
		NIM: req.NIM, Name: req.Name, Grade: req.Grade, OwnerID: &ownerID,
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

	existing, err := s.repo.FindByID(c.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return helper.Fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}
	if err != nil {
		return helper.Fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dijangkau")
	}

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	hasAny := s.perms.Can(authUser.Role, "student:update:any")
	if !CanAccessStudent(authUser.UserID, existing.OwnerID, hasAny) {
		return helper.Fail(c, fiber.StatusForbidden,
			"anda tidak memiliki hak untuk mengubah data ini")
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

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	hasAny := s.perms.Can(authUser.Role, "student:update:any")
	if !CanAccessStudent(authUser.UserID, existingRepo.OwnerID, hasAny) {
		return helper.Fail(c, fiber.StatusForbidden,
			"anda tidak memiliki hak untuk mengubah data ini")
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

// Delete TIDAK memeriksa kepemilikan — haknya sudah diputuskan di
// middleware lewat permission "student:delete" (lihat route.go),
// karena penghapusan dianggap tindakan yang cukup berisiko sehingga
// hanya boleh dilakukan oleh role yang eksplisit diberi izin, bukan
// oleh pemilik data biasa.
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