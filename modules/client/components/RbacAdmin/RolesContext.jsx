"use client";

import { createContext, useContext, useState, useEffect, useCallback } from "react";
import { getRoles } from "../../api";

const RolesContext = createContext(null);

export const RolesProvider = ({ children }) => {
  const [roles, setRoles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const loadRoles = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await getRoles();
      setRoles(data);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadRoles();
  }, [loadRoles]);

  const refreshRoles = useCallback(() => {
    return loadRoles();
  }, [loadRoles]);

  return (
    <RolesContext.Provider value={{ roles, loading, error, refreshRoles }}>
      {children}
    </RolesContext.Provider>
  );
};

export const useRoles = () => {
  const context = useContext(RolesContext);
  if (!context) {
    throw new Error("useRoles must be used within a RolesProvider");
  }
  return context;
};
