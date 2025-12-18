"use client";
import { createContext, useContext, useState, useCallback } from "react";
import { getAgentConfig } from "../api";

const initialState = {
  config: null,
  loading: false,
  error: null,
  isRecording: false,
};

export const AgentConfigContext = createContext({
  ...initialState,
  fetchConfig: () => {},
  refetchConfig: () => {},
  setConfig: () => {},
  setIsRecording: () => {},
});

export const useAgentConfig = () => useContext(AgentConfigContext);

export const AgentConfigProvider = ({ children }) => {
  const [config, setConfig] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [hasFetched, setHasFetched] = useState(false);
  const [isRecording, setIsRecording] = useState(false);

  const fetchConfig = useCallback(async () => {
    if (hasFetched || loading) return config;

    setLoading(true);
    setError(null);

    try {
      const data = await getAgentConfig();
      setConfig(data);
      setHasFetched(true);
      return data;
    } catch (err) {
      setError(err.message);
      return null;
    } finally {
      setLoading(false);
    }
  }, [hasFetched, loading, config]);

  const refetchConfig = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const data = await getAgentConfig();
      setConfig(data);
      setHasFetched(true);
      return data;
    } catch (err) {
      setError(err.message);
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  return (
    <AgentConfigContext.Provider
      value={{ config, loading, error, fetchConfig, refetchConfig, setConfig, isRecording, setIsRecording }}>
      {children}
    </AgentConfigContext.Provider>
  );
};
