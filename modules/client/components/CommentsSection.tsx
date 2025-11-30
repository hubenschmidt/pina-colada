"use client";

import { useState, useEffect } from "react";
import { Send, Check, X, MessageCircle, Reply, ChevronDown, ChevronUp } from "lucide-react";
import { Comment, getComments, createComment, updateComment, deleteComment } from "../api";

interface CommentsSectionProps {
  entityType: string;
  entityId: number | null;
  pendingComments?: string[];
  onPendingCommentsChange?: (comments: string[]) => void;
}

interface CommentThread {
  comment: Comment;
  replies: Comment[];
}

const CommentsSection = ({
  entityType,
  entityId,
  pendingComments,
  onPendingCommentsChange,
}: CommentsSectionProps) => {
  const isCreateMode = entityId === null;
  const [comments, setComments] = useState<Comment[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [newCommentContent, setNewCommentContent] = useState("");
  const [editingCommentId, setEditingCommentId] = useState<number | null>(null);
  const [editContent, setEditContent] = useState("");
  const [replyingToId, setReplyingToId] = useState<number | null>(null);
  const [replyContent, setReplyContent] = useState("");
  const [collapsedThreads, setCollapsedThreads] = useState<Set<number>>(new Set());
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!isCreateMode && entityId) {
      fetchComments();
    }
  }, [entityType, entityId, isCreateMode]);

  const fetchComments = async () => {
    if (!entityId) return;
    setIsLoading(true);
    try {
      const data = await getComments(entityType, entityId);
      setComments(data);
    } catch (err) {
      console.error("Failed to fetch comments:", err);
    } finally {
      setIsLoading(false);
    }
  };

  // Organize comments into threads (top-level + replies)
  const organizeThreads = (): CommentThread[] => {
    const topLevel = comments.filter((c) => !c.parent_comment_id);
    const threads: CommentThread[] = topLevel.map((comment) => ({
      comment,
      replies: comments
        .filter((c) => c.parent_comment_id === comment.id)
        .sort((a, b) => new Date(a.created_at || 0).getTime() - new Date(b.created_at || 0).getTime()),
    }));
    return threads.sort((a, b) =>
      new Date(b.comment.created_at || 0).getTime() - new Date(a.comment.created_at || 0).getTime()
    );
  };

  const handleAddComment = async () => {
    if (!newCommentContent.trim()) return;

    if (isCreateMode && onPendingCommentsChange) {
      onPendingCommentsChange([...(pendingComments || []), newCommentContent.trim()]);
      setNewCommentContent("");
      return;
    }

    if (!entityId) return;

    setError(null);
    try {
      const comment = await createComment(entityType, entityId, newCommentContent.trim());
      setComments([...comments, comment]);
      setNewCommentContent("");
    } catch (err: any) {
      setError(err?.message || "Failed to add comment");
    }
  };

  const handleAddReply = async (parentId: number) => {
    if (!replyContent.trim() || !entityId) return;

    setError(null);
    try {
      const comment = await createComment(entityType, entityId, replyContent.trim(), parentId);
      setComments([...comments, comment]);
      setReplyContent("");
      setReplyingToId(null);
    } catch (err: any) {
      setError(err?.message || "Failed to add reply");
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>, onSubmit: () => void) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      onSubmit();
    }
  };

  const handleUpdateComment = async (commentId: number) => {
    if (!editContent.trim()) return;

    setError(null);
    try {
      const updated = await updateComment(commentId, editContent.trim());
      setComments(comments.map((c) => (c.id === commentId ? updated : c)));
      setEditingCommentId(null);
      setEditContent("");
    } catch (err: any) {
      setError(err?.message || "Failed to update comment");
    }
  };

  const handleDeleteComment = async (commentId: number) => {
    setError(null);
    try {
      await deleteComment(commentId);
      // Remove the comment and all its replies
      setComments(comments.filter((c) => c.id !== commentId && c.parent_comment_id !== commentId));
    } catch (err: any) {
      setError(err?.message || "Failed to delete comment");
    }
  };

  const handleRemovePendingComment = (index: number) => {
    if (onPendingCommentsChange && pendingComments) {
      onPendingCommentsChange(pendingComments.filter((_, i) => i !== index));
    }
  };

  const startEditing = (comment: Comment) => {
    setEditingCommentId(comment.id);
    setEditContent(comment.content);
    setReplyingToId(null);
  };

  const cancelEditing = () => {
    setEditingCommentId(null);
    setEditContent("");
  };

  const startReplying = (commentId: number) => {
    setReplyingToId(commentId);
    setReplyContent("");
    setEditingCommentId(null);
  };

  const cancelReplying = () => {
    setReplyingToId(null);
    setReplyContent("");
  };

  const toggleThread = (commentId: number) => {
    setCollapsedThreads((prev) => {
      const next = new Set(prev);
      if (next.has(commentId)) {
        next.delete(commentId);
      } else {
        next.add(commentId);
      }
      return next;
    });
  };

  const formatRelativeTime = (dateStr: string | null) => {
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

  const getInitials = (name: string | null, email: string | null) => {
    if (name) {
      const parts = name.trim().split(" ");
      if (parts.length >= 2) {
        return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
      }
      return name.slice(0, 2).toUpperCase();
    }
    if (email) {
      return email.slice(0, 2).toUpperCase();
    }
    return "??";
  };

  const getAvatarColor = (name: string | null, email: string | null) => {
    const str = name || email || "unknown";
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      hash = str.charCodeAt(i) + ((hash << 5) - hash);
    }
    const colors = [
      "bg-blue-500",
      "bg-emerald-500",
      "bg-violet-500",
      "bg-amber-500",
      "bg-rose-500",
      "bg-cyan-500",
      "bg-fuchsia-500",
      "bg-lime-600",
    ];
    return colors[Math.abs(hash) % colors.length];
  };

  const renderCommentContent = (comment: Comment, isReply: boolean = false) => {
    const isEditing = editingCommentId === comment.id;
    const displayName = comment.created_by_name || comment.created_by_email?.split("@")[0] || "Unknown";

    if (isEditing) {
      return (
        <div className="space-y-2 flex-1">
          <textarea
            value={editContent}
            onChange={(e) => setEditContent(e.target.value)}
            onKeyDown={(e) => handleKeyDown(e, () => handleUpdateComment(comment.id))}
            className="w-full px-3 py-2 text-sm border border-zinc-300 dark:border-zinc-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 resize-none"
            rows={2}
            autoFocus
          />
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => handleUpdateComment(comment.id)}
              className="px-3 py-1 text-xs bg-lime-600 text-white rounded hover:bg-lime-700 transition-colors"
            >
              Save
            </button>
            <button
              type="button"
              onClick={cancelEditing}
              className="px-3 py-1 text-xs text-zinc-500 hover:text-zinc-700 dark:hover:text-zinc-300 transition-colors"
            >
              Cancel
            </button>
          </div>
        </div>
      );
    }

    return (
      <div className="flex-1 min-w-0">
        {/* Header: name and time */}
        <div className="flex items-center gap-2 mb-1">
          <span className="text-sm font-semibold text-zinc-900 dark:text-zinc-100">
            {displayName}
          </span>
          <span className="text-xs text-zinc-400 dark:text-zinc-500">
            {formatRelativeTime(comment.created_at)}
          </span>
        </div>

        {/* Content */}
        <p className="text-sm text-zinc-700 dark:text-zinc-300 whitespace-pre-wrap break-words">
          {comment.content}
        </p>

        {/* Edited indicator */}
        {comment.updated_at && comment.created_at && comment.updated_at !== comment.created_at && (
          <span className="text-xs text-zinc-400 dark:text-zinc-500 italic">
            (edited)
          </span>
        )}

        {/* Actions */}
        <div className="flex items-center gap-3 text-xs">
          {!isReply && (
            <button
              type="button"
              onClick={() => startReplying(comment.id)}
              className="flex items-center gap-1 text-zinc-400 hover:text-lime-600 dark:hover:text-lime-400 transition-colors"
            >
              <Reply size={12} />
              Reply
            </button>
          )}
          <button
            type="button"
            onClick={() => startEditing(comment)}
            className="text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-300 transition-colors"
          >
            Edit
          </button>
          <button
            type="button"
            onClick={() => handleDeleteComment(comment.id)}
            className="text-zinc-400 hover:text-red-500 transition-colors"
          >
            Delete
          </button>
        </div>
      </div>
    );
  };

  const renderThread = (thread: CommentThread) => {
    const { comment, replies } = thread;
    const isCollapsed = collapsedThreads.has(comment.id);
    const hasReplies = replies.length > 0;
    const isReplying = replyingToId === comment.id;

    return (
      <div key={comment.id} className="group">
        {/* Main comment */}
        <div className="flex gap-3">
          {/* Avatar */}
          <div
            className={`flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center text-white text-xs font-medium ${getAvatarColor(comment.created_by_name, comment.created_by_email)}`}
          >
            {getInitials(comment.created_by_name, comment.created_by_email)}
          </div>

          {renderCommentContent(comment)}
        </div>

        {/* Reply input */}
        {isReplying && (
          <div className="ml-11 mt-3 space-y-2">
            <textarea
              value={replyContent}
              onChange={(e) => setReplyContent(e.target.value)}
              onKeyDown={(e) => handleKeyDown(e, () => handleAddReply(comment.id))}
              className="w-full px-3 py-2 text-sm border border-zinc-200 dark:border-zinc-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-lime-500 bg-zinc-50 dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 placeholder-zinc-400 resize-none"
              rows={2}
              placeholder="Write a reply..."
              autoFocus
            />
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => handleAddReply(comment.id)}
                disabled={!replyContent.trim()}
                className="px-3 py-1.5 text-xs bg-lime-600 text-white rounded hover:bg-lime-700 disabled:opacity-40 transition-colors"
              >
                Reply
              </button>
              <button
                type="button"
                onClick={cancelReplying}
                className="px-3 py-1.5 text-xs text-zinc-500 hover:text-zinc-700 dark:hover:text-zinc-300 transition-colors"
              >
                Cancel
              </button>
            </div>
          </div>
        )}

        {/* Replies */}
        {hasReplies && (
          <div className="ml-11 mt-3">
            {/* Collapse/expand toggle */}
            <button
              type="button"
              onClick={() => toggleThread(comment.id)}
              className="flex items-center gap-1 text-xs text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-300 mb-2 transition-colors"
            >
              {isCollapsed ? (
                <>
                  <ChevronDown size={14} />
                  Show {replies.length} {replies.length === 1 ? "reply" : "replies"}
                </>
              ) : (
                <>
                  <ChevronUp size={14} />
                  Hide replies
                </>
              )}
            </button>

            {/* Reply list */}
            {!isCollapsed && (
              <div className="space-y-3 border-l-2 border-zinc-200 dark:border-zinc-700 pl-4">
                {replies.map((reply) => (
                  <div key={reply.id} className="flex gap-3">
                    {/* Reply avatar (smaller) */}
                    <div
                      className={`flex-shrink-0 w-6 h-6 rounded-full flex items-center justify-center text-white text-[10px] font-medium ${getAvatarColor(reply.created_by_name, reply.created_by_email)}`}
                    >
                      {getInitials(reply.created_by_name, reply.created_by_email)}
                    </div>

                    {renderCommentContent(reply, true)}
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    );
  };

  const renderPendingCommentCard = (content: string, index: number) => (
    <div key={`pending-${index}`} className="flex gap-3">
      <div className="flex-shrink-0 w-8 h-8 rounded-full bg-zinc-300 dark:bg-zinc-600 flex items-center justify-center">
        <span className="text-xs text-zinc-500 dark:text-zinc-400">You</span>
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-1">
          <span className="text-sm font-semibold text-zinc-900 dark:text-zinc-100">You</span>
          <span className="text-xs text-lime-600 dark:text-lime-400 italic">pending</span>
        </div>
        <p className="text-sm text-zinc-700 dark:text-zinc-300 whitespace-pre-wrap break-words mb-2">
          {content}
        </p>
        <button
          type="button"
          onClick={() => handleRemovePendingComment(index)}
          className="text-xs text-zinc-400 hover:text-red-500 transition-colors"
        >
          Remove
        </button>
      </div>
    </div>
  );

  const threads = organizeThreads();
  const totalComments = comments.length + (pendingComments?.length || 0);

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center gap-2">
        <MessageCircle size={18} className="text-zinc-400" />
        <span className="text-sm font-medium text-zinc-700 dark:text-zinc-300">
          Comments
        </span>
        {totalComments > 0 && (
          <span className="text-xs text-zinc-400 bg-zinc-100 dark:bg-zinc-800 px-2 py-0.5 rounded-full">
            {totalComments}
          </span>
        )}
      </div>

      {/* Error */}
      {error && (
        <div className="p-2 text-sm text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
          {error}
        </div>
      )}

      {/* Composer - at top for new comments */}
      <div className="flex gap-3 pb-4 border-b border-zinc-100 dark:border-zinc-800">
        <div className="flex-shrink-0 w-8 h-8 rounded-full bg-lime-600 flex items-center justify-center text-white text-xs font-medium">
          You
        </div>
        <div className="flex-1 space-y-2">
          <textarea
            value={newCommentContent}
            onChange={(e) => setNewCommentContent(e.target.value)}
            onKeyDown={(e) => handleKeyDown(e, handleAddComment)}
            className="w-full px-3 py-2 text-sm border border-zinc-200 dark:border-zinc-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-lime-500 focus:border-transparent bg-zinc-50 dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 placeholder-zinc-400 resize-none"
            rows={2}
            placeholder="Write a comment..."
          />
          <button
            type="button"
            onClick={handleAddComment}
            disabled={!newCommentContent.trim()}
            className="px-4 py-1.5 bg-lime-600 text-white text-sm rounded-lg hover:bg-lime-700 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
          >
            Post
          </button>
        </div>
      </div>

      {/* Comments List */}
      {isLoading ? (
        <div className="flex items-center gap-2 text-sm text-zinc-500 py-4">
          <div className="w-4 h-4 border-2 border-zinc-300 border-t-zinc-600 rounded-full animate-spin" />
          Loading comments...
        </div>
      ) : (
        <div className="space-y-6">
          {pendingComments?.map((content, index) => renderPendingCommentCard(content, index))}
          {threads.map((thread) => renderThread(thread))}
        </div>
      )}

      {/* Empty state */}
      {!isLoading && totalComments === 0 && (
        <p className="text-sm text-zinc-400 dark:text-zinc-500 italic py-2">
          No comments yet. Be the first to comment.
        </p>
      )}
    </div>
  );
};

export default CommentsSection;
