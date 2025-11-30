"use client";

import { useState, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import { Bell } from "lucide-react";
import {
  Notification,
  getNotificationCount,
  getNotifications,
  markNotificationsRead,
} from "../../api";
import NotificationDropdown from "../NotificationDropdown/NotificationDropdown";
import { useProjectContext } from "../../context/projectContext";

const NotificationBell = () => {
  const router = useRouter();
  const { projectState, selectProject } = useProjectContext();
  const [unreadCount, setUnreadCount] = useState(0);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [isOpen, setIsOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const bellRef = useRef<HTMLDivElement>(null);

  // Poll for unread count every 30 seconds
  useEffect(() => {
    const fetchCount = async () => {
      try {
        const { unread_count } = await getNotificationCount();
        setUnreadCount(unread_count);
      } catch (err) {
        console.error("Failed to fetch notification count:", err);
      }
    };

    fetchCount();
    const interval = setInterval(fetchCount, 30000);
    return () => clearInterval(interval);
  }, []);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (bellRef.current && !bellRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isOpen]);

  const handleOpen = async () => {
    if (isOpen) {
      setIsOpen(false);
      return;
    }

    setIsOpen(true);
    setIsLoading(true);

    try {
      const data = await getNotifications(20);
      setNotifications(data.notifications);
      setUnreadCount(data.unread_count);
    } catch (err) {
      console.error("Failed to fetch notifications:", err);
    } finally {
      setIsLoading(false);
    }
  };

  const handleNotificationClick = async (notification: Notification) => {
    // Mark as read
    if (!notification.is_read) {
      try {
        await markNotificationsRead([notification.id]);
        setNotifications((prev) =>
          prev.map((n) => (n.id === notification.id ? { ...n, is_read: true } : n))
        );
        setUnreadCount((prev) => Math.max(0, prev - 1));
      } catch (err) {
        console.error("Failed to mark notification as read:", err);
      }
    }

    // Switch project scope if needed
    const entityProjectId = notification.entity?.project_id;
    const currentProjectId = projectState.selectedProject?.id;
    const needsProjectSwitch = entityProjectId && currentProjectId !== entityProjectId;
    const targetProject = needsProjectSwitch
      ? projectState.projects.find((p) => p.id === entityProjectId)
      : null;

    if (targetProject) {
      selectProject(targetProject);
    }
    if (!entityProjectId && projectState.selectedProject) {
      selectProject(null);
    }

    // Navigate to entity
    if (notification.entity?.url) {
      router.push(notification.entity.url);
    }

    setIsOpen(false);
  };

  const handleMarkAllRead = async () => {
    const unreadIds = notifications.filter((n) => !n.is_read).map((n) => n.id);
    if (unreadIds.length === 0) return;

    try {
      await markNotificationsRead(unreadIds);
      setNotifications((prev) => prev.map((n) => ({ ...n, is_read: true })));
      setUnreadCount(0);
    } catch (err) {
      console.error("Failed to mark all as read:", err);
    }
  };

  return (
    <div ref={bellRef} className="relative mt-1">
      <button
        onClick={handleOpen}
        className="relative p-2 text-zinc-600 dark:text-zinc-300 hover:text-zinc-900 dark:hover:text-zinc-100 hover:bg-zinc-100 dark:hover:bg-zinc-800 rounded-full transition-colors"
        aria-label="Notifications"
      >
        <Bell size={20} />
        {unreadCount > 0 && (
          <span className="absolute -top-0.5 -right-0.5 bg-red-500 text-white text-[10px] font-medium rounded-full min-w-[18px] h-[18px] flex items-center justify-center px-1">
            {unreadCount > 99 ? "99+" : unreadCount}
          </span>
        )}
      </button>

      {isOpen && (
        <NotificationDropdown
          notifications={notifications}
          isLoading={isLoading}
          onNotificationClick={handleNotificationClick}
          onMarkAllRead={handleMarkAllRead}
          onClose={() => setIsOpen(false)}
        />
      )}
    </div>
  );
};

export default NotificationBell;
