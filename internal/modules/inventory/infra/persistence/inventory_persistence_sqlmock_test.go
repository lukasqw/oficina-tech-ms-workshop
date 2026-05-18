package persistence

import (
	"context"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/inventory"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/product"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/inventory/domain/saga_operation"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ─── Shared setup ─────────────────────────────────────────────────────────────

func setupInventoryMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
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

// ─── ProductModel ─────────────────────────────────────────────────────────────

var productCols = []string{"id", "name", "description", "price", "product_type", "created_at", "updated_at", "deleted_at"}

func buildProductRow(id uuid.UUID) *sqlmock.Rows {
	return sqlmock.NewRows(productCols).AddRow(
		id, "Óleo de Motor", "Óleo sintético 5W30", 5000, "CONSUMABLE",
		time.Now(), time.Now(), nil,
	)
}

func TestProductModel_TableNameSqlmock(t *testing.T) {
	if (ProductModel{}).TableName() != "product_models" {
		t.Error("expected table name 'product_models'")
	}
}

func TestProductModel_ToDomainSqlmock(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	m := &ProductModel{
		ID: id, Name: "Óleo de Motor", Description: "Desc", Price: 5000,
		ProductType: "CONSUMABLE", CreatedAt: now, UpdatedAt: now,
	}
	p, err := m.ToDomain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID() != id.String() {
		t.Errorf("ID: got %v, want %v", p.ID(), id.String())
	}
}

func TestProductModel_ToDomain_WithDeletedAt(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	m := &ProductModel{
		ID: id, Name: "Óleo de Motor", Description: "Desc", Price: 5000,
		ProductType: "PARTS", CreatedAt: now, UpdatedAt: now,
		DeletedAt: gorm.DeletedAt{Time: now, Valid: true},
	}
	p, err := m.ToDomain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.DeletedAt() == nil {
		t.Fatal("expected non-nil DeletedAt")
	}
}

func TestFromDomain_WithID(t *testing.T) {
	pt, _ := product.NewProductType("CONSUMABLE")
	p, _ := product.NewProduct("Óleo", "Desc", 5000, pt)
	_ = p.SetID(uuid.New().String())

	m, err := FromDomain(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name != "Óleo" {
		t.Errorf("Name: got %v, want Óleo", m.Name)
	}
}

func TestFromDomain_WithoutID(t *testing.T) {
	pt, _ := product.NewProductType("PARTS")
	p, _ := product.NewProduct("Filtro", "Filtro de óleo", 3000, pt)

	m, err := FromDomain(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ID != uuid.Nil {
		t.Errorf("expected nil UUID, got %v", m.ID)
	}
}

func TestFromDomain_WithDeletedAt(t *testing.T) {
	pt, _ := product.NewProductType("SIMPLE")
	now := time.Now()
	p, _ := product.ReconstructProduct(
		uuid.New().String(), "Produto", "Desc", 1000, pt, now, now, &now,
	)
	m, err := FromDomain(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !m.DeletedAt.Valid {
		t.Error("expected DeletedAt.Valid to be true")
	}
}

// ─── ProductRepository ────────────────────────────────────────────────────────

func TestProductRepository_Sqlmock_Save_Create_Success(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewProductRepository(db)

	pt, _ := product.NewProductType("CONSUMABLE")
	p, _ := product.NewProduct("Óleo de Motor", "Óleo sintético 5W30", 5000, pt)

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "product_models"`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.Save(context.Background(), p); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet: %v", err)
	}
}

func TestProductRepository_Sqlmock_Save_Create_DBError(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewProductRepository(db)

	pt, _ := product.NewProductType("CONSUMABLE")
	p, _ := product.NewProduct("Óleo de Motor", "Óleo sintético 5W30", 5000, pt)

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "product_models"`).WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	if err := repo.Save(context.Background(), p); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProductRepository_Sqlmock_Save_Update_Success(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewProductRepository(db)

	pt, _ := product.NewProductType("PARTS")
	now := time.Now()
	p, _ := product.ReconstructProduct(uuid.New().String(), "Filtro", "Filtro original", 3000, pt, now, now, nil)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "product_models" SET`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.Save(context.Background(), p); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestProductRepository_Sqlmock_Save_Update_DBError(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewProductRepository(db)

	pt, _ := product.NewProductType("PARTS")
	now := time.Now()
	p, _ := product.ReconstructProduct(uuid.New().String(), "Filtro", "Filtro original", 3000, pt, now, now, nil)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "product_models" SET`).WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	if err := repo.Save(context.Background(), p); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProductRepository_Sqlmock_FindByID_InvalidUUID(t *testing.T) {
	db, _, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewProductRepository(db)

	if _, err := repo.FindByID(context.Background(), "bad-uuid"); err != product.ErrInvalidProductID {
		t.Fatalf("expected ErrInvalidProductID, got %v", err)
	}
}

func TestProductRepository_Sqlmock_FindByID_Success(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewProductRepository(db)

	uid := uuid.New()
	mock.ExpectQuery(`SELECT .* FROM "product_models" WHERE id`).WillReturnRows(buildProductRow(uid))

	p, err := repo.FindByID(context.Background(), uid.String())
	if err != nil || p == nil {
		t.Fatalf("expected product, got err=%v", err)
	}
}

func TestProductRepository_Sqlmock_FindByID_NotFound(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewProductRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "product_models" WHERE id`).WillReturnError(gorm.ErrRecordNotFound)

	if _, err := repo.FindByID(context.Background(), uuid.New().String()); err != product.ErrProductNotFound {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}

func TestProductRepository_Sqlmock_FindByID_DBError(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewProductRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "product_models" WHERE id`).WillReturnError(errors.New("db error"))

	if _, err := repo.FindByID(context.Background(), uuid.New().String()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProductRepository_Sqlmock_FindAll_Success(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewProductRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "product_models"`).WillReturnRows(buildProductRow(uuid.New()))

	products, err := repo.FindAll(context.Background())
	if err != nil || len(products) != 1 {
		t.Fatalf("expected 1 product, got err=%v len=%d", err, len(products))
	}
}

func TestProductRepository_Sqlmock_FindAll_Empty(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewProductRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "product_models"`).WillReturnRows(sqlmock.NewRows(productCols))

	products, err := repo.FindAll(context.Background())
	if err != nil || len(products) != 0 {
		t.Fatalf("expected 0 products, got err=%v len=%d", err, len(products))
	}
}

func TestProductRepository_Sqlmock_FindAll_DBError(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewProductRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "product_models"`).WillReturnError(errors.New("db error"))

	if _, err := repo.FindAll(context.Background()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProductRepository_Sqlmock_Delete_InvalidUUID(t *testing.T) {
	db, _, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewProductRepository(db)

	if err := repo.Delete(context.Background(), "bad-uuid"); err != product.ErrInvalidProductID {
		t.Fatalf("expected ErrInvalidProductID, got %v", err)
	}
}

func TestProductRepository_Sqlmock_Delete_NotFound(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewProductRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "product_models" WHERE id`).WillReturnError(gorm.ErrRecordNotFound)

	if err := repo.Delete(context.Background(), uuid.New().String()); err != product.ErrProductNotFound {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}

func TestProductRepository_Sqlmock_Delete_FirstDBError(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewProductRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "product_models" WHERE id`).WillReturnError(errors.New("db error"))

	if err := repo.Delete(context.Background(), uuid.New().String()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProductRepository_Sqlmock_Delete_Success(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewProductRepository(db)

	uid := uuid.New()
	mock.ExpectQuery(`SELECT .* FROM "product_models" WHERE id`).WillReturnRows(buildProductRow(uid))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "product_models" SET "deleted_at"`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.Delete(context.Background(), uid.String()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestProductRepository_Sqlmock_Delete_SoftDeleteError(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewProductRepository(db)

	uid := uuid.New()
	mock.ExpectQuery(`SELECT .* FROM "product_models" WHERE id`).WillReturnRows(buildProductRow(uid))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "product_models" SET "deleted_at"`).WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	if err := repo.Delete(context.Background(), uid.String()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ─── InventoryModel ───────────────────────────────────────────────────────────

var inventoryCols = []string{"id", "product_id", "available_quantity", "reserved_quantity", "pending_quantity", "created_at", "updated_at", "deleted_at"}

func buildInventoryRow(id, productID uuid.UUID) *sqlmock.Rows {
	return sqlmock.NewRows(inventoryCols).AddRow(
		id, productID, 100, 20, 5, time.Now(), time.Now(), nil,
	)
}

func TestInventoryModel_TableNameSqlmock(t *testing.T) {
	if (InventoryModel{}).TableName() != "inventories" {
		t.Error("expected table name 'inventories'")
	}
}

func TestInventoryModel_ToDomainSqlmock(t *testing.T) {
	id := uuid.New()
	productID := uuid.New()
	now := time.Now()
	m := &InventoryModel{
		ID: id, ProductID: productID,
		AvailableQuantity: 100, ReservedQuantity: 20, PendingQuantity: 5,
		CreatedAt: now, UpdatedAt: now,
	}
	inv, err := m.ToDomain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.ID() != id.String() {
		t.Errorf("ID: got %v, want %v", inv.ID(), id.String())
	}
	if inv.AvailableQuantity() != 100 {
		t.Errorf("AvailableQuantity: got %v, want 100", inv.AvailableQuantity())
	}
}

func TestInventoryModel_ToDomain_WithDeletedAt(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	m := &InventoryModel{
		ID: id, ProductID: uuid.New(),
		CreatedAt: now, UpdatedAt: now,
		DeletedAt: gorm.DeletedAt{Time: now, Valid: true},
	}
	inv, err := m.ToDomain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.DeletedAt() == nil {
		t.Fatal("expected non-nil DeletedAt")
	}
}

func TestFromDomainInventory_WithIDSqlmock(t *testing.T) {
	productID := uuid.New()
	id := uuid.New()
	inv, _ := inventory.ReconstructInventory(
		id.String(), productID.String(), 100, 20, 5, time.Now(), time.Now(), nil,
	)
	m, err := FromDomainInventory(inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ID.String() != id.String() {
		t.Errorf("ID: got %v, want %v", m.ID.String(), id.String())
	}
}

func TestFromDomainInventory_WithoutIDSqlmock(t *testing.T) {
	productID := uuid.New()
	inv, _ := inventory.NewInventory(productID.String())
	m, err := FromDomainInventory(inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ProductID.String() != productID.String() {
		t.Errorf("ProductID: got %v, want %v", m.ProductID.String(), productID.String())
	}
}

func TestFromDomainInventory_WithDeletedAt(t *testing.T) {
	productID := uuid.New()
	id := uuid.New()
	now := time.Now()
	inv, _ := inventory.ReconstructInventory(id.String(), productID.String(), 10, 0, 0, now, now, &now)
	m, err := FromDomainInventory(inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !m.DeletedAt.Valid {
		t.Error("expected DeletedAt.Valid to be true")
	}
}

// ─── InventoryRepository ──────────────────────────────────────────────────────

func TestInventoryRepository_Sqlmock_Save_Create_Success(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	productID := uuid.New()
	inv, _ := inventory.NewInventory(productID.String())

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "inventories"`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.Save(context.Background(), inv); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestInventoryRepository_Sqlmock_Save_Create_DBError(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	productID := uuid.New()
	inv, _ := inventory.NewInventory(productID.String())

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "inventories"`).WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	if err := repo.Save(context.Background(), inv); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestInventoryRepository_Sqlmock_Save_Update_Success(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	id := uuid.New()
	productID := uuid.New()
	now := time.Now()
	inv, _ := inventory.ReconstructInventory(id.String(), productID.String(), 100, 0, 0, now, now, nil)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "inventories" SET`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.Save(context.Background(), inv); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestInventoryRepository_Sqlmock_Save_Update_DBError(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	id := uuid.New()
	productID := uuid.New()
	now := time.Now()
	inv, _ := inventory.ReconstructInventory(id.String(), productID.String(), 100, 0, 0, now, now, nil)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "inventories" SET`).WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	if err := repo.Save(context.Background(), inv); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestInventoryRepository_Sqlmock_FindByID_InvalidUUID(t *testing.T) {
	db, _, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	if _, err := repo.FindByID(context.Background(), "bad-uuid"); err != inventory.ErrInvalidInventoryID {
		t.Fatalf("expected ErrInvalidInventoryID, got %v", err)
	}
}

func TestInventoryRepository_Sqlmock_FindByID_Success(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	id := uuid.New()
	mock.ExpectQuery(`SELECT .* FROM "inventories" WHERE id`).
		WillReturnRows(buildInventoryRow(id, uuid.New()))

	inv, err := repo.FindByID(context.Background(), id.String())
	if err != nil || inv == nil {
		t.Fatalf("expected inventory, got err=%v", err)
	}
}

func TestInventoryRepository_Sqlmock_FindByID_NotFound(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "inventories" WHERE id`).WillReturnError(gorm.ErrRecordNotFound)

	if _, err := repo.FindByID(context.Background(), uuid.New().String()); err != inventory.ErrInventoryNotFound {
		t.Fatalf("expected ErrInventoryNotFound, got %v", err)
	}
}

func TestInventoryRepository_Sqlmock_FindByID_DBError(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "inventories" WHERE id`).WillReturnError(errors.New("db error"))

	if _, err := repo.FindByID(context.Background(), uuid.New().String()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestInventoryRepository_Sqlmock_FindByProductID_InvalidUUID(t *testing.T) {
	db, _, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	if _, err := repo.FindByProductID(context.Background(), "bad-uuid"); err != inventory.ErrInvalidProductID {
		t.Fatalf("expected ErrInvalidProductID, got %v", err)
	}
}

func TestInventoryRepository_Sqlmock_FindByProductID_Success(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	productID := uuid.New()
	mock.ExpectQuery(`SELECT .* FROM "inventories" WHERE product_id`).
		WillReturnRows(buildInventoryRow(uuid.New(), productID))

	inv, err := repo.FindByProductID(context.Background(), productID.String())
	if err != nil || inv == nil {
		t.Fatalf("expected inventory, got err=%v", err)
	}
}

func TestInventoryRepository_Sqlmock_FindByProductID_NotFound(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "inventories" WHERE product_id`).WillReturnError(gorm.ErrRecordNotFound)

	if _, err := repo.FindByProductID(context.Background(), uuid.New().String()); err != inventory.ErrInventoryNotFound {
		t.Fatalf("expected ErrInventoryNotFound, got %v", err)
	}
}

func TestInventoryRepository_Sqlmock_FindAll_Success(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "inventories"`).
		WillReturnRows(buildInventoryRow(uuid.New(), uuid.New()))

	invs, err := repo.FindAll(context.Background())
	if err != nil || len(invs) != 1 {
		t.Fatalf("expected 1 inventory, got err=%v len=%d", err, len(invs))
	}
}

func TestInventoryRepository_Sqlmock_FindAll_Empty(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "inventories"`).WillReturnRows(sqlmock.NewRows(inventoryCols))

	invs, err := repo.FindAll(context.Background())
	if err != nil || len(invs) != 0 {
		t.Fatalf("expected 0 inventories, got err=%v len=%d", err, len(invs))
	}
}

func TestInventoryRepository_Sqlmock_FindAll_DBError(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "inventories"`).WillReturnError(errors.New("db error"))

	if _, err := repo.FindAll(context.Background()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestInventoryRepository_Sqlmock_Delete_InvalidUUID(t *testing.T) {
	db, _, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	if err := repo.Delete(context.Background(), "bad-uuid"); err != inventory.ErrInvalidInventoryID {
		t.Fatalf("expected ErrInvalidInventoryID, got %v", err)
	}
}

func TestInventoryRepository_Sqlmock_Delete_NotFound(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "inventories" WHERE id`).WillReturnError(gorm.ErrRecordNotFound)

	if err := repo.Delete(context.Background(), uuid.New().String()); err != inventory.ErrInventoryNotFound {
		t.Fatalf("expected ErrInventoryNotFound, got %v", err)
	}
}

func TestInventoryRepository_Sqlmock_Delete_FirstDBError(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "inventories" WHERE id`).WillReturnError(errors.New("db error"))

	if err := repo.Delete(context.Background(), uuid.New().String()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestInventoryRepository_Sqlmock_Delete_Success(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	id := uuid.New()
	mock.ExpectQuery(`SELECT .* FROM "inventories" WHERE id`).WillReturnRows(buildInventoryRow(id, uuid.New()))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "inventories" SET "deleted_at"`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.Delete(context.Background(), id.String()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestInventoryRepository_Sqlmock_Delete_SoftDeleteError(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	id := uuid.New()
	mock.ExpectQuery(`SELECT .* FROM "inventories" WHERE id`).WillReturnRows(buildInventoryRow(id, uuid.New()))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "inventories" SET "deleted_at"`).WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	if err := repo.Delete(context.Background(), id.String()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestInventoryRepository_Sqlmock_ExistsByProductID_InvalidUUID(t *testing.T) {
	db, _, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	if _, err := repo.ExistsByProductID(context.Background(), "bad-uuid"); err != inventory.ErrInvalidProductID {
		t.Fatalf("expected ErrInvalidProductID, got %v", err)
	}
}

func TestInventoryRepository_Sqlmock_ExistsByProductID_True(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "inventories" WHERE product_id`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	exists, err := repo.ExistsByProductID(context.Background(), uuid.New().String())
	if err != nil || !exists {
		t.Fatalf("expected true/nil, got %v/%v", exists, err)
	}
}

func TestInventoryRepository_Sqlmock_ExistsByProductID_False(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "inventories" WHERE product_id`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	exists, err := repo.ExistsByProductID(context.Background(), uuid.New().String())
	if err != nil || exists {
		t.Fatalf("expected false/nil, got %v/%v", exists, err)
	}
}

func TestInventoryRepository_Sqlmock_ExistsByProductID_DBError(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewInventoryRepository(db)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "inventories" WHERE product_id`).
		WillReturnError(errors.New("db error"))

	if _, err := repo.ExistsByProductID(context.Background(), uuid.New().String()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ─── SagaOperationModel ───────────────────────────────────────────────────────

var sagaCols = []string{"id", "saga_id", "order_id", "operation", "status", "result_payload", "processed_at"}

func buildSagaRow(id, sagaID, orderID uuid.UUID) *sqlmock.Rows {
	return sqlmock.NewRows(sagaCols).AddRow(
		id, sagaID, orderID, "RESERVE", "PROCESSING", []byte("{}"), nil,
	)
}

func TestSagaOperationModel_TableName(t *testing.T) {
	if (SagaOperationModel{}).TableName() != "saga_operations" {
		t.Error("expected table name 'saga_operations'")
	}
}

func TestSagaOperationModel_ToDomainSqlmock(t *testing.T) {
	id := uuid.New()
	sagaID := uuid.New()
	orderID := uuid.New()
	m := &SagaOperationModel{
		ID: id, SagaID: sagaID, OrderID: orderID,
		Operation: "RESERVE", Status: "PROCESSING",
		ResultPayload: []byte("{}"), ProcessedAt: nil,
	}
	op, err := m.ToDomain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if op.ID() != id.String() {
		t.Errorf("ID: got %v, want %v", op.ID(), id.String())
	}
	if string(op.Operation()) != "RESERVE" {
		t.Errorf("Operation: got %v, want RESERVE", op.Operation())
	}
}

func TestFromDomainSagaOperation_WithIDSqlmock(t *testing.T) {
	sagaID := uuid.New()
	orderID := uuid.New()
	id := uuid.New()
	op, _ := saga_operation.ReconstructSagaOperation(
		id.String(), sagaID.String(), orderID.String(),
		saga_operation.OperationReserve, saga_operation.StatusProcessing,
		[]byte("{}"), nil,
	)
	m, err := FromDomainSagaOperation(op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ID.String() != id.String() {
		t.Errorf("ID: got %v, want %v", m.ID.String(), id.String())
	}
}

func TestFromDomainSagaOperation_WithoutIDSqlmock(t *testing.T) {
	sagaID := uuid.New()
	orderID := uuid.New()
	op, _ := saga_operation.NewSagaOperation(sagaID.String(), orderID.String(), saga_operation.OperationReserve)
	m, err := FromDomainSagaOperation(op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.SagaID.String() != sagaID.String() {
		t.Errorf("SagaID: got %v, want %v", m.SagaID.String(), sagaID.String())
	}
}

// ─── SagaOperationRepository ──────────────────────────────────────────────────

func TestSagaOperationRepository_Sqlmock_Save_Create_Success(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewSagaOperationRepository(db)

	sagaID := uuid.New()
	orderID := uuid.New()
	op, _ := saga_operation.NewSagaOperation(sagaID.String(), orderID.String(), saga_operation.OperationReserve)

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "saga_operations"`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.Save(context.Background(), op); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestSagaOperationRepository_Sqlmock_Save_Create_DBError(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewSagaOperationRepository(db)

	sagaID := uuid.New()
	orderID := uuid.New()
	op, _ := saga_operation.NewSagaOperation(sagaID.String(), orderID.String(), saga_operation.OperationReserve)

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "saga_operations"`).WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	if err := repo.Save(context.Background(), op); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSagaOperationRepository_Sqlmock_Save_Update_Success(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewSagaOperationRepository(db)

	id := uuid.New()
	sagaID := uuid.New()
	orderID := uuid.New()
	op, _ := saga_operation.ReconstructSagaOperation(
		id.String(), sagaID.String(), orderID.String(),
		saga_operation.OperationReserve, saga_operation.StatusProcessing,
		[]byte("{}"), nil,
	)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "saga_operations" SET`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.Save(context.Background(), op); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestSagaOperationRepository_Sqlmock_FindBySagaAndOperation_InvalidUUID(t *testing.T) {
	db, _, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewSagaOperationRepository(db)

	if _, err := repo.FindBySagaAndOperation(context.Background(), "bad-uuid", saga_operation.OperationReserve); err != saga_operation.ErrInvalidSagaID {
		t.Fatalf("expected ErrInvalidSagaID, got %v", err)
	}
}

func TestSagaOperationRepository_Sqlmock_FindBySagaAndOperation_Success(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewSagaOperationRepository(db)

	sagaID := uuid.New()
	mock.ExpectQuery(`SELECT .* FROM "saga_operations" WHERE saga_id`).
		WillReturnRows(buildSagaRow(uuid.New(), sagaID, uuid.New()))

	op, err := repo.FindBySagaAndOperation(context.Background(), sagaID.String(), saga_operation.OperationReserve)
	if err != nil || op == nil {
		t.Fatalf("expected operation, got err=%v", err)
	}
}

func TestSagaOperationRepository_Sqlmock_FindBySagaAndOperation_NotFound(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewSagaOperationRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "saga_operations" WHERE saga_id`).WillReturnError(gorm.ErrRecordNotFound)

	if _, err := repo.FindBySagaAndOperation(context.Background(), uuid.New().String(), saga_operation.OperationReserve); err != saga_operation.ErrSagaOperationNotFound {
		t.Fatalf("expected ErrSagaOperationNotFound, got %v", err)
	}
}

func TestSagaOperationRepository_Sqlmock_FindBySagaAndOperation_DBError(t *testing.T) {
	db, mock, cleanup := setupInventoryMockDB(t)
	defer cleanup()
	repo := NewSagaOperationRepository(db)

	mock.ExpectQuery(`SELECT .* FROM "saga_operations" WHERE saga_id`).WillReturnError(errors.New("db error"))

	if _, err := repo.FindBySagaAndOperation(context.Background(), uuid.New().String(), saga_operation.OperationReserve); err == nil {
		t.Fatal("expected error, got nil")
	}
}
