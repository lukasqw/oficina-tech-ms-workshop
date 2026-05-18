package observability

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupOtelMockDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	_ = mock
	dialector := postgres.New(postgres.Config{Conn: sqlDB, DriverName: "postgres"})
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	return db, func() { _ = sqlDB.Close() }
}

func TestOTelPlugin_NewAndName(t *testing.T) {
	p := NewOTelPlugin()
	if p == nil {
		t.Fatal("expected non-nil OTelPlugin")
	}
	if p.Name() != "otel:gorm" {
		t.Errorf("Name() = %v, want otel:gorm", p.Name())
	}
}

func TestOTelPlugin_Initialize(t *testing.T) {
	db, cleanup := setupOtelMockDB(t)
	defer cleanup()

	p := NewOTelPlugin()
	if err := p.Initialize(db); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}
}

func TestBefore_NilStatement(t *testing.T) {
	db, cleanup := setupOtelMockDB(t)
	defer cleanup()
	// Statement is nil on a fresh DB — before must return without panic
	before("create")(db)
}

func TestAfter_NilStatement(t *testing.T) {
	db, cleanup := setupOtelMockDB(t)
	defer cleanup()
	// Statement is nil on a fresh DB — after must return without panic
	after("query")(db)
}
