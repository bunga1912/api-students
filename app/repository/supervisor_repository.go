package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"api-students/app/model"
)

var (
	ErrSupervisorNotFound        = errors.New("pembimbing tidak ditemukan")
	ErrStudentNotFound           = errors.New("mahasiswa tidak ditemukan")
	ErrSupervisorFull            = errors.New("kuota pembimbing sudah penuh")
	ErrAlreadySupervised         = errors.New("mahasiswa sudah memiliki pembimbing")
	ErrNotSupervised             = errors.New("mahasiswa belum memiliki pembimbing")
	ErrSupervisorQuotaBelowCount = errors.New("kuota lebih kecil dari jumlah mahasiswa bimbingan")
)

type SupervisorRepository interface {
	// WithTx mengembalikan salinan repository yang menjalankan seluruh
	// query-nya di dalam transaction tx.
	WithTx(tx DBTX) SupervisorRepository

	FindAll(ctx context.Context) ([]model.Supervisor, error)
	FindByID(ctx context.Context, id int) (model.Supervisor, error)

	LockStudent(ctx context.Context, studentID int) error
	FindSupervisorOf(ctx context.Context, studentID int) (int, error)
	LockSupervisors(ctx context.Context, ids ...int) error

	IncrementSupervised(ctx context.Context, supervisorID int) (model.Supervisor, error)
	DecrementSupervised(ctx context.Context, supervisorID int) (model.Supervisor, error)
	InsertSupervision(ctx context.Context, studentID, supervisorID int) error
	MoveSupervision(ctx context.Context, studentID, toSupervisorID int) error

	UpdateQuota(ctx context.Context, id, quota, version int) (model.Supervisor, error)
}

const supervisorColumns = `id, name, quota, supervised_count, version, created_at`

type supervisorPostgresRepository struct{ db DBTX }

func NewSupervisorRepository(db DBTX) SupervisorRepository {
	return &supervisorPostgresRepository{db: db}
}

func (r *supervisorPostgresRepository) WithTx(tx DBTX) SupervisorRepository {
	return &supervisorPostgresRepository{db: tx}
}

func scanSupervisor(row pgx.Row) (model.Supervisor, error) {
	var s model.Supervisor
	err := row.Scan(&s.ID, &s.Name, &s.Quota, &s.SupervisedCount, &s.Version, &s.CreatedAt)
	return s, err
}

func (r *supervisorPostgresRepository) FindAll(ctx context.Context) ([]model.Supervisor, error) {
	rows, err := r.db.Query(ctx,
		"SELECT "+supervisorColumns+" FROM supervisors ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar pembimbing: %w", err)
	}
	defer rows.Close()

	result := []model.Supervisor{}
	for rows.Next() {
		s, err := scanSupervisor(rows)
		if err != nil {
			return nil, fmt.Errorf("membaca row pembimbing: %w", err)
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

func (r *supervisorPostgresRepository) FindByID(ctx context.Context, id int) (model.Supervisor, error) {
	s, err := scanSupervisor(r.db.QueryRow(ctx,
		"SELECT "+supervisorColumns+" FROM supervisors WHERE id = $1", id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Supervisor{}, ErrSupervisorNotFound
		}
		return model.Supervisor{}, fmt.Errorf("mengambil pembimbing: %w", err)
	}
	return s, nil
}

// LockStudent mengunci row milik satu mahasiswa sampai transaction selesai.
//
// Row students dipakai sebagai "gembok pembimbing": selama satu transaction
// memegang kunci ini, transaction lain untuk mahasiswa YANG SAMA harus
// antre. Mahasiswa lain tidak terpengaruh, karena row yang dikunci berbeda.
func (r *supervisorPostgresRepository) LockStudent(ctx context.Context, studentID int) error {
	var id int
	err := r.db.QueryRow(ctx,
		"SELECT id FROM students WHERE id = $1 FOR UPDATE", studentID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrStudentNotFound
		}
		return fmt.Errorf("mengunci data mahasiswa: %w", err)
	}
	return nil
}

// FindSupervisorOf membaca pembimbing mahasiswa saat ini. Hanya aman dipakai
// sesudah LockStudent, karena kunci itulah yang menjamin nilainya tidak
// berubah sampai transaction selesai.
func (r *supervisorPostgresRepository) FindSupervisorOf(ctx context.Context, studentID int) (int, error) {
	var supervisorID int
	err := r.db.QueryRow(ctx,
		"SELECT supervisor_id FROM supervisions WHERE student_id = $1", studentID).Scan(&supervisorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotSupervised
		}
		return 0, fmt.Errorf("membaca pembimbing mahasiswa: %w", err)
	}
	return supervisorID, nil
}

// LockSupervisors mengunci beberapa pembimbing sekaligus, SELALU dalam
// urutan id menaik, apa pun urutan argumennya. Urutan yang seragam inilah
// yang mencegah deadlock pada pergantian pembimbing.
func (r *supervisorPostgresRepository) LockSupervisors(ctx context.Context, ids ...int) error {
	rows, err := r.db.Query(ctx,
		"SELECT id FROM supervisors WHERE id = ANY($1) ORDER BY id FOR UPDATE", ids)
	if err != nil {
		return fmt.Errorf("mengunci pembimbing: %w", err)
	}
	defer rows.Close()

	found := 0
	for rows.Next() {
		found++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("mengunci pembimbing: %w", err)
	}
	if found != len(ids) {
		return ErrSupervisorNotFound
	}
	return nil
}

// IncrementSupervised mengambil satu slot bimbingan secara ATOMIK: pemeriksaan
// kuota dan penambahan jumlah terjadi dalam satu pernyataan UPDATE.
func (r *supervisorPostgresRepository) IncrementSupervised(ctx context.Context, supervisorID int) (model.Supervisor, error) {
	s, err := scanSupervisor(r.db.QueryRow(ctx,
		`UPDATE supervisors SET supervised_count = supervised_count + 1
		 WHERE id = $1 AND supervised_count < quota
		 RETURNING `+supervisorColumns, supervisorID))
	if err == nil {
		return s, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return model.Supervisor{}, fmt.Errorf("menambah mahasiswa bimbingan: %w", err)
	}

	// Tidak ada row yang berubah: pembimbingnya tidak ada, atau sudah penuh.
	if _, findErr := r.FindByID(ctx, supervisorID); findErr != nil {
		return model.Supervisor{}, findErr
	}
	return model.Supervisor{}, ErrSupervisorFull
}

func (r *supervisorPostgresRepository) DecrementSupervised(ctx context.Context, supervisorID int) (model.Supervisor, error) {
	s, err := scanSupervisor(r.db.QueryRow(ctx,
		`UPDATE supervisors SET supervised_count = supervised_count - 1
		 WHERE id = $1
		 RETURNING `+supervisorColumns, supervisorID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Supervisor{}, ErrSupervisorNotFound
		}
		return model.Supervisor{}, fmt.Errorf("mengurangi mahasiswa bimbingan: %w", err)
	}
	return s, nil
}

func (r *supervisorPostgresRepository) InsertSupervision(ctx context.Context, studentID, supervisorID int) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO supervisions (student_id, supervisor_id) VALUES ($1, $2)", studentID, supervisorID)
	if err != nil {
		if pgErrorCode(err) == codeUniqueViolation {
			return ErrAlreadySupervised
		}
		return fmt.Errorf("mencatat pembimbing: %w", err)
	}
	return nil
}

// MoveSupervision memindahkan mahasiswa ke pembimbing lain, lalu MEMERIKSA
// apakah memang ada row yang berubah (UPDATE dengan nol row bukan error
// menurut PostgreSQL).
func (r *supervisorPostgresRepository) MoveSupervision(ctx context.Context, studentID, toSupervisorID int) error {
	tag, err := r.db.Exec(ctx,
		"UPDATE supervisions SET supervisor_id = $2 WHERE student_id = $1", studentID, toSupervisorID)
	if err != nil {
		return fmt.Errorf("memindahkan pembimbing: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotSupervised
	}
	return nil
}

// UpdateQuota menerapkan optimistic locking, sama seperti pada courses.
func (r *supervisorPostgresRepository) UpdateQuota(ctx context.Context, id, quota, version int) (model.Supervisor, error) {
	s, err := scanSupervisor(r.db.QueryRow(ctx,
		`UPDATE supervisors SET quota = $1, version = version + 1
		 WHERE id = $2 AND version = $3 AND $1 >= supervised_count
		 RETURNING `+supervisorColumns, quota, id, version))
	if err == nil {
		return s, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return model.Supervisor{}, fmt.Errorf("mengubah kuota pembimbing: %w", err)
	}

	current, findErr := r.FindByID(ctx, id)
	if findErr != nil {
		return model.Supervisor{}, findErr
	}
	if current.Version != version {
		return model.Supervisor{}, ErrVersionConflict
	}
	return model.Supervisor{}, ErrSupervisorQuotaBelowCount
}