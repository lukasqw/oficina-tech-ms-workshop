package persistence

import (
	"context"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/domain/service"
	"gorm.io/gorm"
)

func TestServiceRepository_Sqlmock_Save_Update_DuplicateKey(t *testing.T) {
	db, mock, cleanup := setupServiceMockDB(t)
	defer cleanup()
	repo := NewServiceRepository(db)

	uid := uuid.New()
	now := time.Now()
	svc, _ := service.ReconstructService(uid.String(), "Troca de Óleo", "Troca de óleo do motor com filtro", 15000, now, now, nil)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "services" SET`).WillReturnError(gorm.ErrDuplicatedKey)
	mock.ExpectRollback()

	if err := repo.Save(context.Background(), svc); err != service.ErrDuplicateServiceName {
		t.Fatalf("expected ErrDuplicateServiceName, got %v", err)
	}
}

func TestServiceRepository_Sqlmock_Save_Create_DuplicateKey(t *testing.T) {
	db, mock, cleanup := setupServiceMockDB(t)
	defer cleanup()
	repo := NewServiceRepository(db)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "services"`).WillReturnError(gorm.ErrDuplicatedKey)
	mock.ExpectRollback()

	svc, _ := service.NewService("Troca de Óleo", "Troca de óleo do motor com filtro", 15000)
	if err := repo.Save(context.Background(), svc); err != service.ErrDuplicateServiceName {
		t.Fatalf("expected ErrDuplicateServiceName, got %v", err)
	}
}

func TestServiceRepository_Sqlmock_FindAll_MultipleRows(t *testing.T) {
	db, mock, cleanup := setupServiceMockDB(t)
	defer cleanup()
	repo := NewServiceRepository(db)

	rows := sqlmock.NewRows(serviceColumns).
		AddRow(uuid.New(), "Serviço 1", "Descrição do serviço um completo", 10000, time.Now(), time.Now(), nil).
		AddRow(uuid.New(), "Serviço 2", "Descrição do serviço dois completo", 20000, time.Now(), time.Now(), nil)
	mock.ExpectQuery(`SELECT .* FROM "services"`).WillReturnRows(rows)

	svcs, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(svcs) != 2 {
		t.Fatalf("expected 2 services, got %d", len(svcs))
	}
}
