package application

import (
    "devops_console/internal/domain/tenant"
)

type TenantService struct {
    tenantRepository tenant.TenantRepository
}

func NewTenantService(tenantRepository tenant.TenantRepository) *TenantService {
    return &TenantService{tenantRepository: tenantRepository}
}

func (s *TenantService) CreateTenant(tenantCreate tenant.TenantCreate) (*tenant.Tenant, error) {
    return s.tenantRepository.Create(tenantCreate)
}

func (s *TenantService) GetAllTenants() ([]*tenant.Tenant, error) {
    return s.tenantRepository.GetAll()
}

func (s *TenantService) GetTenantByID(tenantID string) (*tenant.Tenant, error) {
    return s.tenantRepository.GetByID(tenantID)
}

func (s *TenantService) UpdateTenant(tenantID string, tenantUpdate tenant.TenantUpdate) (*tenant.Tenant, error) {
    return s.tenantRepository.Update(tenantID, tenantUpdate)
}

func (s *TenantService) DeleteTenant(tenantID string) error {
    return s.tenantRepository.Delete(tenantID)
}