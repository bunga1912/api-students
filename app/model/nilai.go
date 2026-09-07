package model

type Nilai struct {
	ID          int     `json:"id"`
	NamaMatkul  string  `json:"nama_matkul"`
	Nilai       float64 `json:"nilai"`
	NIM         string  `json:"nim"`
	NamaStudent string  `json:"nama_student"`
}