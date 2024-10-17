package application

import (
    "devops_console/internal/domain/workspace"
)

type WorkspaceService struct {
    workspaceRepository workspace.WorkspaceRepository
}

func NewWorkspaceService(workspaceRepository workspace.WorkspaceRepository) *WorkspaceService {
    return &WorkspaceService{workspaceRepository: workspaceRepository}
}

func (s *WorkspaceService) CreateWorkspace(workspaceCreate workspace.WorkspaceCreate) (*workspace.Workspace, error) {
    return s.workspaceRepository.Create(workspaceCreate)
}

func (s *WorkspaceService) GetAllWorkspaces() ([]*workspace.Workspace, error) {
    return s.workspaceRepository.GetAll()
}

func (s *WorkspaceService) GetWorkspaceByID(workspaceID string) (*workspace.Workspace, error) {
    return s.workspaceRepository.GetByID(workspaceID)
}

func (s *WorkspaceService) UpdateWorkspace(workspaceID string, workspaceUpdate workspace.WorkspaceUpdate) (*workspace.Workspace, error) {
    return s.workspaceRepository.Update(workspaceID, workspaceUpdate)
}

func (s *WorkspaceService) DeleteWorkspace(workspaceID string) error {
    return s.workspaceRepository.Delete(workspaceID)
}