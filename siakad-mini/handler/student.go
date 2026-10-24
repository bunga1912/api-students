package handler

import (
	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/app/service"
)

type createStudentRequest struct {
	NIM         string   `json:"nim"`
	Nama        string   `json:"nama"`
	Email       string   `json:"email"`
	Prodi       string   `json:"prodi"`
	Angkatan    int      `json:"angkatan"`
	IPKTerakhir *float64 `json:"ipk_terakhir"`
}

// NIM tidak ada di sini: NIM tidak boleh diubah.
type updateStudentRequest struct {
	Nama        string   `json:"nama"`
	Prodi       string   `json:"prodi"`
	Angkatan    int      `json:"angkatan"`
	IPKTerakhir *float64 `json:"ipk_terakhir"`
}

type studentDetailResponse struct {
	model.Student
	MataKuliah []model.EnrollmentDetail `json:"mata_kuliah"`
	TotalSKS   int                      `json:"total_sks"`
	BatasSKS   int                      `json:"batas_sks"`
}

type StudentHandler struct {
	svc *service.StudentService
}

func NewStudentHandler(svc *service.StudentService) *StudentHandler {
	return &StudentHandler{svc: svc}
}

// GET /api/v1/students?page=&per_page=&prodi=&angkatan=&search=&sort=
func (h *StudentHandler) List(c *fiber.Ctx) error {
	verr := model.NewValidationError()

	filter := model.StudentFilter{
		Prodi:    c.Query("prodi"),
		Search:   c.Query("search"),
		Sort:     c.Query("sort"),
		Angkatan: queryInt(c, "angkatan", 0, verr),
		Page:     queryInt(c, "page", 1, verr),
		PerPage:  queryInt(c, "per_page", 10, verr),
	}

	if verr.HasErrors() {
		return respondError(c, verr)
	}

	page, err := h.svc.List(c.UserContext(), filter)
	if err != nil {
		return respondError(c, err)
	}

	return respondPaginated(
		c,
		"Data mahasiswa berhasil diambil",
		page.Items,
		pageMeta{
			CurrentPage: page.Page,
			PerPage:     page.PerPage,
			Total:       page.Total,
			LastPage:    page.LastPage,
		},
	)
}

// GET /api/v1/students/:id[?tahun_akademik=2026/2027-Ganjil]
// Admin boleh melihat semua, mahasiswa hanya data dirinya sendiri.
func (h *StudentHandler) GetByID(c *fiber.Ctx) error {
	id, ok := parseID(c, "id")
	if !ok {
		return respondError(c, model.ErrNotFound)
	}

	if !isAdmin(c) {
		own, err := h.svc.GetByUserID(c.UserContext(), currentUserID(c))
		if err != nil {
			return respondError(c, err)
		}

		if own.ID != id {
			return respondError(c, model.ErrForbidden)
		}
	}

	detail, err := h.svc.GetDetail(c.UserContext(), id, c.Query("tahun_akademik"))
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, "Detail mahasiswa berhasil diambil", studentDetailResponse{
		Student:    *detail.Student,
		MataKuliah: detail.Enrollments,
		TotalSKS:   detail.TotalSKS,
		BatasSKS:   detail.BatasSKS,
	})
}

// POST /api/v1/students
func (h *StudentHandler) Create(c *fiber.Ctx) error {
	var req createStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return invalidRequest(c, "Format request tidak valid")
	}

	created, err := h.svc.Create(c.UserContext(), service.CreateStudentInput{
		NIM:         req.NIM,
		Nama:        req.Nama,
		Email:       req.Email,
		Prodi:       req.Prodi,
		Angkatan:    req.Angkatan,
		IPKTerakhir: req.IPKTerakhir,
	})
	if err != nil {
		return respondError(c, err)
	}

	return respondCreated(c, "Mahasiswa berhasil ditambahkan", created)
}

// PUT /api/v1/students/:id
func (h *StudentHandler) Update(c *fiber.Ctx) error {
	id, ok := parseID(c, "id")
	if !ok {
		return respondError(c, model.ErrNotFound)
	}

	var req updateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return invalidRequest(c, "Format request tidak valid")
	}

	updated, err := h.svc.Update(c.UserContext(), id, service.UpdateStudentInput{
		Nama:        req.Nama,
		Prodi:       req.Prodi,
		Angkatan:    req.Angkatan,
		IPKTerakhir: req.IPKTerakhir,
	})
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, "Data mahasiswa berhasil diperbarui", updated)
}

// DELETE /api/v1/students/:id  (soft delete)
func (h *StudentHandler) Delete(c *fiber.Ctx) error {
	id, ok := parseID(c, "id")
	if !ok {
		return respondError(c, model.ErrNotFound)
	}

	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		return respondError(c, err)
	}

	return respondNoContent(c)
}