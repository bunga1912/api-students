package service

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

// PortfolioService memakai DUA database: keberadaan mahasiswa diperiksa ke
// PostgreSQL, portofolionya disimpan di MongoDB. Tidak ada FOREIGN KEY lintas
// database, jadi service-lah yang menjaga hubungan itu.
type PortfolioService struct {
	portfolios repository.PortfolioRepository
	students   repository.StudentLookup
}

func NewPortfolioService(
	portfolios repository.PortfolioRepository,
	students repository.StudentLookup,
) *PortfolioService {
	return &PortfolioService{portfolios: portfolios, students: students}
}

// requireStudent menolak bila mahasiswanya tidak ada di PostgreSQL.
func (s *PortfolioService) requireStudent(ctx context.Context, studentID int) error {
	exists, err := s.students.Exists(ctx, studentID)
	if err != nil {
		return helper.Internal(err)
	}
	if !exists {
		return helper.NewError(fiber.StatusNotFound, "STUDENT_NOT_FOUND", "mahasiswa tidak ditemukan")
	}
	return nil
}

// normalizeProject merapikan masukan proyek sebelum divalidasi: teks tanpa
// spasi di tepi, teknologi huruf kecil tanpa duplikat.
func normalizeProject(p *model.ProjectInput) {
	p.Title = strings.TrimSpace(p.Title)
	p.URL = strings.TrimSpace(p.URL)
	p.Technologies = NormalizeTags(p.Technologies)
}

func projectFromInput(in model.ProjectInput) model.Project {
	return model.Project{Title: in.Title, Technologies: in.Technologies, URL: in.URL, Year: in.Year}
}

// ---------- GET /students/:id/portfolio ----------

func (s *PortfolioService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}
	if err := s.requireStudent(ctx, studentID); err != nil {
		return err
	}

	portfolio, err := s.portfolios.FindByStudent(ctx, studentID)
	if err != nil {
		return translatePortfolioError(err)
	}
	return helper.Success(c, fiber.StatusOK, "portofolio ditemukan", portfolio)
}

// ---------- PUT /students/:id/portfolio ----------
// Membuat portofolio bila belum ada, atau mengganti SELURUH isinya bila sudah.
// Proyek yang tidak ikut dikirim akan hilang, dan setiap proyek mendapat id baru.

func (s *PortfolioService) Put(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	var req model.PortfolioRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	req.Headline = strings.TrimSpace(req.Headline)
	req.Bio = strings.TrimSpace(req.Bio)
	for i := range req.Projects {
		normalizeProject(&req.Projects[i])
	}
	for i := range req.Certificates {
		req.Certificates[i].Name = strings.TrimSpace(req.Certificates[i].Name)
		req.Certificates[i].Issuer = strings.TrimSpace(req.Certificates[i].Issuer)
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	if err := s.requireStudent(ctx, studentID); err != nil {
		return err
	}

	portfolio := model.Portfolio{
		StudentID: studentID, Headline: req.Headline, Bio: req.Bio,
		Projects:     make([]model.Project, 0, len(req.Projects)),
		Certificates: make([]model.Certificate, 0, len(req.Certificates)),
	}
	for _, p := range req.Projects {
		portfolio.Projects = append(portfolio.Projects, projectFromInput(p))
	}
	for _, cert := range req.Certificates {
		portfolio.Certificates = append(portfolio.Certificates,
			model.Certificate{Name: cert.Name, Issuer: cert.Issuer, Date: cert.Date})
	}

	saved, created, err := s.portfolios.Upsert(ctx, studentID, portfolio)
	if err != nil {
		return translatePortfolioError(err)
	}
	if created {
		return helper.Created(c, "portofolio berhasil dibuat", saved,
			"/api/v1/students/"+strconv.Itoa(studentID)+"/portfolio")
	}
	return helper.Success(c, fiber.StatusOK, "portofolio berhasil diganti", saved)
}

// ---------- POST /students/:id/portfolio/projects ----------

func (s *PortfolioService) AddProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	var req model.ProjectInput
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	normalizeProject(&req)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	if err := s.requireStudent(ctx, studentID); err != nil {
		return err
	}

	updated, err := s.portfolios.AddProject(ctx, studentID, projectFromInput(req))
	if err != nil {
		return translatePortfolioError(err)
	}
	return helper.Created(c, "proyek berhasil ditambahkan", updated,
		"/api/v1/students/"+strconv.Itoa(studentID)+"/portfolio")
}

// ---------- DELETE /students/:id/portfolio/projects/:projectId ----------

func (s *PortfolioService) RemoveProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}
	if err := s.requireStudent(ctx, studentID); err != nil {
		return err
	}

	if err := s.portfolios.RemoveProject(ctx, studentID, c.Params("projectId")); err != nil {
		return translatePortfolioError(err)
	}
	return helper.NoContent(c)
}

// ---------- GET /portfolios/stats/technologies ----------

func (s *PortfolioService) TechnologyStats(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	stats, err := s.portfolios.TopTechnologies(ctx, 10)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "statistik teknologi berhasil diambil", stats)
}

func translatePortfolioError(err error) error {
	switch {
	case errors.Is(err, repository.ErrInvalidID):
		return helper.BadRequest("id proyek tidak valid")
	case errors.Is(err, repository.ErrPortfolioNotFound):
		return helper.NewError(fiber.StatusNotFound, "PORTFOLIO_NOT_FOUND",
			"portofolio belum dibuat, buat dengan PUT terlebih dahulu")
	case errors.Is(err, repository.ErrProjectNotFound):
		return helper.NewError(fiber.StatusNotFound, "PROJECT_NOT_FOUND", "proyek tidak ditemukan")
	case errors.Is(err, repository.ErrProjectLimit):
		return helper.NewError(fiber.StatusConflict, "PROJECT_LIMIT",
			"jumlah proyek sudah mencapai batas "+strconv.Itoa(repository.MaxProjectsPerPortfolio))
	default:
		return helper.Internal(err)
	}
}