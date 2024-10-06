from fastapi import APIRouter, HTTPException
from src.domain.workspace.workspace import Workspace, WorkspaceCreate, WorkspaceUpdate
from src.application.workspace_service import WorkspaceService
from typing import List

def create_workspace_router(workspace_service: WorkspaceService):
    router = APIRouter()

    @router.get('/workspaces', response_model=List[Workspace])
    def get_workspaces():
        return workspace_service.get_all_workspaces()

    @router.post('/workspaces', response_model=Workspace)
    def create_workspace(workspace: WorkspaceCreate):
        return workspace_service.create_workspace(workspace)

    @router.get('/workspaces/{workspace_id}', response_model=Workspace)
    def get_workspace(workspace_id: str):
        workspace = workspace_service.get_workspace_by_id(workspace_id)
        if workspace is None:
            raise HTTPException(status_code=404, detail='Workspace not found')
        return workspace

    @router.put('/workspaces/{workspace_id}', response_model=Workspace)
    def update_workspace(workspace_id: str, workspace_update: WorkspaceUpdate):
        updated_workspace = workspace_service.update_workspace(workspace_id, workspace_update)
        if updated_workspace is None:
            raise HTTPException(status_code=404, detail='Workspace not found')
        return updated_workspace

    @router.delete('/workspaces/{workspace_id}', response_model=dict)
    def delete_workspace(workspace_id: str):
        deleted = workspace_service.delete_workspace(workspace_id)
        if not deleted:
            raise HTTPException(status_code=404, detail='Workspace not found')
        return {"message": "Workspace deleted successfully"}

    return router
