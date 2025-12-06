"use client";

import { useState } from "react";
import { useRouter, usePathname } from "next/navigation";
import { MessageSquare, Plus, Archive, ChevronDown, ChevronRight } from "lucide-react";
import { useConversationContext } from "../../context/conversationContext";
import styles from "./ChatHistory.module.css";

const MAX_SIDEBAR_CHATS = 3;

const truncateTitle = (title, maxLen = 28) => {
  if (!title) return "New conversation";
  if (title.length <= maxLen) return title;
  return title.slice(0, maxLen - 3) + "...";
};

export const ChatHistory = ({ collapsed = false }) => {
  const router = useRouter();
  const pathname = usePathname();
  const { conversationState, archiveConversation } = useConversationContext();

  const { conversations, isLoading } = conversationState;
  const [hoveredId, setHoveredId] = useState(null);
  const [isExpanded, setIsExpanded] = useState(true);

  // Extract current threadId from URL
  const currentThreadId = pathname?.startsWith("/chat/") ? pathname.split("/")[2] : null;

  const handleNewChat = () => {
    router.push("/chat");
  };

  const handleSelect = (threadId) => {
    router.push(`/chat/${threadId}`);
  };

  const handleArchive = async (e, threadId) => {
    e.stopPropagation();
    await archiveConversation(threadId);
    // If we archived the current conversation, go to new chat
    if (currentThreadId === threadId) {
      router.push("/chat");
    }
  };

  if (collapsed) {
    return (
      <div className={styles.collapsedContainer}>
        <button onClick={handleNewChat} className={styles.collapsedButton} title="New Chat">
          <Plus className="h-4 w-4" />
        </button>
      </div>
    );
  }

  // Only show first 3 conversations in sidebar
  const recentConversations = conversations.slice(0, MAX_SIDEBAR_CHATS);

  return (
    <div>
      <div
        className={`flex w-full items-center gap-2 rounded px-3 py-2 text-md text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 ${
          pathname === "/chats" ? "font-bold" : "font-normal"
        }`}>
        <button
          onClick={() => router.push("/chats")}
          className="flex items-center gap-2 cursor-pointer hover:text-zinc-900 dark:hover:text-zinc-100">
          <MessageSquare className="h-4 w-4 text-lime-600 dark:text-lime-400" />
          Chats
        </button>
        <span className="ml-auto flex items-center gap-1">
          <button
            onClick={handleNewChat}
            className="p-1 rounded hover:bg-zinc-200 dark:hover:bg-zinc-700"
            title="New Chat">
            <Plus className="h-3 w-3" />
          </button>
          <button onClick={() => setIsExpanded(!isExpanded)}>
            {isExpanded ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronRight className="h-4 w-4" />
            )}
          </button>
        </span>
      </div>

      {isExpanded && (
        <div className="ml-6 mt-1">
          {isLoading && conversations.length === 0 && (
            <div className="px-3 py-1 text-xs text-zinc-400 italic">Loading...</div>
          )}

          {!isLoading && conversations.length === 0 && (
            <div className="px-3 py-1 text-xs text-zinc-400 italic">No conversations yet</div>
          )}

          {recentConversations.map((conv) => (
            <div
              key={conv.thread_id}
              onClick={() => handleSelect(conv.thread_id)}
              onMouseEnter={() => setHoveredId(conv.thread_id)}
              onMouseLeave={() => setHoveredId(null)}
              className={`flex items-center gap-2 rounded px-3 py-1 text-sm cursor-pointer text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 hover:text-zinc-900 dark:hover:text-zinc-100 ${
                currentThreadId === conv.thread_id
                  ? "font-bold bg-zinc-100 dark:bg-zinc-800"
                  : "font-normal"
              }`}>
              <span className="flex-1 truncate">{truncateTitle(conv.title)}</span>
              {hoveredId === conv.thread_id && (
                <button
                  onClick={(e) => handleArchive(e, conv.thread_id)}
                  className="p-1 rounded hover:bg-zinc-200 dark:hover:bg-zinc-700"
                  title="Archive">
                  <Archive className="h-3 w-3" />
                </button>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default ChatHistory;
