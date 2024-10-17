import React from "react";
import { Tenant } from "../types/tenant";
import { useTenantStore } from "../store/tenantStore";

interface TenantListProps {
  tenants: Tenant[];
}

const TenantList: React.FC<TenantListProps> = ({ tenants }) => {
  const deleteTenant = useTenantStore((state) => state.deleteTenant);

  return (
    <ul className="space-y-2">
      {tenants.map((tenant) => (
        <li key={tenant.id} className="border p-2 rounded flex justify-between">
          <div>
            <h3 className="font-bold">{tenant.name}</h3>
            <p>{tenant.description}</p>
          </div>
          <button
            onClick={() => deleteTenant(tenant.id)}
            className="text-red-500"
          >
            Eliminar
          </button>
        </li>
      ))}
    </ul>
  );
};

export default TenantList;
