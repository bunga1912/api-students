package model

type Course struct {
	ID       int64  `json:"id"`
	KodeMK   string `json:"kode_mk"`
	NamaMK   string `json:"nama_mk"`
	SKS      int    `json:"sks"`
	Semester int    `json:"semester"`
	Kuota    int    `json:"kuota"`
}
