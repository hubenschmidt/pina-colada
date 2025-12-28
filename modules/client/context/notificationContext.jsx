"use client";
import { createContext, useContext, useState, useCallback, useEffect } from "react";
import { getNotificationCount } from "../api";
import { useUserContext } from "./userContext";

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
  const { userState } = useUserContext();
  const [unreadCount, setUnreadCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const [hasFetched, setHasFetched] = useState(false);

  const fetchCount = useCallback(async () => {
    if (loading || !userState.isAuthed) return;

    setLoading(true);
    try {
      const data = await getNotificationCount();
      setUnreadCount(data?.unread_count || 0);
    } catch {
      // Silently fail - notifications aren't critical
    } finally {
      setHasFetched(true);
      setLoading(false);
    }
  }, [loading, userState.isAuthed]);

  const decrementCount = useCallback((amount = 1) => {
    setUnreadCount((prev) => Math.max(0, prev - amount));
  }, []);

  // Fetch when user becomes authenticated
  useEffect(() => {
    if (!userState.isAuthed || hasFetched) return;
    fetchCount();
  }, [userState.isAuthed, hasFetched, fetchCount]);

  // Refresh when tab becomes visible (only if authenticated)
  useEffect(() => {
    const handleVisibility = () => {
      if (document.visibilityState === "visible" && userState.isAuthed) {
        fetchCount();
      }
    };
    document.addEventListener("visibilitychange", handleVisibility);
    return () => document.removeEventListener("visibilitychange", handleVisibility);
  }, [fetchCount, userState.isAuthed]);

  return (
    <NotificationContext.Provider value={{ unreadCount, loading, fetchCount, decrementCount, setUnreadCount }}>
      {children}
    </NotificationContext.Provider>
  );
};
