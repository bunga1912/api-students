package main

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/repository"
)

// repo diisi dari main() setelah koneksi database berhasil dibuat.
// Package-level variable ini yang jadi jembatan antara handler
// dan lapisan repository.
var repo repository.StudentRepository

func paramID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

// toModel mengubah representasi repository menjadi representasi
// yang dikirim sebagai JSON ke klien.
func toModel(s repository.Student) Student {
	return Student{
		ID:       s.ID,
		NIM:      s.NIM,
		Name:     s.Name,
		Grade:    s.Grade,
		IsActive: s.IsActive,
	}
}

func listStudents(c *fiber.Ctx) error {
	q := parseListQuery(c)

	filter := repository.StudentListFilter{
		Page:     q.Page,
		Limit:    q.Limit,
		Search:   q.Search,
		Sort:     q.Sort,
		Order:    q.Order,
		IsActive: q.IsActive,
	}

	hasil, total, err := repo.FindAll(c.Context(), filter)
	if err != nil {
		return fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dijangkau")
	}

	data := make([]Student, 0, len(hasil))
	for _, s := range hasil {
		data = append(data, toModel(s))
	}

	totalPages := (total + q.Limit - 1) / q.Limit
	return okList(c, "daftar mahasiswa berhasil diambil", data, &Meta{
		Page: q.Page, Limit: q.Limit, Total: total, TotalPages: totalPages,
	})
}

func getStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	s, err := repo.FindByID(c.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}
	if err != nil {
		return fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dijangkau")
	}

	return ok(c, "mahasiswa ditemukan", toModel(s))
}

func createStudent(c *fiber.Ctx) error {
	var req CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	errs := map[string]string{}
	if req.NIM == "" {
		errs["nim"] = "wajib diisi"
	}
	if req.Name == "" {
		errs["name"] = "wajib diisi"
	}
	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "harus antara 0 dan 100"
	}
	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	baru, err := repo.Create(c.Context(), repository.Student{
		NIM:   req.NIM,
		Name:  req.Name,
		Grade: req.Grade,
	})
	if errors.Is(err, repository.ErrDuplicate) {
		return fail(c, fiber.StatusConflict, "NIM sudah terdaftar")
	}
	if err != nil {
		return fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dijangkau")
	}

	return created(c, "mahasiswa berhasil ditambahkan", toModel(baru),
		"/api/v1/students/"+strconv.Itoa(baru.ID))
}

func replaceStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	errs := map[string]string{}
	if req.NIM == "" {
		errs["nim"] = "wajib diisi pada PUT"
	}
	if req.Name == "" {
		errs["name"] = "wajib diisi pada PUT"
	}
	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "harus antara 0 dan 100"
	}
	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	hasil, err := repo.Update(c.Context(), repository.Student{
		ID:       id,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	})
	if errors.Is(err, repository.ErrNotFound) {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}
	if errors.Is(err, repository.ErrDuplicate) {
		return fail(c, fiber.StatusConflict, "NIM sudah dipakai mahasiswa lain")
	}
	if err != nil {
		return fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dijangkau")
	}

	return ok(c, "mahasiswa berhasil diganti seluruhnya", toModel(hasil))
}

func patchStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil {
		return fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	// PATCH tetap butuh data lengkap saat ini sebelum ditimpa sebagian,
	// karena query UPDATE pada repository ini mengganti seluruh baris.
	existing, err := repo.FindByID(c.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}
	if err != nil {
		return fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dijangkau")
	}

	if req.NIM != nil {
		nimBaru := strings.TrimSpace(*req.NIM)
		if nimBaru == "" {
			return failValidation(c, map[string]string{"nim": "tidak boleh kosong"})
		}
		existing.NIM = nimBaru
	}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return failValidation(c, map[string]string{"name": "tidak boleh kosong"})
		}
		existing.Name = *req.Name
	}
	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			return failValidation(c, map[string]string{"grade": "harus antara 0 dan 100"})
		}
		existing.Grade = *req.Grade
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	hasil, err := repo.Update(c.Context(), existing)
	if errors.Is(err, repository.ErrDuplicate) {
		return fail(c, fiber.StatusConflict, "NIM sudah dipakai mahasiswa lain")
	}
	if err != nil {
		return fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dijangkau")
	}

	return ok(c, "mahasiswa berhasil diperbarui sebagian", toModel(hasil))
}

func deleteStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	err := repo.Delete(c.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}
	if err != nil {
		return fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dijangkau")
	}

	return noContent(c)
}