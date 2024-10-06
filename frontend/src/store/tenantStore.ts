import { create } from "zustand";
import { Tenant, TenantCreate } from "../types/tenantTypes";
import { config } from "../config";

export interface TenantStore {
  tenants: Tenant[];
  currentTenant: Tenant | null;
  fetchTenants: () => Promise<void>;
  createTenant: (tenant: TenantCreate) => Promise<void>;
  updateTenant: (id: string, tenant: TenantCreate) => Promise<void>;
  deleteTenant: (id: string) => Promise<void>;
}

export const useTenantStore = create<TenantStore>((set) => ({
  tenants: [],
  currentTenant: null,
  fetchTenants: async () => {
    const response = await fetch(`${config.apiUrl}/tenants`);
    const tenants = await response.json();
    set({ tenants });
  },
  createTenant: async (tenant) => {
    const response = await fetch(`${config.apiUrl}/tenants`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(tenant),
    });
    const newTenant = await response.json();
    set((state) => ({ tenants: [...state.tenants, newTenant] }));
  },
  updateTenant: async (id, tenant) => {
    const response = await fetch(`${config.apiUrl}/tenants/${id}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(tenant),
    });
    const updatedTenant = await response.json();
    set((state) => ({
      tenants: state.tenants.map((t) => (t.id === id ? updatedTenant : t)),
    }));
  },
  deleteTenant: async (id) => {
    await fetch(`${config.apiUrl}/tenants/${id}`, { method: "DELETE" });
    set((state) => ({
      tenants: state.tenants.filter((t) => t.id !== id),
    }));
  },
}));
