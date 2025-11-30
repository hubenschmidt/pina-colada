"use client";

import { useState, useEffect } from "react";
import { Plus, Trash2, Edit2, Check, X } from "lucide-react";
import { Comment, getComments, createComment, updateComment, deleteComment } from "../api";

interface CommentsSectionProps {
  entityType: string;
  entityId: number | null;
  // For create mode - collect comments to save after entity creation
  pendingComments?: string[];
  onPendingCommentsChange?: (comments: string[]) => void;
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
  const [showAddForm, setShowAddForm] = useState(false);
  const [newCommentContent, setNewCommentContent] = useState("");
  const [editingCommentId, setEditingCommentId] = useState<number | null>(null);
  const [editContent, setEditContent] = useState("");
  const [error, setError] = useState<string | null>(null);

  // Fetch comments when entityId changes (edit mode only)
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

  const handleAddComment = async () => {
    if (!newCommentContent.trim()) return;

    if (isCreateMode && onPendingCommentsChange) {
      // Create mode: add to pending comments
      onPendingCommentsChange([...(pendingComments || []), newCommentContent.trim()]);
      setNewCommentContent("");
      setShowAddForm(false);
      return;
    }

    if (!entityId) return;

    setError(null);
    try {
      const comment = await createComment(entityType, entityId, newCommentContent.trim());
      setComments([...comments, comment]);
      setNewCommentContent("");
      setShowAddForm(false);
    } catch (err: any) {
      setError(err?.message || "Failed to add comment");
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
      setComments(comments.filter((c) => c.id !== commentId));
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
  };

  const cancelEditing = () => {
    setEditingCommentId(null);
    setEditContent("");
  };

  const formatDate = (dateStr: string | null) => {
    if (!dateStr) return "";
    const date = new Date(dateStr);
    return date.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
      hour: "numeric",
      minute: "2-digit",
    });
  };

  const inputClasses =
    "w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100";

  const renderCommentCard = (comment: Comment) => {
    const isEditing = editingCommentId === comment.id;

    return (
      <div
        key={comment.id}
        className="p-3 border border-zinc-200 dark:border-zinc-700 rounded"
      >
        {isEditing ? (
          <div className="space-y-2">
            <textarea
              value={editContent}
              onChange={(e) => setEditContent(e.target.value)}
              className={inputClasses}
              rows={3}
              autoFocus
            />
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => handleUpdateComment(comment.id)}
                className="p-1 text-lime-600 hover:text-lime-700"
                title="Save"
              >
                <Check size={16} />
              </button>
              <button
                type="button"
                onClick={cancelEditing}
                className="p-1 text-zinc-400 hover:text-zinc-600"
                title="Cancel"
              >
                <X size={16} />
              </button>
            </div>
          </div>
        ) : (
          <>
            <p className="text-sm text-zinc-900 dark:text-zinc-100 whitespace-pre-wrap">
              {comment.content}
            </p>
            <div className="flex items-center justify-between mt-2">
              <span className="text-xs text-zinc-500">
                {comment.created_by_name || comment.created_by_email || "Unknown"}
                {" · "}
                {formatDate(comment.created_at)}
              </span>
              <div className="flex gap-1">
                <button
                  type="button"
                  onClick={() => startEditing(comment)}
                  className="p-1 text-zinc-400 hover:text-zinc-600"
                  title="Edit"
                >
                  <Edit2 size={14} />
                </button>
                <button
                  type="button"
                  onClick={() => handleDeleteComment(comment.id)}
                  className="p-1 text-zinc-400 hover:text-red-500"
                  title="Delete"
                >
                  <Trash2 size={14} />
                </button>
              </div>
            </div>
          </>
        )}
      </div>
    );
  };

  const renderPendingCommentCard = (content: string, index: number) => (
    <div
      key={`pending-${index}`}
      className="p-3 border border-zinc-200 dark:border-zinc-700 rounded"
    >
      <p className="text-sm text-zinc-900 dark:text-zinc-100 whitespace-pre-wrap">
        {content}
      </p>
      <div className="flex items-center justify-between mt-2">
        <span className="text-xs text-zinc-500 italic">Pending</span>
        <button
          type="button"
          onClick={() => handleRemovePendingComment(index)}
          className="p-1 text-zinc-400 hover:text-red-500"
          title="Remove"
        >
          <Trash2 size={14} />
        </button>
      </div>
    </div>
  );

  const renderAddForm = () => (
    <div className="mb-4 p-4 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-zinc-50 dark:bg-zinc-800/50">
      <textarea
        value={newCommentContent}
        onChange={(e) => setNewCommentContent(e.target.value)}
        className={inputClasses}
        rows={3}
        placeholder="Add a comment..."
        autoFocus
      />
      <div className="flex gap-2 mt-3">
        <button
          type="button"
          onClick={handleAddComment}
          disabled={!newCommentContent.trim()}
          className="px-3 py-1 text-sm bg-lime-600 text-white rounded hover:bg-lime-700 disabled:opacity-50"
        >
          Add
        </button>
        <button
          type="button"
          onClick={() => {
            setShowAddForm(false);
            setNewCommentContent("");
          }}
          className="px-3 py-1 text-sm bg-zinc-200 dark:bg-zinc-700 text-zinc-700 dark:text-zinc-300 rounded hover:bg-zinc-300 dark:hover:bg-zinc-600"
        >
          Cancel
        </button>
      </div>
    </div>
  );

  return (
    <div className="space-y-2">
      {/* Header */}
      <div className="flex items-center justify-between mb-3">
        <span className="text-sm font-medium text-zinc-700 dark:text-zinc-300">
          Comments
        </span>
        {!showAddForm && (
          <button
            type="button"
            onClick={() => setShowAddForm(true)}
            className="flex items-center gap-1 text-sm text-lime-600 dark:text-lime-400 hover:text-lime-700 dark:hover:text-lime-300"
          >
            <Plus size={16} />
            Add Comment
          </button>
        )}
      </div>

      {/* Error */}
      {error && (
        <div className="p-2 text-sm text-red-600 dark:text-red-400 bg-red-100 dark:bg-red-900/30 rounded">
          {error}
        </div>
      )}

      {/* Add Form */}
      {showAddForm && renderAddForm()}

      {/* Loading */}
      {isLoading && (
        <p className="text-sm text-zinc-500">Loading comments...</p>
      )}

      {/* Comments List */}
      <div className="space-y-2">
        {/* Pending comments (create mode) - show first */}
        {pendingComments?.map((content, index) => renderPendingCommentCard(content, index))}

        {/* Existing comments (edit mode) */}
        {comments.map((comment) => renderCommentCard(comment))}
      </div>

      {/* Empty state */}
      {!isLoading && comments.length === 0 && (!pendingComments || pendingComments.length === 0) && !showAddForm && (
        <p className="text-sm text-zinc-500 dark:text-zinc-400">
          No comments yet.
        </p>
      )}
    </div>
  );
};

export default CommentsSection;
