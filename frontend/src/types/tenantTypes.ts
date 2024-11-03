export interface Tenant {
  id: string;
  name: string;
  description: string;
  created_at: Date;
  updated_at: Date;
}

export interface TenantCreate {
  name: string;
  description: string;
}

export interface TenantUpdate {
  name?: string;
  description?: string;
}