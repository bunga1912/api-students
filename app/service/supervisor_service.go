package service

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

// errSameSupervisor adalah aturan bisnis layer service: pembimbing tujuan
// tidak boleh sama dengan pembimbing saat ini.
var errSameSupervisor = errors.New("pembimbing tujuan sama dengan pembimbing saat ini")

type SupervisorService struct {
	supervisors repository.SupervisorRepository
	tx          *repository.TxManager
}

func NewSupervisorService(supervisors repository.SupervisorRepository, tx *repository.TxManager) *SupervisorService {
	return &SupervisorService{supervisors: supervisors, tx: tx}
}

// ---------- GET /supervisors ----------
func (s *SupervisorService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	list, err := s.supervisors.FindAll(ctx)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "daftar pembimbing berhasil diambil", list)
}

// ---------- GET /supervisors/:id ----------
func (s *SupervisorService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	supervisor, err := s.supervisors.FindByID(ctx, id)
	if err != nil {
		return translateSupervisorError(err)
	}
	return helper.Success(c, fiber.StatusOK, "pembimbing ditemukan", supervisor)
}

// ---------- POST /supervisors/:id/assign ----------
// Dijaga middleware dengan permission supervision:manage.
func (s *SupervisorService) Assign(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	supervisorID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.AssignSupervisorRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	var assigned model.Supervisor

	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		// Seluruh query di bawah ini WAJIB memakai repo, bukan s.supervisors
		// (yang terikat pool dan berjalan di luar transaction).
		repo := s.supervisors.WithTx(tx)

		// 1. Kunci baris mahasiswa. Dua request serentak untuk mahasiswa
		//    yang SAMA harus antre, sehingga satu mahasiswa tidak mungkin
		//    ditetapkan ke dua pembimbing sekaligus.
		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}

		// 2. Ambil satu slot secara atomik (UPDATE bersyarat pada kuota).
		var err error
		assigned, err = repo.IncrementSupervised(ctx, supervisorID)
		if err != nil {
			return err
		}

		// 3. Catat penetapannya. PRIMARY KEY (student_id) menolak mahasiswa
		//    yang sudah punya pembimbing; error di sini membatalkan juga
		//    penambahan slot pada langkah 2 lewat rollback.
		return repo.InsertSupervision(ctx, req.StudentID, supervisorID)
	})
	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Created(c, "pembimbing berhasil ditetapkan", assigned,
		"/api/v1/supervisors/"+strconv.Itoa(supervisorID))
}

// ---------- POST /supervisions/switch ----------
// Dijaga middleware dengan permission supervision:manage.
//
// Mengganti pembimbing harus ATOMIK: bila pembimbing baru ternyata penuh,
// mahasiswa TETAP bersama pembimbing lamanya.
func (s *SupervisorService) Switch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.SwitchSupervisionRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	var joined model.Supervisor

	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.supervisors.WithTx(tx)

		// 1. Kunci baris mahasiswa SEBELUM membaca pembimbing lamanya.
		//    Tanpa kunci ini, dua switch serentak untuk mahasiswa yang
		//    sama sama-sama membaca pembimbing lama X, sama-sama
		//    mengurangi X, dan sama-sama menambah tujuannya masing-masing.
		//    Penghitung pembimbing menjadi salah walau constraint tidak
		//    terpicu.
		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}

		from, err := repo.FindSupervisorOf(ctx, req.StudentID)
		if err != nil {
			return err
		}
		if from == req.ToSupervisorID {
			return errSameSupervisor
		}

		// 2. Kunci KEDUA pembimbing dalam urutan id menaik. Tanpa ini,
		//    mahasiswa A yang pindah 1 -> 2 dan mahasiswa B yang pindah
		//    2 -> 1 pada saat yang sama akan saling menunggu: deadlock.
		if err := repo.LockSupervisors(ctx, from, req.ToSupervisorID); err != nil {
			return err
		}

		// 3. Lepas slot lama, ambil slot baru. Bila pembimbing baru penuh,
		//    error di sini membatalkan juga pengurangan pada pembimbing lama.
		if _, err := repo.DecrementSupervised(ctx, from); err != nil {
			return err
		}
		joined, err = repo.IncrementSupervised(ctx, req.ToSupervisorID)
		if err != nil {
			return err
		}

		return repo.MoveSupervision(ctx, req.StudentID, req.ToSupervisorID)
	})
	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Success(c, fiber.StatusOK, "pembimbing berhasil diganti", joined)
}

// ---------- PATCH /supervisors/:id ----------
// Dijaga middleware dengan permission supervision:manage.
func (s *SupervisorService) UpdateQuota(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.UpdateSupervisorQuotaRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	// Satu pernyataan UPDATE saja, sehingga tidak perlu WithTx.
	updated, err := s.supervisors.UpdateQuota(ctx, id, req.Quota, req.Version)
	if err != nil {
		return translateSupervisorError(err)
	}
	return helper.Success(c, fiber.StatusOK, "kuota pembimbing berhasil diubah", updated)
}

// translateSupervisorError memetakan kegagalan ke response. Kegagalan aturan
// bisnis memakai 409 dengan kode berbeda; yang tidak dikenali menjadi 500,
// termasuk pelanggaran CHECK constraint supervisors_count_within_quota.
func translateSupervisorError(err error) error {
	switch {
	case errors.Is(err, repository.ErrSupervisorNotFound):
		return helper.NotFound("pembimbing tidak ditemukan")
	case errors.Is(err, repository.ErrStudentNotFound):
		return helper.NotFound("mahasiswa tidak ditemukan")
	case errors.Is(err, repository.ErrSupervisorFull):
		return helper.NewError(fiber.StatusConflict, "SUPERVISOR_FULL", "kuota pembimbing sudah penuh")
	case errors.Is(err, repository.ErrAlreadySupervised):
		return helper.NewError(fiber.StatusConflict, "ALREADY_SUPERVISED", "mahasiswa sudah memiliki pembimbing")
	case errors.Is(err, repository.ErrNotSupervised):
		return helper.NewError(fiber.StatusConflict, "NOT_SUPERVISED", "mahasiswa belum memiliki pembimbing")
	case errors.Is(err, errSameSupervisor):
		return helper.Validation(map[string]string{
			"to_supervisor_id": "tidak boleh sama dengan pembimbing saat ini",
		})
	case errors.Is(err, repository.ErrVersionConflict):
		return helper.NewError(fiber.StatusConflict, "VERSION_CONFLICT",
			"data pembimbing sudah diubah pihak lain, muat ulang lalu coba lagi")
	case errors.Is(err, repository.ErrSupervisorQuotaBelowCount):
		return helper.NewError(fiber.StatusConflict, "QUOTA_BELOW_SUPERVISED",
			"kuota tidak boleh lebih kecil dari jumlah mahasiswa bimbingan saat ini")
	default:
		return helper.Internal(err)
	}
}