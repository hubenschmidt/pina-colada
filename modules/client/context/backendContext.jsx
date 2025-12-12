"use client";

import { createContext, useState, useContext, useEffect } from "react";

const STORAGE_KEY = "pina-colada-backend";
const DEFAULT_BACKEND = "agent";

const BackendContext = createContext({
  backend: DEFAULT_BACKEND,
  setBackend: () => {},
});

export const useBackendContext = () => useContext(BackendContext);

export const BackendProvider = ({ children }) => {
  const [backend, setBackendState] = useState(DEFAULT_BACKEND);

  useEffect(() => {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      setBackendState(stored);
    }
  }, []);

  const setBackend = (value) => {
    setBackendState(value);
    localStorage.setItem(STORAGE_KEY, value);
  };

  return (
    <BackendContext.Provider value={{ backend, setBackend }}>
      {children}
    </BackendContext.Provider>
  );
};
