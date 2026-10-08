package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"siakad-mini/config"
	"siakad-mini/database"
)

const bcryptCost = 12

type seedStudent struct {
	nama     string
	prodi    string
	angkatan int
	ipk      float64
}

type seedCourse struct {
	kode     string
	nama     string
	sks      int
	semester int
	kuota    int
}

var students = []seedStudent{
	{"Rina Putri", "Sistem Informasi", 2022, 3.45},
	{"Budi Santoso", "Teknik Informatika", 2021, 3.12},
	{"Citra Ayu", "Sistem Informasi", 2023, 2.85},
	{"Dimas Pratama", "Teknik Informatika", 2022, 2.40},
	{"Eka Wulandari", "Sistem Informasi", 2021, 3.78},
	{"Fajar Nugroho", "Teknik Informatika", 2023, 2.95},
	{"Gita Lestari", "Sistem Informasi", 2024, 3.30},
	{"Hendra Wijaya", "Teknik Informatika", 2022, 2.15},
	{"Indah Permata", "Sistem Informasi", 2021, 3.60},
	{"Joko Susilo", "Teknik Informatika", 2023, 2.70},
	{"Kartika Sari", "Sistem Informasi", 2022, 3.05},
	{"Lukman Hakim", "Teknik Informatika", 2024, 2.50},
	{"Maya Anggraini", "Sistem Informasi", 2023, 3.90},
	{"Nanda Kurniawan", "Teknik Informatika", 2021, 2.30},
	{"Olivia Rahma", "Sistem Informasi", 2022, 3.20},
	{"Putra Mahendra", "Teknik Informatika", 2023, 2.88},
	{"Qori Aulia", "Sistem Informasi", 2024, 3.50},
	{"Rizky Ramadhan", "Teknik Informatika", 2022, 2.99},
	{"Salsabila Nur", "Sistem Informasi", 2021, 3.00},
	{"Teguh Prasetyo", "Teknik Informatika", 2023, 2.49},
}

var courses = []seedCourse{
	{"TI101", "Algoritma dan Pemrograman", 3, 1, 40},
	{"TI102", "Basis Data", 3, 2, 35},
	{"TI103", "Struktur Data", 3, 2, 35},
	{"TI201", "Pemrograman Web", 3, 3, 30},
	{"TI202", "Jaringan Komputer", 3, 3, 30},
	{"TI203", "Sistem Operasi", 3, 3, 30},
	{"TI301", "Rekayasa Perangkat Lunak", 3, 4, 25},
	{"TI302", "Kecerdasan Buatan", 3, 5, 2},
	{"TI303", "Keamanan Informasi", 2, 5, 2},
	{"TI401", "Pemrograman Backend Lanjut", 4, 6, 20},
}

func hash(plain string) string {
	hashed, err := bcrypt.GenerateFromPassword(
		[]byte(plain),
		bcryptCost,
	)

	if err != nil {
		log.Fatalf(
			"gagal membuat hash password: %v",
			err,
		)
	}

	return string(hashed)
}

func must(err error, pesan string) {
	if err != nil {
		log.Fatalf("%s: %v", pesan, err)
	}
}

func main() {
	err := config.LoadEnv()
	must(err, "memuat environment")

	ctx := context.Background()

	pool, err := database.NewPool(
		ctx,
		database.Config{
			Host:     config.GetEnv("DB_HOST", "localhost"),
			Port:     config.GetEnv("DB_PORT", "5432"),
			User:     config.GetEnv("DB_USER", "postgres"),
			Password: config.GetEnv("DB_PASSWORD", ""),
			Name:     config.GetEnv("DB_NAME", "siakad_mini"),
			SSLMode:  config.GetEnv("DB_SSLMODE", "disable"),
			MaxConns: 10,
		},
	)

	must(err, "database")
	defer pool.Close()

	var jumlah int

	must(
		pool.QueryRow(
			ctx,
			"SELECT COUNT(*) FROM users",
		).Scan(&jumlah),
		"menghitung user",
	)

	if jumlah > 0 {
		log.Println(
			"tabel users sudah berisi data, seeder dilewati",
		)
		return
	}

	tx, err := pool.Begin(ctx)
	must(err, "memulai transaction")

	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		`INSERT INTO users
			(email, password, role)
		 VALUES ($1, $2, 'admin')`,
		"admin@siakad.test",
		hash("Admin12345"),
	)

	must(err, "menyimpan admin")

	for i, s := range students {
		nim := fmt.Sprintf(
			"1872210%05d",
			i+1,
		)

		email := strings.ToLower(
			strings.ReplaceAll(
				s.nama,
				" ",
				".",
			),
		) + "@student.test"

		var userID int64

		err := tx.QueryRow(
			ctx,
			`INSERT INTO users
				(email, password, role)
			 VALUES ($1, $2, 'mahasiswa')
			 RETURNING id`,
			email,
			hash(nim),
		).Scan(&userID)

		must(
			err,
			"menyimpan user "+email,
		)

		_, err = tx.Exec(
			ctx,
			`INSERT INTO students
				(user_id, nim, nama, prodi, angkatan, ipk_terakhir)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			userID,
			nim,
			s.nama,
			s.prodi,
			s.angkatan,
			s.ipk,
		)

		must(
			err,
			"menyimpan mahasiswa "+s.nama,
		)
	}

	for _, c := range courses {
		_, err := tx.Exec(
			ctx,
			`INSERT INTO courses
				(kode_mk, nama_mk, sks, semester, kuota)
			 VALUES ($1, $2, $3, $4, $5)`,
			c.kode,
			c.nama,
			c.sks,
			c.semester,
			c.kuota,
		)

		must(
			err,
			"menyimpan mata kuliah "+c.kode,
		)
	}

	must(
		tx.Commit(ctx),
		"commit transaction",
	)

	log.Printf(
		"seeder selesai: 1 admin, %d mahasiswa, %d mata kuliah",
		len(students),
		len(courses),
	)
}
