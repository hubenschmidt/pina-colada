"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Plus, Trash2, Edit2, Check, X } from "lucide-react";
import { getNotes, createNote, updateNote, deleteNote } from "../../api";

const NotesSection = ({
  entityType,
  entityId,
  pendingNotes,
  onPendingNotesChange,
}) => {
  const isCreateMode = entityId === null;
  const [notes, setNotes] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [showAddForm, setShowAddForm] = useState(false);
  const [newNoteContent, setNewNoteContent] = useState("");
  const [editingNoteId, setEditingNoteId] = useState(null);
  const [editContent, setEditContent] = useState("");
  const [error, setError] = useState(null);

  // Fetch notes when entityId changes (edit mode only)
  useEffect(() => {
    if (!isCreateMode && entityId) {
      fetchNotes();
    }
  }, [entityType, entityId, isCreateMode]);

  const fetchNotes = async () => {
    if (!entityId) return;
    setIsLoading(true);
    try {
      const data = await getNotes(entityType, entityId);
      setNotes(data);
    } catch (err) {
      console.error("Failed to fetch notes:", err);
    } finally {
      setIsLoading(false);
    }
  };

  const handleAddNote = async () => {
    if (!newNoteContent.trim()) return;

    if (isCreateMode && onPendingNotesChange) {
      // Create mode: add to pending notes
      onPendingNotesChange([...(pendingNotes || []), newNoteContent.trim()]);
      setNewNoteContent("");
      setShowAddForm(false);
      return;
    }

    if (!entityId) return;

    setError(null);
    try {
      const note = await createNote(
        entityType,
        entityId,
        newNoteContent.trim(),
      );
      setNotes([note, ...notes]);
      setNewNoteContent("");
      setShowAddForm(false);
    } catch (err) {
      setError(err?.message || "Failed to add note");
    }
  };

  const handleUpdateNote = async (noteId) => {
    if (!editContent.trim()) return;

    setError(null);
    try {
      const updated = await updateNote(noteId, editContent.trim());
      setNotes(notes.map((n) => (n.id === noteId ? updated : n)));
      setEditingNoteId(null);
      setEditContent("");
    } catch (err) {
      setError(err?.message || "Failed to update note");
    }
  };

  const handleDeleteNote = async (noteId) => {
    setError(null);
    try {
      await deleteNote(noteId);
      setNotes(notes.filter((n) => n.id !== noteId));
    } catch (err) {
      setError(err?.message || "Failed to delete note");
    }
  };

  const handleRemovePendingNote = (index) => {
    if (onPendingNotesChange && pendingNotes) {
      onPendingNotesChange(pendingNotes.filter((_, i) => i !== index));
    }
  };

  const startEditing = (note) => {
    setEditingNoteId(note.id);
    setEditContent(note.content);
  };

  const cancelEditing = () => {
    setEditingNoteId(null);
    setEditContent("");
  };

  const formatDate = (dateStr) => {
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

  const renderNoteCard = (note) => {
    const isEditing = editingNoteId === note.id;

    return (
      <div
        key={note.id}
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
                onClick={() => handleUpdateNote(note.id)}
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
              {note.content}
            </p>
            <div className="flex items-center justify-between mt-2">
              <div className="flex items-center gap-2 text-xs text-zinc-500">
                {note.created_by_name && (
                  <>
                    {note.individual_id ? (
                      <Link
                        href={`/accounts/individuals/${note.individual_id}`}
                        className="text-zinc-700 dark:text-zinc-300 hover:text-zinc-500 dark:hover:text-zinc-400 hover:underline"
                      >
                        {note.created_by_name}
                      </Link>
                    ) : (
                      <span>{note.created_by_name}</span>
                    )}
                    <span>·</span>
                  </>
                )}
                <span>{formatDate(note.created_at)}</span>
              </div>
              <div className="flex gap-1">
                <button
                  type="button"
                  onClick={() => startEditing(note)}
                  className="p-1 text-zinc-400 hover:text-zinc-600"
                  title="Edit"
                >
                  <Edit2 size={14} />
                </button>
                <button
                  type="button"
                  onClick={() => handleDeleteNote(note.id)}
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

  const renderPendingNoteCard = (content, index) => (
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
          onClick={() => handleRemovePendingNote(index)}
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
        value={newNoteContent}
        onChange={(e) => setNewNoteContent(e.target.value)}
        className={inputClasses}
        rows={3}
        placeholder="Add a note..."
        autoFocus
      />

      <div className="flex gap-2 mt-3">
        <button
          type="button"
          onClick={handleAddNote}
          disabled={!newNoteContent.trim()}
          className="px-3 py-1 text-sm bg-lime-600 text-white rounded hover:bg-lime-700 disabled:opacity-50"
        >
          Add
        </button>
        <button
          type="button"
          onClick={() => {
            setShowAddForm(false);
            setNewNoteContent("");
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
          Notes
        </span>
        {!showAddForm && (
          <button
            type="button"
            onClick={() => setShowAddForm(true)}
            className="flex items-center gap-1 text-sm text-lime-600 dark:text-lime-400 hover:text-lime-700 dark:hover:text-lime-300"
          >
            <Plus size={16} />
            Add Note
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
      {isLoading && <p className="text-sm text-zinc-500">Loading notes...</p>}

      {/* Notes List */}
      <div className="space-y-2">
        {/* Existing notes (edit mode) */}
        {notes.map((note) => renderNoteCard(note))}

        {/* Pending notes (create mode) */}
        {pendingNotes?.map((content, index) =>
          renderPendingNoteCard(content, index),
        )}
      </div>

      {/* Empty state */}
      {!isLoading &&
        notes.length === 0 &&
        (!pendingNotes || pendingNotes.length === 0) &&
        !showAddForm && (
          <p className="text-sm text-zinc-500 dark:text-zinc-400">
            No notes yet.
          </p>
        )}
    </div>
  );
};

export default NotesSection;
