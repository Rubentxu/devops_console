import React, { useState } from "react";
import { useTenantStore } from "../store/tenantStore";

export const TenantForm: React.FC = () => {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const createTenant = useTenantStore((state) => state.createTenant);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    createTenant({ name, description });
    setName("");
    setDescription("");
  };

  return (
    <form onSubmit={handleSubmit} className="mb-4">
      <input
        type="text"
        value={name}
        onChange={(e) => setName(e.target.value)}
        placeholder="Nombre"
        className="mr-2 p-2 border rounded"
        required
      />
      <input
        type="text"
        value={description}
        onChange={(e) => setDescription(e.target.value)}
        placeholder="Descripción"
        className="mr-2 p-2 border rounded"
        required
      />
      <button type="submit" className="p-2 bg-blue-500 text-white rounded">
        Crear Tenant
      </button>
    </form>
  );
};
