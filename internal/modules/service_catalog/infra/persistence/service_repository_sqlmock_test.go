package persistence

import (
	"context"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/domain/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var serviceColumns = []string{"id", "name", "description", "price", "created_at", "updated_at", "deleted_at"}

func setupServiceMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	dialector := postgres.New(postgres.Config{Conn: sqlDB, DriverName: "postgres"})
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}
	return db, mock, func() { _ = sqlDB.Close() }
}

func buildServiceRow(id uuid.UUID) *sqlmock.Rows {
	return sqlmock.NewRows(serviceColumns).AddRow(
		id, "Troca de Óleo", "Troca de óleo do motor com filtro", 15000,
		time.Now(), time.Now(), nil,
	)
}

// ─── TableName ────────────────────────────────────────────────────────────────

func TestServiceModel_TableName(t *testing.T) {
	if (ServiceModel{}).TableName() != "services" {
		t.Error("expected table name 'services'")
	}
}

// ─── Save / Create ────────────────────────────────────────────────────────────

func TestServiceRepository_Sqlmock_Save_Create_Success(t *testing.T) {
	db, mock, cleanup := setupServiceMockDB(t)
	defer cleanup()
	repo := NewServiceRepository(db)

	uid := uuid.New()
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "services"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uid))
	mock.ExpectCommit()

	svc, _ := service.NewService("Troca de Óleo", "Troca de óleo do motor com filtro", 15000)
	if err := repo.Save(context.Background(), svc); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet: %v", err)
	}
}

func TestServiceRepository_Sqlmock_Save_Create_DBError(t *testing.T) {
	db, mock, cleanup := setupServiceMockDB(t)
	defer cleanup()
	repo := NewServiceRepository(db)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "services"`).WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	svc, _ := service.NewService("Troca de Óleo", "Troca de óleo do motor com filtro", 15000)
	if err := repo.Save(context.Background(), svc); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ─── Save / Update ────────────────────────────────────────────────────────────

func TestServiceRepository_Sqlmock_Save_Update_Success(t *testing.T) {
	db, mock, cleanup := setupServiceMockDB(t)
	defer cleanup()
	repo := NewServiceRepository(db)

	uid := uuid.New()
	now := time.Now()
	svc, _ := service.ReconstructService(uid.String(), "Troca de Óleo", "Troca de óleo do motor com filtro", 15000, now, now, nil)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "services" SET`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.Save(context.Background(), svc); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet: %v", err)
	}
}

func TestServiceRepository_Sqlmock_Save_Update_DBError(t *testing.T) {
	db, mock, cleanup := setupServiceMockDB(t)
	defer cleanup()
	repo := NewServiceRepository(db)

	uid := uuid.New()
	now := time.Now()
	svc, _ := service.ReconstructService(uid.String(), "Troca de Óleo", "Troca de óleo do motor com filtro", 15000, now, now, nil)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "services" SET`).WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	if err := repo.Save(context.Background(), svc); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ─── FindByID ─────────────────────────────────────────────────────────────────

func TestServiceRepository_Sqlmock_FindByID_InvalidUUID(t *testing.T) {
	db, _, cleanup := setupServiceMockDB(t)
	defer cleanup()
	repo := NewServiceRepository(db)

	if _, err := repo.FindByID(context.Background(), "not-a-uuid"); err != service.ErrInvalidServiceID {
		t.Fatalf("expected ErrInvalidServiceID, got %v", err)
	}
}

func TestServiceRepository_Sqlmock_FindByID_Success(t *testing.T) {
	db, mock, cleanup := setupServiceMockDB(t)
	defer cleanup()
	repo := NewServiceRepository(db)

	uid := uuid.New()
	mock.ExpectQuery(`SELECT .* FROM "services" WHERE id`).WillReturnRows(buildServiceRow(uid))

	svc, err := repo.FindByID(context.Background(), uid.String())
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if svc == nil {
		t.Fatal("expected service, got nil")
	}
}

func TestServiceRepository_Sqlmock_FindByID_NotFound(t *testing.T) {
	db, mock, cleanup := setupServiceMockDB(t)
	defer cleanup()
	repo := NewServiceRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "services" WHERE id`).WillReturnError(gorm.ErrRecordNotFound)

	if _, err := repo.FindByID(context.Background(), uuid.New().String()); err != service.ErrServiceNotFound {
		t.Fatalf("expected ErrServiceNotFound, got %v", err)
	}
}

func TestServiceRepository_Sqlmock_FindByID_DBError(t *testing.T) {
	db, mock, cleanup := setupServiceMockDB(t)
	defer cleanup()
	repo := NewServiceRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "services" WHERE id`).WillReturnError(errors.New("db error"))

	if _, err := repo.FindByID(context.Background(), uuid.New().String()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ─── FindAll ──────────────────────────────────────────────────────────────────

func TestServiceRepository_Sqlmock_FindAll_Success(t *testing.T) {
	db, mock, cleanup := setupServiceMockDB(t)
	defer cleanup()
	repo := NewServiceRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "services"`).WillReturnRows(buildServiceRow(uuid.New()))

	svcs, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(svcs) != 1 {
		t.Fatalf("expected 1 service, got %d", len(svcs))
	}
}

func TestServiceRepository_Sqlmock_FindAll_Empty(t *testing.T) {
	db, mock, cleanup := setupServiceMockDB(t)
	defer cleanup()
	repo := NewServiceRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "services"`).WillReturnRows(sqlmock.NewRows(serviceColumns))

	svcs, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(svcs) != 0 {
		t.Fatalf("expected 0 services, got %d", len(svcs))
	}
}

func TestServiceRepository_Sqlmock_FindAll_DBError(t *testing.T) {
	db, mock, cleanup := setupServiceMockDB(t)
	defer cleanup()
	repo := NewServiceRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "services"`).WillReturnError(errors.New("db error"))

	if _, err := repo.FindAll(context.Background()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ─── ExistsByName ─────────────────────────────────────────────────────────────

func TestServiceRepository_Sqlmock_ExistsByName_True(t *testing.T) {
	db, mock, cleanup := setupServiceMockDB(t)
	defer cleanup()
	repo := NewServiceRepository(db)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "services" WHERE name`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	exists, err := repo.ExistsByName(context.Background(), "Troca de Óleo")
	if err != nil || !exists {
		t.Fatalf("expected true/nil, got %v/%v", exists, err)
	}
}

func TestServiceRepository_Sqlmock_ExistsByName_False(t *testing.T) {
	db, mock, cleanup := setupServiceMockDB(t)
	defer cleanup()
	repo := NewServiceRepository(db)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "services" WHERE name`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	exists, err := repo.ExistsByName(context.Background(), "Troca de Óleo")
	if err != nil || exists {
		t.Fatalf("expected false/nil, got %v/%v", exists, err)
	}
}

func TestServiceRepository_Sqlmock_ExistsByName_DBError(t *testing.T) {
	db, mock, cleanup := setupServiceMockDB(t)
	defer cleanup()
	repo := NewServiceRepository(db)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "services" WHERE name`).
		WillReturnError(errors.New("db error"))

	if _, err := repo.ExistsByName(context.Background(), "Troca de Óleo"); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func TestServiceRepository_Sqlmock_Delete_InvalidUUID(t *testing.T) {
	db, _, cleanup := setupServiceMockDB(t)
	defer cleanup()
	repo := NewServiceRepository(db)

	if err := repo.Delete(context.Background(), "not-a-uuid"); err != service.ErrInvalidServiceID {
		t.Fatalf("expected ErrInvalidServiceID, got %v", err)
	}
}

func TestServiceRepository_Sqlmock_Delete_NotFound(t *testing.T) {
	db, mock, cleanup := setupServiceMockDB(t)
	defer cleanup()
	repo := NewServiceRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "services" WHERE id`).WillReturnError(gorm.ErrRecordNotFound)

	if err := repo.Delete(context.Background(), uuid.New().String()); err != service.ErrServiceNotFound {
		t.Fatalf("expected ErrServiceNotFound, got %v", err)
	}
}

func TestServiceRepository_Sqlmock_Delete_FirstDBError(t *testing.T) {
	db, mock, cleanup := setupServiceMockDB(t)
	defer cleanup()
	repo := NewServiceRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "services" WHERE id`).WillReturnError(errors.New("db error"))

	if err := repo.Delete(context.Background(), uuid.New().String()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestServiceRepository_Sqlmock_Delete_Success(t *testing.T) {
	db, mock, cleanup := setupServiceMockDB(t)
	defer cleanup()
	repo := NewServiceRepository(db)

	uid := uuid.New()
	mock.ExpectQuery(`SELECT .* FROM "services" WHERE id`).WillReturnRows(buildServiceRow(uid))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "services" SET "deleted_at"`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.Delete(context.Background(), uid.String()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet: %v", err)
	}
}

func TestServiceRepository_Sqlmock_Delete_SoftDeleteError(t *testing.T) {
	db, mock, cleanup := setupServiceMockDB(t)
	defer cleanup()
	repo := NewServiceRepository(db)

	uid := uuid.New()
	mock.ExpectQuery(`SELECT .* FROM "services" WHERE id`).WillReturnRows(buildServiceRow(uid))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "services" SET "deleted_at"`).WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	if err := repo.Delete(context.Background(), uid.String()); err == nil {
		t.Fatal("expected error, got nil")
	}
}
