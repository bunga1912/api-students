package service

import (
	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type NilaiService struct {
	repo repository.NilaiRepository
}

func NewNilaiService(repo repository.NilaiRepository) *NilaiService {
	return &NilaiService{repo: repo}
}

func (s *NilaiService) GetByNIM(c *fiber.Ctx) error {
	nim := c.Params("nim")
	if nim == "" {
		return helper.Fail(c, fiber.StatusBadRequest, "nim wajib diisi")
	}

	data, err := s.repo.FindByNIM(c.Context(), nim)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data nilai")
	}

	if len(data) == 0 {
		return helper.Fail(c, fiber.StatusNotFound, "data nilai tidak ditemukan untuk NIM ini")
	}

	return helper.Success(c, fiber.StatusOK, "daftar nilai berhasil diambil", model.WebResponse{
		Data: data,
	})
}