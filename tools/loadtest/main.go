// Alat uji beban untuk Modul 9. BUKAN bagian aplikasi, tidak ikut di-deploy.
//
// Pemakaian (server harus sedang berjalan di terminal lain):
//
//	$env:DATABASE_URL = "postgres://postgres:PASSWORD@localhost:5432/praktikum_backend"
//	$env:JWT_SECRET   = "isi-sama-dengan-.env-server"
//	$env:JWT_ISSUER   = "isi-sama-dengan-issuer-di-main.go/.env"
//	go run ./tools/loadtest -mode crowd
//	go run ./tools/loadtest -mode credits
//	go run ./tools/loadtest -mode swap
//
// Jalankan reset.sql sebelum SETIAP uji.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
	"api-students/helper"
)

type student struct {
	ID       int
	Username string
	Role     string
	Token    string
}

// fire menjalankan n pekerjaan secara serentak dan menghitung hasilnya.
//
// Setiap goroutine menunggu di depan "gerbang" start yang sama. Gerbang
// baru dibuka setelah SELURUH goroutine siap.
func fire(n int, job func(i int) string) map[string]int {
	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make(chan string, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int, out chan<- string) {
			defer wg.Done()
			<-start
			out <- job(i)
		}(i, results)
	}

	close(start) // buka gerbang
	wg.Wait()    // tunggu SEMUA goroutine selesai
	close(results)

	tally := map[string]int{}
	for r := range results {
		tally[r]++
	}
	return tally
}

func main() {
	mode := flag.String("mode", "crowd", "crowd | credits | swap | token")
	userFlag := flag.String("user", "mhs001", "mode token: username akun uji (mhs001..mhs060)")
	roleFlag := flag.String("role", "", "mode token: ganti role di token (mis. admin); kosong = role asli")
	base := flag.String("base", "http://localhost:3000", "alamat server")
	flag.Parse()

	dbURL := os.Getenv("DATABASE_URL")
	secret := os.Getenv("JWT_SECRET")
	issuer := os.Getenv("JWT_ISSUER")
	if dbURL == "" || secret == "" || issuer == "" {
		fmt.Println("set dulu env DATABASE_URL, JWT_SECRET, dan JWT_ISSUER")
		os.Exit(1)
	}
	jwtManager := helper.NewJWTManager(secret, issuer, time.Hour)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Println("koneksi database:", err)
		os.Exit(1)
	}
	defer pool.Close()

	students, err := loadStudents(ctx, pool, jwtManager)
	if err != nil {
		fmt.Println("memuat akun mhs:", err)
		os.Exit(1)
	}
	if len(students) < 60 {
		fmt.Printf("hanya ada %d akun mhs, jalankan migration 006\n", len(students))
		os.Exit(1)
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        200,
			MaxIdleConnsPerHost: 200,
		},
	}
	api := *base + "/api/v1"

	switch *mode {
	case "crowd":
		runCrowd(ctx, pool, client, api, students)
	case "credits":
		runCredits(ctx, pool, client, api, students)
	case "swap":
		runSwap(ctx, pool, client, api, students)
	case "token":
		// Mencetak SATU token untuk dipakai di curl (uji bug 7 dan 9).
		for _, st := range students {
			if st.Username != *userFlag {
				continue
			}
			role := st.Role
			if *roleFlag != "" {
				role = *roleFlag
			}
			tok, err := jwtManager.GenerateAccess(model.User{ID: st.ID, Username: st.Username, Role: role})
			if err != nil {
				fmt.Println("gagal membuat token:", err)
				os.Exit(1)
			}
			fmt.Print(tok)
			return
		}
		fmt.Println("akun tidak ditemukan:", *userFlag)
		os.Exit(1)
	default:
		fmt.Println("mode tidak dikenal:", *mode)
		os.Exit(1)
	}
}

// ---------------------------------------------------------------- modes

func runCrowd(ctx context.Context, pool *pgxpool.Pool, c *http.Client, api string, st []student) {
	ids := courseIDs(ctx, pool, "IF399")
	fmt.Println("60 mahasiswa berebut kelas IF399 secara serentak...")

	begin := time.Now()
	tally := fire(60, func(i int) string {
		return call(c, st[i].Token, http.MethodPost,
			fmt.Sprintf("%s/courses/%d/enroll", api, ids["IF399"]), nil)
	})
	elapsed := time.Since(begin)

	printTally(tally, elapsed)
	checkConsistency(ctx, pool, "IF399")
}

func runCredits(ctx context.Context, pool *pgxpool.Pool, c *http.Client, api string, st []student) {
	codes := []string{"IF101", "IF102", "IF301", "IF302", "IF303", "IF304",
		"IF305", "IF306", "IF307", "IF308", "IF309", "IF310"}
	ids := courseIDs(ctx, pool, codes...)
	const attackers = 10
	fmt.Println("10 mahasiswa, masing-masing mendaftar ke 12 kelas 3 SKS secara serentak (batas 24 SKS)...")

	begin := time.Now()
	tally := fire(attackers*len(codes), func(i int) string {
		s := st[i/len(codes)]
		code := codes[i%len(codes)]
		return call(c, s.Token, http.MethodPost,
			fmt.Sprintf("%s/courses/%d/enroll", api, ids[code]), nil)
	})
	elapsed := time.Since(begin)

	printTally(tally, elapsed)

	rows, err := pool.Query(ctx, `
		SELECT u.username, COALESCE(SUM(c.credits), 0)
		FROM users u
		LEFT JOIN enrollments e ON e.user_id = u.id
		LEFT JOIN courses c ON c.id = e.course_id
		WHERE u.id = ANY($1)
		GROUP BY u.username ORDER BY u.username`, studentIDs(st[:attackers]))
	if err != nil {
		fmt.Println("kueri SKS:", err)
		return
	}
	defer rows.Close()
	ok := true
	for rows.Next() {
		var name string
		var total int
		_ = rows.Scan(&name, &total)
		if total > 24 {
			ok = false
			fmt.Printf("  %s = %d SKS  <- MELEBIHI 24\n", name, total)
		}
	}
	if ok {
		fmt.Println("seluruh 10 mahasiswa <= 24 SKS -> KONSISTEN")
	} else {
		fmt.Println("ada mahasiswa > 24 SKS -> TIDAK KONSISTEN")
	}
	checkConsistency(ctx, pool, codes...)
}

func runSwap(ctx context.Context, pool *pgxpool.Pool, c *http.Client, api string, st []student) {
	ids := courseIDs(ctx, pool, "IF101", "IF102")
	a, b := ids["IF101"], ids["IF102"]
	const pairs = 20

	// Persiapan (tidak diukur): mhs001-020 masuk IF101, mhs021-040 masuk IF102.
	fmt.Println("persiapan: mendaftarkan 40 mahasiswa (bergiliran)...")
	for i := 0; i < pairs; i++ {
		r1 := call(c, st[i].Token, http.MethodPost, fmt.Sprintf("%s/courses/%d/enroll", api, a), nil)
		r2 := call(c, st[pairs+i].Token, http.MethodPost, fmt.Sprintf("%s/courses/%d/enroll", api, b), nil)
		if r1 != "201 OK" || r2 != "201 OK" {
			fmt.Printf("persiapan gagal (%s / %s). Jalankan reset.sql dulu.\n", r1, r2)
			return
		}
	}

	var before int
	_ = pool.QueryRow(ctx,
		`SELECT deadlocks FROM pg_stat_database WHERE datname = current_database()`).Scan(&before)

	fmt.Println("20 pasang mahasiswa pindah kelas berlawanan arah secara serentak...")
	begin := time.Now()
	tally := fire(pairs*2, func(i int) string {
		if i < pairs {
			return swapCall(c, api, st[i].Token, a, b) // IF101 -> IF102
		}
		return swapCall(c, api, st[i].Token, b, a) // IF102 -> IF101
	})
	elapsed := time.Since(begin)

	var after int
	_ = pool.QueryRow(ctx,
		`SELECT deadlocks FROM pg_stat_database WHERE datname = current_database()`).Scan(&after)

	printTally(tally, elapsed)
	fmt.Printf("deadlock terdeteksi PostgreSQL selama uji: %d\n", after-before)
	checkConsistency(ctx, pool, "IF101", "IF102")
}

// ---------------------------------------------------------------- helpers

func swapCall(c *http.Client, api, token string, from, to int) string {
	body, _ := json.Marshal(map[string]int{"from_course_id": from, "to_course_id": to})
	return call(c, token, http.MethodPost, api+"/enrollments/swap", body)
}

// call mengirim satu request dan mengembalikan label hasil, misalnya
// "201 OK" atau "409 COURSE_FULL".
func call(c *http.Client, token, method, url string, body []byte) string {
	var rd io.Reader
	if body != nil {
		rd = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, rd)
	if err != nil {
		return "client-error"
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return "network-error (timeout?)"
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 300 {
		return fmt.Sprintf("%d OK", resp.StatusCode)
	}
	var e struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(raw, &e)
	if e.Code == "" {
		e.Code = "ERROR"
	}
	return fmt.Sprintf("%d %s", resp.StatusCode, e.Code)
}

func printTally(tally map[string]int, elapsed time.Duration) {
	fmt.Println("hasil request:")
	keys := make([]string, 0, len(tally))
	for k := range tally {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  %-32s %d\n", k, tally[k])
	}
	fmt.Printf("waktu total : %dms\n", elapsed.Milliseconds())
}

// checkConsistency membandingkan enrolled_count dengan jumlah baris enrollments.
func checkConsistency(ctx context.Context, pool *pgxpool.Pool, codes ...string) {
	rows, err := pool.Query(ctx, `
		SELECT c.code, c.quota, c.enrolled_count,
		       (SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id)
		FROM courses c WHERE c.code = ANY($1) ORDER BY c.code`, codes)
	if err != nil {
		fmt.Println("kueri konsistensi:", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var code string
		var quota, counter, actual int
		_ = rows.Scan(&code, &quota, &counter, &actual)
		status := "KONSISTEN"
		if counter != actual || actual > quota {
			status = "TIDAK KONSISTEN"
		}
		fmt.Printf("%s kuota=%d enrolled_count=%d baris enrollments=%d -> %s\n",
			code, quota, counter, actual, status)
	}
}

func loadStudents(ctx context.Context, pool *pgxpool.Pool, jwtManager *helper.JWTManager) ([]student, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, username, role FROM users
		WHERE username LIKE 'mhs%' ORDER BY username LIMIT 60`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []student
	for rows.Next() {
		var s student
		if err := rows.Scan(&s.ID, &s.Username, &s.Role); err != nil {
			return nil, err
		}
		s.Token, err = jwtManager.GenerateAccess(model.User{ID: s.ID, Username: s.Username, Role: s.Role})
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func courseIDs(ctx context.Context, pool *pgxpool.Pool, codes ...string) map[string]int {
	out := map[string]int{}
	rows, err := pool.Query(ctx, `SELECT code, id FROM courses WHERE code = ANY($1)`, codes)
	if err != nil {
		fmt.Println("mencari id kelas:", err)
		os.Exit(1)
	}
	defer rows.Close()
	for rows.Next() {
		var code string
		var id int
		_ = rows.Scan(&code, &id)
		out[code] = id
	}
	return out
}

func studentIDs(st []student) []int {
	ids := make([]int, len(st))
	for i, s := range st {
		ids[i] = s.ID
	}
	return ids
}