
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Result struct {
	StudentID int64
	Status    int
	Duration  time.Duration
	Body      string
	Err       error
}

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	token := os.Getenv("JWT_TOKEN")
	baseURL := os.Getenv("BASE_URL")

	if databaseURL == "" || token == "" {
		fmt.Println("DATABASE_URL dan JWT_TOKEN wajib diisi.")
		os.Exit(1)
	}

	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}

	supervisorID := int64(4)
	if value := os.Getenv("SUPERVISOR_ID"); value != "" {
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil || id <= 0 {
			fmt.Println("SUPERVISOR_ID tidak valid.")
			os.Exit(1)
		}
		supervisorID = id
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		fmt.Println("Gagal konek database:", err)
		os.Exit(1)
	}
	defer pool.Close()

	rows, err := pool.Query(ctx, `
		SELECT s.id
		FROM students s
		LEFT JOIN supervisions su ON su.student_id = s.id
		WHERE s.nim LIKE 'LOADTEST%'
		  AND s.is_active = TRUE
		  AND su.student_id IS NULL
		ORDER BY s.nim
		LIMIT 30
	`)
	if err != nil {
		fmt.Println("Gagal mengambil mahasiswa:", err)
		os.Exit(1)
	}

	var studentIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			fmt.Println("Gagal membaca ID:", err)
			os.Exit(1)
		}
		studentIDs = append(studentIDs, id)
	}

	err = rows.Err()
	rows.Close()
	if err != nil {
		fmt.Println("Gagal membaca hasil query:", err)
		os.Exit(1)
	}

	if len(studentIDs) != 30 {
		fmt.Printf(
			"Mahasiswa yang siap diuji hanya %d. Dibutuhkan 30 mahasiswa aktif yang belum ditugaskan.\n",
			len(studentIDs),
		)
		os.Exit(1)
	}

	endpoint := fmt.Sprintf(
		"%s/api/v1/supervisors/%d/assign",
		strings.TrimRight(baseURL, "/"),
		supervisorID,
	)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	results := make(chan Result, len(studentIDs))
	start := make(chan struct{})
	var wg sync.WaitGroup

	fmt.Printf("Mulai load test: %d mahasiswa, pembimbing ID %d\n",
		len(studentIDs), supervisorID)

	testStart := time.Now()

	for _, studentID := range studentIDs {
		wg.Add(1)

		go func(id int64) {
			defer wg.Done()

			// Semua goroutine menunggu sinyal mulai yang sama.
			<-start

			payload, _ := json.Marshal(map[string]int64{
				"student_id": id,
			})

			req, err := http.NewRequest(
				http.MethodPost,
				endpoint,
				bytes.NewBuffer(payload),
			)
			if err != nil {
				results <- Result{StudentID: id, Err: err}
				return
			}

			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")

			requestStart := time.Now()
			resp, err := client.Do(req)
			duration := time.Since(requestStart)

			if err != nil {
				results <- Result{
					StudentID: id,
					Duration:  duration,
					Err:       err,
				}
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)

			results <- Result{
				StudentID: id,
				Status:    resp.StatusCode,
				Duration:  duration,
				Body:      strings.TrimSpace(string(body)),
			}
		}(studentID)
	}

	// Lepaskan seluruh goroutine hampir bersamaan.
	close(start)
	wg.Wait()
	close(results)

	total := 0
	success := 0
	full := 0
	other := 0
	networkErrors := 0
	var durations []time.Duration

	for result := range results {
		total++

		if result.Err != nil {
			networkErrors++
			fmt.Printf("Mahasiswa %d: ERROR %v\n",
				result.StudentID, result.Err)
			continue
		}

		durations = append(durations, result.Duration)

		switch {
		case result.Status >= 200 && result.Status < 300:
			success++
		case result.Status == http.StatusConflict:
			full++
		default:
			other++
		}

		fmt.Printf(
			"Mahasiswa %d | HTTP %d | %v | %s\n",
			result.StudentID,
			result.Status,
			result.Duration,
			result.Body,
		)
	}

	elapsed := time.Since(testStart)

	fmt.Println("\n========== HASIL LOAD TEST ==========")
	fmt.Println("Total request :", total)
	fmt.Println("Berhasil      :", success)
	fmt.Println("Kuota penuh   :", full)
	fmt.Println("Status lain   :", other)
	fmt.Println("Network error :", networkErrors)
	fmt.Println("Durasi total  :", elapsed)

	if len(durations) > 0 {
		sort.Slice(durations, func(i, j int) bool {
			return durations[i] < durations[j]
		})
		fmt.Println("Respons tercepat :", durations[0])
		fmt.Println("Respons terlambat:", durations[len(durations)-1])
	}

	fmt.Println("=====================================")
}
