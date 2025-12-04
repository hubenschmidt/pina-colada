"use client";

import { Reply, MessageCircle, Check } from "lucide-react";










const NotificationDropdown = ({
  notifications,
  isLoading,
  onNotificationClick,
  onMarkAllRead,
  onClose
}) => {
  const formatRelativeTime = (dateStr) => {
    if (!dateStr) return "";
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return "just now";
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return date.toLocaleDateString("en-US", { month: "short", day: "numeric" });
  };

  const getNotificationIcon = (type) => {
    if (type === "direct_reply") {
      return <Reply size={14} className="text-lime-500" />;
    }
    return <MessageCircle size={14} className="text-blue-500" />;
  };

  const getNotificationLabel = (notification) => {
    const name = notification.comment?.created_by_name || "Someone";
    if (notification.type === "direct_reply") {
      return `${name} replied to your comment`;
    }
    return `${name} commented`;
  };

  const hasUnread = notifications.some((n) => !n.is_read);

  return (
    <div className="absolute right-0 top-full mt-2 w-80 sm:w-96 bg-white dark:bg-zinc-900 rounded-lg shadow-lg border border-zinc-200 dark:border-zinc-700 z-50 overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-zinc-100 dark:border-zinc-800">
        <h3 className="text-sm font-semibold text-zinc-900 dark:text-zinc-100">
          Notifications
        </h3>
        {hasUnread &&
        <button
          onClick={onMarkAllRead}
          className="flex items-center gap-1 text-xs text-lime-600 hover:text-lime-700 dark:text-lime-400 dark:hover:text-lime-300 transition-colors">

            <Check size={12} />
            Mark all read
          </button>
        }
      </div>

      {/* Content */}
      <div className="max-h-[400px] overflow-y-auto">
        {isLoading ?
        <div className="flex items-center justify-center py-8">
            <div className="w-5 h-5 border-2 border-zinc-300 border-t-zinc-600 rounded-full animate-spin" />
          </div> :
        notifications.length === 0 ?
        <div className="py-8 text-center">
            <MessageCircle size={24} className="mx-auto text-zinc-300 dark:text-zinc-600 mb-2" />
            <p className="text-sm text-zinc-500 dark:text-zinc-400">
              No notifications yet
            </p>
          </div> :

        <ul className="divide-y divide-zinc-100 dark:divide-zinc-800">
            {notifications.map((notification) =>
          <li key={notification.id}>
                <button
              onClick={(e) => {
                e.stopPropagation();
                onNotificationClick(notification);
              }}
              className={`w-full px-4 py-3 text-left hover:bg-zinc-50 dark:hover:bg-zinc-800 transition-colors ${
              !notification.is_read ? "bg-lime-50/50 dark:bg-lime-900/10" : ""}`
              }>

                  <div className="flex gap-3">
                    {/* Icon */}
                    <div className="flex-shrink-0 mt-0.5">
                      {getNotificationIcon(notification.type)}
                    </div>

                    {/* Content */}
                    <div className="flex-1 min-w-0">
                      {/* Entity type badge */}
                      {notification.entity?.type &&
                  <div className="flex justify-end mb-1">
                          <span className="text-[10px] font-medium text-zinc-400 dark:text-zinc-500 uppercase tracking-wide">
                            {notification.entity.type}
                          </span>
                        </div>
                  }
                      {/* Label */}
                      <p className="text-sm text-zinc-900 dark:text-zinc-100">
                        {getNotificationLabel(notification)}
                        {notification.entity?.display_name &&
                    <span className="text-zinc-500 dark:text-zinc-400">
                            {" "}on{" "}
                            <span className="text-zinc-700 dark:text-zinc-300">
                              {notification.entity.display_name}
                            </span>
                          </span>
                    }
                      </p>

                      {/* Comment preview */}
                      {notification.comment?.content &&
                  <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400 line-clamp-2">
                          "{notification.comment.content}"
                        </p>
                  }

                      {/* Time */}
                      <p className="mt-1 text-xs text-zinc-400 dark:text-zinc-500">
                        {formatRelativeTime(notification.created_at)}
                      </p>
                    </div>

                    {/* Unread indicator */}
                    {!notification.is_read &&
                <div className="flex-shrink-0">
                        <div className="w-2 h-2 bg-lime-500 rounded-full" />
                      </div>
                }
                  </div>
                </button>
              </li>
          )}
          </ul>
        }
      </div>
    </div>);

};

export default NotificationDropdown;