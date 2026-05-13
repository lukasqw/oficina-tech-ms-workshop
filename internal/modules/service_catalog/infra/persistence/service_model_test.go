package persistence

import (
	"errors"
	"github.com/lukasqw/oficina-tech-ms3-workshop/internal/modules/service_catalog/domain/service"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestServiceModel_ToDomain(t *testing.T) {
	validUUID := uuid.New()
	now := time.Now()
	deletedTime := now.Add(24 * time.Hour)

	tests := []struct {
		name    string
		model   ServiceModel
		wantErr error
		check   func(*testing.T, *service.Service)
	}{
		{
			name: "conversão de modelo sem DeletedAt",
			model: ServiceModel{
				ID:          validUUID,
				Name:        "Troca de Óleo",
				Description: "Troca de óleo do motor com filtro",
				Price:       15000,
				CreatedAt:   now,
				UpdatedAt:   now,
				DeletedAt:   gorm.DeletedAt{Valid: false},
			},
			wantErr: nil,
			check: func(t *testing.T, s *service.Service) {
				if s.ID() != validUUID.String() {
					t.Errorf("ID: got %v, want %v", s.ID(), validUUID.String())
				}
				if s.Name() != "Troca de Óleo" {
					t.Errorf("Name: got %v, want %v", s.Name(), "Troca de Óleo")
				}
				if s.Description() != "Troca de óleo do motor com filtro" {
					t.Errorf("Description: got %v, want %v", s.Description(), "Troca de óleo do motor com filtro")
				}
				if s.Price() != 15000 {
					t.Errorf("Price: got %v, want %v", s.Price(), 15000)
				}
				if !s.CreatedAt().Equal(now) {
					t.Errorf("CreatedAt: got %v, want %v", s.CreatedAt(), now)
				}
				if !s.UpdatedAt().Equal(now) {
					t.Errorf("UpdatedAt: got %v, want %v", s.UpdatedAt(), now)
				}
				if s.DeletedAt() != nil {
					t.Errorf("DeletedAt: got %v, want nil", s.DeletedAt())
				}
			},
		},
		{
			name: "conversão de modelo com DeletedAt válido",
			model: ServiceModel{
				ID:          validUUID,
				Name:        "Alinhamento",
				Description: "Alinhamento e balanceamento das rodas",
				Price:       8000,
				CreatedAt:   now,
				UpdatedAt:   now,
				DeletedAt: gorm.DeletedAt{
					Time:  deletedTime,
					Valid: true,
				},
			},
			wantErr: nil,
			check: func(t *testing.T, s *service.Service) {
				if s.ID() != validUUID.String() {
					t.Errorf("ID: got %v, want %v", s.ID(), validUUID.String())
				}
				if s.Name() != "Alinhamento" {
					t.Errorf("Name: got %v, want %v", s.Name(), "Alinhamento")
				}
				if s.Description() != "Alinhamento e balanceamento das rodas" {
					t.Errorf("Description: got %v, want %v", s.Description(), "Alinhamento e balanceamento das rodas")
				}
				if s.Price() != 8000 {
					t.Errorf("Price: got %v, want %v", s.Price(), 8000)
				}
				if !s.CreatedAt().Equal(now) {
					t.Errorf("CreatedAt: got %v, want %v", s.CreatedAt(), now)
				}
				if !s.UpdatedAt().Equal(now) {
					t.Errorf("UpdatedAt: got %v, want %v", s.UpdatedAt(), now)
				}
				if s.DeletedAt() == nil {
					t.Fatal("DeletedAt: got nil, want non-nil")
				}
				if !s.DeletedAt().Equal(deletedTime) {
					t.Errorf("DeletedAt: got %v, want %v", s.DeletedAt(), deletedTime)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.model.ToDomain()

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got == nil {
				t.Fatal("expected non-nil result")
			}

			tt.check(t, got)
		})
	}
}

func TestFromDomain(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	now := time.Now()
	deletedTime := now.Add(24 * time.Hour)

	tests := []struct {
		name    string
		service func() *service.Service
		wantErr error
		check   func(*testing.T, *ServiceModel)
	}{
		{
			name: "conversão de entidade com ID preenchido (UUID válido)",
			service: func() *service.Service {
				s, _ := service.ReconstructService(
					validUUID,
					"Troca de Óleo",
					"Troca de óleo do motor com filtro",
					15000,
					now,
					now,
					nil,
				)
				return s
			},
			wantErr: nil,
			check: func(t *testing.T, m *ServiceModel) {
				if m.ID.String() != validUUID {
					t.Errorf("ID: got %v, want %v", m.ID.String(), validUUID)
				}
				if m.Name != "Troca de Óleo" {
					t.Errorf("Name: got %v, want %v", m.Name, "Troca de Óleo")
				}
				if m.Description != "Troca de óleo do motor com filtro" {
					t.Errorf("Description: got %v, want %v", m.Description, "Troca de óleo do motor com filtro")
				}
				if m.Price != 15000 {
					t.Errorf("Price: got %v, want %v", m.Price, 15000)
				}
				if !m.CreatedAt.Equal(now) {
					t.Errorf("CreatedAt: got %v, want %v", m.CreatedAt, now)
				}
				if !m.UpdatedAt.Equal(now) {
					t.Errorf("UpdatedAt: got %v, want %v", m.UpdatedAt, now)
				}
				if m.DeletedAt.Valid {
					t.Errorf("DeletedAt.Valid: got true, want false")
				}
			},
		},
		{
			name: "conversão de entidade com ID vazio (novo serviço)",
			service: func() *service.Service {
				s, _ := service.NewService(
					"Alinhamento",
					"Alinhamento e balanceamento das rodas",
					8000,
				)
				return s
			},
			wantErr: nil,
			check: func(t *testing.T, m *ServiceModel) {
				if m.ID != uuid.Nil {
					t.Errorf("ID: got %v, want zero UUID", m.ID)
				}
				if m.Name != "Alinhamento" {
					t.Errorf("Name: got %v, want %v", m.Name, "Alinhamento")
				}
				if m.Description != "Alinhamento e balanceamento das rodas" {
					t.Errorf("Description: got %v, want %v", m.Description, "Alinhamento e balanceamento das rodas")
				}
				if m.Price != 8000 {
					t.Errorf("Price: got %v, want %v", m.Price, 8000)
				}
				if m.DeletedAt.Valid {
					t.Errorf("DeletedAt.Valid: got true, want false")
				}
			},
		},
		{
			name: "conversão de entidade com deletedAt preenchido",
			service: func() *service.Service {
				s, _ := service.ReconstructService(
					validUUID,
					"Revisão",
					"Revisão completa do veículo",
					25000,
					now,
					now,
					&deletedTime,
				)
				return s
			},
			wantErr: nil,
			check: func(t *testing.T, m *ServiceModel) {
				if m.ID.String() != validUUID {
					t.Errorf("ID: got %v, want %v", m.ID.String(), validUUID)
				}
				if m.Name != "Revisão" {
					t.Errorf("Name: got %v, want %v", m.Name, "Revisão")
				}
				if !m.DeletedAt.Valid {
					t.Fatal("DeletedAt.Valid: got false, want true")
				}
				if !m.DeletedAt.Time.Equal(deletedTime) {
					t.Errorf("DeletedAt.Time: got %v, want %v", m.DeletedAt.Time, deletedTime)
				}
			},
		},
		{
			name: "conversão de entidade sem deletedAt (nil)",
			service: func() *service.Service {
				s, _ := service.ReconstructService(
					validUUID,
					"Pintura",
					"Pintura completa do veículo",
					50000,
					now,
					now,
					nil,
				)
				return s
			},
			wantErr: nil,
			check: func(t *testing.T, m *ServiceModel) {
				if m.ID.String() != validUUID {
					t.Errorf("ID: got %v, want %v", m.ID.String(), validUUID)
				}
				if m.Name != "Pintura" {
					t.Errorf("Name: got %v, want %v", m.Name, "Pintura")
				}
				if m.DeletedAt.Valid {
					t.Errorf("DeletedAt.Valid: got true, want false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.service()
			got, err := FromDomain(s)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.wantErr)
					return
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got == nil {
				t.Fatal("expected non-nil result")
			}

			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}
