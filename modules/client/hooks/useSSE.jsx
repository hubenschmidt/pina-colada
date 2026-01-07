"use client";

import { useEffect, useState, useCallback } from "react";
import { fetchBearerToken } from "../lib/fetch-bearer-token";

const SSE_STATES = { CONNECTING: 0, OPEN: 1, CLOSED: 2 };

const parseLine = (line, currentEvent) => {
  if (line.startsWith("event:")) return { ...currentEvent, type: line.slice(6).trim() };
  if (line.startsWith("data:")) return { ...currentEvent, data: currentEvent.data + line.slice(5).trim() };
  if (line.startsWith("id:")) return { ...currentEvent, id: line.slice(3).trim() };
  return currentEvent;
};

const parseSSEBuffer = (buffer) => {
  const lines = buffer.split("\n");
  const events = [];
  let currentEvent = { type: "message", data: "" };
  let lastCompleteIndex = -1;
  let hasKeepAlive = false;

  lines.forEach((line, i) => {
    if (line === "" && currentEvent.data) {
      events.push({ ...currentEvent });
      currentEvent = { type: "message", data: "" };
      lastCompleteIndex = i;
      return;
    }

    if (line === "") {
      lastCompleteIndex = i;
      return;
    }

    if (line.startsWith(":")) {
      hasKeepAlive = true;
      lastCompleteIndex = i;
      return;
    }

    currentEvent = parseLine(line, currentEvent);
  });

  const remaining = lastCompleteIndex < lines.length - 1
    ? lines.slice(lastCompleteIndex + 1).join("\n")
    : "";

  return { parsed: events, remaining, hasKeepAlive };
};

const getSSEBaseUrl = () => {
  if (typeof window === "undefined") return "";
  return process.env.NEXT_PUBLIC_SSE_URL || "http://localhost:8000";
};

const parseEventData = (data) => {
  try {
    return JSON.parse(data);
  } catch {
    return data;
  }
};

/**
 * Generic SSE hook for real-time data streaming
 */
export const useSSE = (url, { enabled = true, onEvent = null } = {}) => {
  const [state, setState] = useState(SSE_STATES.CLOSED);
  const [data, setData] = useState(null);
  const [error, setError] = useState(null);
  const [lastPulse, setLastPulse] = useState(null);
  const [reconnectTrigger, setReconnectTrigger] = useState(0);

  const processEvent = useCallback((event) => {
    const parsedData = parseEventData(event.data);

    if (event.type === "init") {
      setData(parsedData);
      return;
    }

    if (event.type === "error") {
      setError(parsedData?.error || "Unknown error");
      return;
    }

    onEvent?.(event.type, parsedData);
  }, [onEvent]);

  useEffect(() => {
    if (!enabled) {
      setState(SSE_STATES.CLOSED);
      setData(null);
      return;
    }

    const abortController = new AbortController();
    let buffer = "";

    const readChunk = async (reader, decoder) => {
      const { value, done } = await reader.read();
      if (done) return false;

      buffer += decoder.decode(value, { stream: true });
      const { parsed, remaining, hasKeepAlive } = parseSSEBuffer(buffer);
      buffer = remaining;

      if (hasKeepAlive) setLastPulse(Date.now());
      parsed.forEach(processEvent);

      return true;
    };

    let reconnectTimeout = null;

    const scheduleReconnect = () => {
      if (abortController.signal.aborted) return;
      reconnectTimeout = setTimeout(() => setReconnectTrigger((n) => n + 1), 2000);
    };

    const connect = async () => {
      setState(SSE_STATES.CONNECTING);
      setError(null);

      const { headers } = await fetchBearerToken();
      const fullUrl = `${getSSEBaseUrl()}${url}`;
      const response = await fetch(fullUrl, {
        headers,
        credentials: "include",
        signal: abortController.signal,
      });

      if (!response.ok) throw new Error(`HTTP ${response.status}`);

      setState(SSE_STATES.OPEN);

      const reader = response.body.getReader();
      const decoder = new TextDecoder();

      while (await readChunk(reader, decoder)) {
        // Continue reading
      }

      // Stream closed unexpectedly - auto-reconnect
      setState(SSE_STATES.CLOSED);
      scheduleReconnect();
    };

    connect().catch((err) => {
      if (err.name === "AbortError") {
        setState(SSE_STATES.CLOSED);
        return;
      }
      setError(err.message);
      setState(SSE_STATES.CLOSED);
      scheduleReconnect();
    });

    return () => {
      if (reconnectTimeout) clearTimeout(reconnectTimeout);
      abortController.abort();
      setState(SSE_STATES.CLOSED);
    };
  }, [url, enabled, processEvent, reconnectTrigger]);

  const reconnect = useCallback(() => {
    setReconnectTrigger((n) => n + 1);
  }, []);

  return {
    data,
    error,
    isConnecting: state === SSE_STATES.CONNECTING,
    isConnected: state === SSE_STATES.OPEN,
    lastPulse,
    reconnect,
  };
};

export default useSSE;
