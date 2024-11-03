package workspace

import (
	"time"
)

type Workspace struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	TenantID    int       `json:"tenant_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type WorkspaceCreate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	TenantID    int    `json:"tenant_id"`
}

type WorkspaceUpdate struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	TenantID    *int    `json:"tenant_id,omitempty"`
}

type WorkspaceRepository interface {
	Create(workspace WorkspaceCreate) (*Workspace, error)
	GetAll() ([]*Workspace, error)
	GetByID(workspaceID string) (*Workspace, error)
	Update(workspaceID string, workspaceUpdate WorkspaceUpdate) (*Workspace, error)
	Delete(workspaceID string) error
}