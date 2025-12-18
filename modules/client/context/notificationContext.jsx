"use client";
import { createContext, useContext, useState, useCallback, useEffect } from "react";
import { getNotificationCount } from "../api";

const initialState = {
  unreadCount: 0,
  loading: false,
};

export const NotificationContext = createContext({
  ...initialState,
  fetchCount: () => {},
  decrementCount: () => {},
  setUnreadCount: () => {},
});

export const useNotification = () => useContext(NotificationContext);

export const NotificationProvider = ({ children }) => {
  const [unreadCount, setUnreadCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const [hasFetched, setHasFetched] = useState(false);

  const fetchCount = useCallback(async () => {
    if (loading) return;

    setLoading(true);
    try {
      const data = await getNotificationCount();
      setUnreadCount(data?.unread_count || 0);
      setHasFetched(true);
    } catch {
      // Silently fail - notifications aren't critical
    } finally {
      setLoading(false);
    }
  }, [loading]);

  const decrementCount = useCallback((amount = 1) => {
    setUnreadCount((prev) => Math.max(0, prev - amount));
  }, []);

  // Fetch on initial mount
  useEffect(() => {
    if (hasFetched) return;
    fetchCount();
  }, [hasFetched, fetchCount]);

  // Refresh when tab becomes visible
  useEffect(() => {
    const handleVisibility = () => {
      if (document.visibilityState === "visible") {
        fetchCount();
      }
    };
    document.addEventListener("visibilitychange", handleVisibility);
    return () => document.removeEventListener("visibilitychange", handleVisibility);
  }, [fetchCount]);

  return (
    <NotificationContext.Provider value={{ unreadCount, loading, fetchCount, decrementCount, setUnreadCount }}>
      {children}
    </NotificationContext.Provider>
  );
};
