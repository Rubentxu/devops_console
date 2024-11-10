package entities

import (
	"time"
)

type Tenant struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type TenantCreate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TenantUpdate struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type TenantRepository interface {
	Create(tenant TenantCreate) (*Tenant, error)
	GetAll() ([]*Tenant, error)
	GetByID(tenantID string) (*Tenant, error)
	Update(tenantID string, tenantUpdate TenantUpdate) (*Tenant, error)
	Delete(tenantID string) error
}
