import React, { useEffect } from "react";
import { useTenantStore } from "../store/tenantStore";
import { TenantForm } from "../components/TenantForm";
import TenantList from "../components/TenantList";

export const Tenants: React.FC = () => {
  const { tenants, fetchTenants } = useTenantStore();

  useEffect(() => {
    fetchTenants();
  }, [fetchTenants]);

  return (
    <div className="p-4">
      <h1 className="text-2xl font-bold mb-4">Tenants</h1>
      <TenantForm />
      <TenantList tenants={tenants} />
    </div>
  );
};
