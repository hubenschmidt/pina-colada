"use client";

import { useState, useEffect, useId } from "react";
import { useRouter } from "next/navigation";
import { Inbox } from "lucide-react";
import { getNotifications, markNotificationsRead } from "../../api";
import NotificationDropdown from "../NotificationDropdown/NotificationDropdown";
import { useProjectContext } from "../../context/projectContext";
import { useNotification } from "../../context/notificationContext";

const NotificationBell = () => {
  const router = useRouter();
  const { projectState, selectProject } = useProjectContext();
  const { unreadCount, decrementCount, setUnreadCount } = useNotification();
  const [notifications, setNotifications] = useState([]);
  const [isOpen, setIsOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const dropdownId = useId();

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event) => {
      const dropdown = document.getElementById(dropdownId);
      if (dropdown && !dropdown.contains(event.target)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isOpen, dropdownId]);

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

  const handleNotificationClick = async (notification) => {
    // Mark as read
    if (!notification.is_read) {
      try {
        await markNotificationsRead([notification.id]);
        setNotifications((prev) =>
          prev.map((n) => (n.id === notification.id ? { ...n, is_read: true } : n))
        );
        decrementCount(1);
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
      decrementCount(unreadIds.length);
    } catch (err) {
      console.error("Failed to mark all as read:", err);
    }
  };

  return (
    <div id={dropdownId} className="relative mt-1">
      <button
        onClick={handleOpen}
        className="relative p-2 text-zinc-600 dark:text-zinc-300 hover:text-zinc-900 dark:hover:text-zinc-100 hover:bg-zinc-100 dark:hover:bg-zinc-800 rounded-full transition-colors"
        aria-label="Notifications">
        <Inbox size={20} />
        {unreadCount > 0 && (
          <span className="absolute -top-0.5 -right-0.5 bg-pink-500 text-white text-[10px] font-medium rounded-full min-w-[18px] h-[18px] flex items-center justify-center px-1">
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
