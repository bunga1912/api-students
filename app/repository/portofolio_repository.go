package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"api-students/app/model"
)

var (
	ErrPortfolioNotFound = errors.New("portofolio belum dibuat")
	ErrProjectNotFound   = errors.New("proyek tidak ditemukan")
	ErrProjectLimit      = errors.New("jumlah proyek sudah mencapai batas")
)

// MaxProjectsPerPortfolio membatasi pertumbuhan array projects. Satu dokumen
// MongoDB maksimal 16 MB; array yang dibiarkan tumbuh tanpa batas suatu saat
// akan menabrak batas itu.
const MaxProjectsPerPortfolio = 20

const dateLayout = "2006-01-02"

type PortfolioRepository interface {
	EnsureIndexes(ctx context.Context) error
	FindByStudent(ctx context.Context, studentID int) (model.Portfolio, error)
	Upsert(ctx context.Context, studentID int, p model.Portfolio) (saved model.Portfolio, created bool, err error)
	AddProject(ctx context.Context, studentID int, p model.Project) (model.Portfolio, error)
	RemoveProject(ctx context.Context, studentID int, projectID string) error
	TopTechnologies(ctx context.Context, limit int) ([]model.TechnologyCount, error)
}

// portfolioDocument adalah bentuk portofolio SEBAGAIMANA TERSIMPAN di MongoDB.
// Tag bson menentukan nama field di database; package model tetap bebas dari
// tipe milik driver.
type portfolioDocument struct {
	ID           bson.ObjectID                  `bson:"_id,omitempty"`
	StudentID    int                            `bson:"student_id"`
	Headline     string                         `bson:"headline"`
	Bio          string                         `bson:"bio"`
	Projects     []portfolioProjectDocument     `bson:"projects"`
	Certificates []portfolioCertificateDocument `bson:"certificates"`
	CreatedAt    time.Time                      `bson:"created_at"`
	UpdatedAt    time.Time                      `bson:"updated_at"`
}

type portfolioProjectDocument struct {
	ID           bson.ObjectID `bson:"_id"`
	Title        string        `bson:"title"`
	Technologies []string      `bson:"technologies"`
	URL          string        `bson:"url"`
	Year         int           `bson:"year"`
}

type portfolioCertificateDocument struct {
	Name     string    `bson:"name"`
	Issuer   string    `bson:"issuer"`
	IssuedAt time.Time `bson:"issued_at"`
}

func newProjectDocument(p model.Project) portfolioProjectDocument {
	techs := p.Technologies
	if techs == nil {
		techs = []string{}
	}
	return portfolioProjectDocument{
		ID: bson.NewObjectID(), Title: p.Title, Technologies: techs, URL: p.URL, Year: p.Year,
	}
}

func (d portfolioProjectDocument) toModel() model.Project {
	techs := d.Technologies
	if techs == nil {
		techs = []string{}
	}
	return model.Project{
		ID: d.ID.Hex(), Title: d.Title, Technologies: techs, URL: d.URL, Year: d.Year,
	}
}

func (d portfolioDocument) toModel() model.Portfolio {
	projects := make([]model.Project, 0, len(d.Projects))
	for _, p := range d.Projects {
		projects = append(projects, p.toModel())
	}
	certs := make([]model.Certificate, 0, len(d.Certificates))
	for _, c := range d.Certificates {
		certs = append(certs, model.Certificate{
			Name: c.Name, Issuer: c.Issuer, Date: c.IssuedAt.UTC().Format(dateLayout),
		})
	}
	return model.Portfolio{
		ID: d.ID.Hex(), StudentID: d.StudentID, Headline: d.Headline, Bio: d.Bio,
		Projects: projects, Certificates: certs,
		CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}
}

type portfolioMongoRepository struct{ coll *mongo.Collection }

func NewPortfolioRepository(db *mongo.Database) PortfolioRepository {
	return &portfolioMongoRepository{coll: db.Collection("student_portfolios")}
}

// EnsureIndexes dijalankan sekali saat aplikasi menyala. Aman dipanggil berulang.
func (r *portfolioMongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			// Satu mahasiswa hanya boleh punya SATU portofolio. Aturan ini
			// dijaga database, bukan hanya diperiksa di kode. Index yang sama
			// juga melayani seluruh pencarian berdasarkan student_id.
			Keys:    bson.D{{Key: "student_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("student_unique"),
		},
	})
	if err != nil {
		return fmt.Errorf("membuat index student_portfolios: %w", err)
	}
	return nil
}

func (r *portfolioMongoRepository) FindByStudent(ctx context.Context, studentID int) (model.Portfolio, error) {
	var doc portfolioDocument
	err := r.coll.FindOne(ctx, bson.M{"student_id": studentID}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Portfolio{}, ErrPortfolioNotFound
		}
		return model.Portfolio{}, fmt.Errorf("mengambil portofolio: %w", err)
	}
	return doc.toModel(), nil
}

// Upsert membuat portofolio bila belum ada, atau mengganti isinya bila sudah.
//
// $set menimpa headline, bio, projects, dan certificates; $setOnInsert hanya
// berlaku saat dokumen BARU dibuat, sehingga created_at tidak berubah ketika
// portofolio diganti. student_id tidak perlu ditulis ulang: pada upsert, field
// berfilter-sama-dengan otomatis ikut masuk ke dokumen baru.
//
// Bila dua PUT pertama datang bersamaan, index unik menggagalkan salah satunya
// dengan duplicate key. Kegagalan itu dicoba ulang sekali, dan percobaan kedua
// menemukan dokumen yang baru dibuat lalu mengubahnya.
func (r *portfolioMongoRepository) Upsert(ctx context.Context, studentID int, p model.Portfolio) (model.Portfolio, bool, error) {
	projects := make([]portfolioProjectDocument, 0, len(p.Projects))
	for _, pr := range p.Projects {
		projects = append(projects, newProjectDocument(pr))
	}

	certs := make([]portfolioCertificateDocument, 0, len(p.Certificates))
	for _, c := range p.Certificates {
		issued, err := time.Parse(dateLayout, c.Date)
		if err != nil {
			return model.Portfolio{}, false, fmt.Errorf("tanggal sertifikat tidak valid: %w", err)
		}
		certs = append(certs, portfolioCertificateDocument{Name: c.Name, Issuer: c.Issuer, IssuedAt: issued})
	}

	t := nowMillis()
	update := bson.M{
		"$set": bson.M{
			"headline":     p.Headline,
			"bio":          p.Bio,
			"projects":     projects,
			"certificates": certs,
			"updated_at":   t,
		},
		"$setOnInsert": bson.M{"created_at": t},
	}

	created := false
	for attempt := 0; attempt < 2; attempt++ {
		res, err := r.coll.UpdateOne(ctx,
			bson.M{"student_id": studentID}, update,
			options.UpdateOne().SetUpsert(true))
		if err != nil {
			if mongo.IsDuplicateKeyError(err) && attempt == 0 {
				continue
			}
			return model.Portfolio{}, false, fmt.Errorf("menyimpan portofolio: %w", err)
		}
		created = res.UpsertedCount > 0
		break
	}

	saved, err := r.FindByStudent(ctx, studentID)
	if err != nil {
		return model.Portfolio{}, false, err
	}
	return saved, created, nil
}

// AddProject menambahkan satu proyek ke dalam array projects.
//
// Syarat "projects.19 tidak ada" berarti array belum berisi 20 elemen
// (indeks 0 sampai 19). Pemeriksaan batas dan penambahan terjadi dalam satu
// operasi, sehingga sepuluh request serentak ketika tinggal dua tempat
// tersisa tidak mungkin menghasilkan lebih dari dua keberhasilan.
func (r *portfolioMongoRepository) AddProject(ctx context.Context, studentID int, p model.Project) (model.Portfolio, error) {
	doc := newProjectDocument(p)
	limitField := "projects." + strconv.Itoa(MaxProjectsPerPortfolio-1)

	var updated portfolioDocument
	err := r.coll.FindOneAndUpdate(ctx,
		bson.M{"student_id": studentID, limitField: bson.M{"$exists": false}},
		bson.M{
			"$push": bson.M{"projects": doc},
			"$set":  bson.M{"updated_at": nowMillis()},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	if err == nil {
		return updated.toModel(), nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return model.Portfolio{}, fmt.Errorf("menambah proyek: %w", err)
	}

	// Tidak ada dokumen yang cocok: portofolionya belum ada, atau sudah penuh.
	count, err := r.coll.CountDocuments(ctx, bson.M{"student_id": studentID})
	if err != nil {
		return model.Portfolio{}, fmt.Errorf("memeriksa portofolio: %w", err)
	}
	if count == 0 {
		return model.Portfolio{}, ErrPortfolioNotFound
	}
	return model.Portfolio{}, ErrProjectLimit
}

// RemoveProject menghapus satu proyek dengan $pull.
//
// Filter memuat "projects._id", sehingga operasi hanya mengenai dokumen yang
// MEMANG berisi proyek itu. Bila proyeknya tidak ada, tidak ada yang berubah,
// termasuk updated_at.
func (r *portfolioMongoRepository) RemoveProject(ctx context.Context, studentID int, projectID string) error {
	pid, err := parseObjectID(projectID)
	if err != nil {
		return err
	}

	res, err := r.coll.UpdateOne(ctx,
		bson.M{"student_id": studentID, "projects._id": pid},
		bson.M{
			"$pull": bson.M{"projects": bson.M{"_id": pid}},
			"$set":  bson.M{"updated_at": nowMillis()},
		})
	if err != nil {
		return fmt.Errorf("menghapus proyek: %w", err)
	}
	if res.MatchedCount > 0 {
		return nil
	}

	count, err := r.coll.CountDocuments(ctx, bson.M{"student_id": studentID})
	if err != nil {
		return fmt.Errorf("memeriksa portofolio: %w", err)
	}
	if count == 0 {
		return ErrPortfolioNotFound
	}
	return ErrProjectNotFound
}

// TopTechnologies menghitung teknologi yang paling banyak dipakai di seluruh
// portofolio. Dua $unwind dibutuhkan karena datanya bersarang dua tingkat:
// portofolio berisi proyek, dan proyek berisi teknologi. Tiap $unwind
// "membongkar" satu tingkat array menjadi satu dokumen per elemen.
func (r *portfolioMongoRepository) TopTechnologies(ctx context.Context, limit int) ([]model.TechnologyCount, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: "$projects"}},
		{{Key: "$unwind", Value: "$projects.technologies"}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$projects.technologies",
			"total": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "total", Value: -1}, {Key: "_id", Value: 1}}}},
		{{Key: "$limit", Value: int64(limit)}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("menghitung teknologi: %w", err)
	}
	defer cursor.Close(ctx)

	var rows []struct {
		Technology string `bson:"_id"`
		Total      int    `bson:"total"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, fmt.Errorf("membaca statistik teknologi: %w", err)
	}

	result := make([]model.TechnologyCount, 0, len(rows))
	for _, row := range rows {
		result = append(result, model.TechnologyCount{Technology: row.Technology, Total: row.Total})
	}
	return result, nil
}