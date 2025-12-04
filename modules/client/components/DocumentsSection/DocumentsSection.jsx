"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Trash2, Download, Plus, Upload } from "lucide-react";
import { Collapse } from "@mantine/core";
import {
  getDocuments,
  getDocument,
  linkDocumentToEntity,
  unlinkDocumentFromEntity,
  downloadDocument,
} from "../../api";
import { DocumentUpload } from "../Documents/DocumentUpload";
import { useRouter } from "next/navigation";

const DocumentsSection = ({
  entityType,
  entityId,
  pendingDocumentIds,
  onPendingDocumentIdsChange,
}) => {
  const router = useRouter();
  const isCreateMode = entityId === null;
  const [documents, setDocuments] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [showLinkForm, setShowLinkForm] = useState(false);
  const [showUploadForm, setShowUploadForm] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState([]);
  const [isSearching, setIsSearching] = useState(false);
  const [error, setError] = useState(null);
  const [unlinkConfirmId, setUnlinkConfirmId] = useState(null);

  // Fetch linked documents when entityId changes (edit mode only)
  useEffect(() => {
    if (!isCreateMode && entityId) {
      fetchLinkedDocuments();
    }
  }, [entityType, entityId, isCreateMode]);

  const fetchLinkedDocuments = async () => {
    if (!entityId) return;
    setIsLoading(true);
    setError(null);
    try {
      // Fetch documents filtered by this entity
      const pageData = await getDocuments(
        1,
        100, // Get up to 100 documents
        "updated_at",
        "DESC",
        undefined,
        undefined,
        entityType,
        entityId,
      );
      setDocuments(pageData.items || []);
    } catch (err) {
      console.error("Failed to fetch documents:", err);
      setError("Failed to load documents");
    } finally {
      setIsLoading(false);
    }
  };

  const handleSearchDocuments = async (query) => {
    if (!query.trim()) {
      setSearchResults([]);
      return;
    }

    setIsSearching(true);
    try {
      const pageData = await getDocuments(
        1,
        20, // Limit search results
        "updated_at",
        "DESC",
        query.trim(),
      );
      // Filter out already linked documents
      const linkedIds = new Set(
        isCreateMode ? pendingDocumentIds || [] : documents.map((d) => d.id),
      );
      setSearchResults(
        (pageData.items || []).filter((d) => !linkedIds.has(d.id)),
      );
    } catch (err) {
      console.error("Failed to search documents:", err);
    } finally {
      setIsSearching(false);
    }
  };

  const handleLinkDocument = async (documentId) => {
    if (isCreateMode && onPendingDocumentIdsChange) {
      // Create mode: add to pending document IDs
      onPendingDocumentIdsChange([...(pendingDocumentIds || []), documentId]);
      setSearchQuery("");
      setSearchResults([]);
      setShowLinkForm(false);
      return;
    }

    if (!entityId) return;

    setError(null);
    try {
      await linkDocumentToEntity(documentId, entityType, entityId);
      await fetchLinkedDocuments();
      setSearchQuery("");
      setSearchResults([]);
      setShowLinkForm(false);
    } catch (err) {
      setError(err?.message || "Failed to link document");
    }
  };

  const handleUploadComplete = async (document) => {
    if (isCreateMode && onPendingDocumentIdsChange) {
      // Create mode: add to pending document IDs
      // Note: DocumentUpload already links during upload if entityType/entityId are provided,
      // so we just need to track it for pending state
      onPendingDocumentIdsChange([...(pendingDocumentIds || []), document.id]);
      setShowUploadForm(false);
      return;
    }

    if (!entityId) return;

    // Edit mode: DocumentUpload already links during upload if entityType/entityId are provided
    // Just refresh the document list to show the newly uploaded document
    setError(null);
    try {
      await fetchLinkedDocuments();
      setShowUploadForm(false);
    } catch (err) {
      setError(err?.message || "Failed to refresh documents");
    }
  };

  const handleUnlinkDocument = (documentId) => {
    setUnlinkConfirmId(documentId);
  };

  const confirmUnlink = async () => {
    if (!entityId || !unlinkConfirmId) return;

    setError(null);
    try {
      await unlinkDocumentFromEntity(unlinkConfirmId, entityType, entityId);
      setUnlinkConfirmId(null);
      await fetchLinkedDocuments();
    } catch (err) {
      setError(err?.message || "Failed to unlink document");
    }
  };

  const cancelUnlink = () => {
    setUnlinkConfirmId(null);
  };

  const handleDownload = async (doc) => {
    try {
      await downloadDocument(doc.id, doc.filename);
    } catch (err) {
      console.error("Download failed:", err);
      setError("Failed to download document");
    }
  };

  const handleRemovePendingDocument = (documentId) => {
    if (onPendingDocumentIdsChange && pendingDocumentIds) {
      onPendingDocumentIdsChange(
        pendingDocumentIds.filter((id) => id !== documentId),
      );
    }
  };

  const formatFileSize = (bytes) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return "";
    const date = new Date(dateStr);
    return date.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    });
  };

  // Get pending documents (for create mode)
  const getPendingDocuments = async () => {
    if (!pendingDocumentIds || pendingDocumentIds.length === 0) return [];
    try {
      const results = await Promise.all(
        pendingDocumentIds.map(async (id) => {
          try {
            const doc = await getDocument(id);
            return doc;
          } catch {
            return null;
          }
        }),
      );
      return results.filter((d) => d !== null);
    } catch {
      return [];
    }
  };

  const [pendingDocuments, setPendingDocuments] = useState([]);

  useEffect(() => {
    if (
      !isCreateMode ||
      !pendingDocumentIds ||
      pendingDocumentIds.length === 0
    ) {
      setPendingDocuments([]);
      return;
    }
    getPendingDocuments().then(setPendingDocuments);
  }, [isCreateMode, pendingDocumentIds]);

  const inputClasses =
    "w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100";

  const renderDocumentCard = (doc, isPending = false) => (
    <div key={doc.id}>
      <div className="p-3 border border-zinc-200 dark:border-zinc-700 rounded flex items-center justify-between">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <Link
              href={`/assets/documents/${doc.id}?v=${doc.version_number}`}
              className="text-sm font-medium text-zinc-900 dark:text-zinc-100 hover:text-zinc-500 dark:hover:text-zinc-400 hover:underline"
              onClick={(e) => {
                e.stopPropagation();
                router.push(
                  `/assets/documents/${doc.id}?v=${doc.version_number}`,
                );
              }}
            >
              {doc.filename}
            </Link>
            <span className="text-xs px-1.5 py-0.5 bg-zinc-200 dark:bg-zinc-700 text-zinc-600 dark:text-zinc-400 rounded">
              v{doc.version_number}
            </span>
          </div>
          <div className="flex items-center gap-2 mt-1">
            <span className="text-xs text-zinc-500">
              {formatFileSize(doc.file_size)}
            </span>
            <span className="text-xs text-zinc-400">•</span>
            <span className="text-xs text-zinc-500">
              {formatDate(doc.created_at)}
            </span>
            {isPending && (
              <>
                <span className="text-xs text-zinc-400">•</span>
                <span className="text-xs text-zinc-500 italic">Pending</span>
              </>
            )}
          </div>
        </div>
        <div className="flex items-center gap-1 ml-2">
          <button
            type="button"
            onClick={() => handleDownload(doc)}
            className="p-1.5 text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-300"
            title="Download"
          >
            <Download size={16} />
          </button>
          {!isPending && (
            <button
              type="button"
              onClick={() =>
                isCreateMode
                  ? handleRemovePendingDocument(doc.id)
                  : handleUnlinkDocument(doc.id)
              }
              className="p-1.5 text-zinc-400 hover:text-red-500"
              title={isCreateMode ? "Remove" : "Unlink"}
            >
              <Trash2 size={16} />
            </button>
          )}
        </div>
      </div>
      {unlinkConfirmId === doc.id && (
        <div className="mt-1 p-2 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded text-sm">
          <div className="flex items-center justify-between">
            <span className="text-red-700 dark:text-red-400">
              Unlink this document?
            </span>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={cancelUnlink}
                className="px-2 py-1 text-xs bg-zinc-200 dark:bg-zinc-700 text-zinc-700 dark:text-zinc-300 rounded hover:bg-zinc-300 dark:hover:bg-zinc-600"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={confirmUnlink}
                className="px-2 py-1 text-xs bg-red-500 text-white rounded hover:bg-red-600"
              >
                Unlink
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );

  const renderLinkForm = () => (
    <div className="mb-4 p-4 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-zinc-50 dark:bg-zinc-800/50">
      <div className="space-y-3">
        <input
          type="text"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          onKeyDown={(e) =>
            e.key === "Enter" && handleSearchDocuments(searchQuery)
          }
          className={inputClasses}
          placeholder="Search documents... (Enter to search)"
          autoFocus
        />

        {isSearching && <p className="text-sm text-zinc-500">Searching...</p>}
        {searchResults.length > 0 && (
          <div className="space-y-2 max-h-48 overflow-y-auto">
            {searchResults.map((doc) => (
              <div
                key={doc.id}
                className="p-2 border border-zinc-200 dark:border-zinc-700 rounded hover:bg-zinc-100 dark:hover:bg-zinc-700 cursor-pointer flex items-center justify-between"
                onClick={() => handleLinkDocument(doc.id)}
              >
                <span className="text-sm text-lime-600 dark:text-lime-400 hover:text-lime-700 dark:hover:text-lime-300 hover:underline truncate flex-1 min-w-0">
                  {doc.filename}
                </span>
                <Plus size={16} className="text-lime-600 flex-shrink-0 ml-2" />
              </div>
            ))}
          </div>
        )}
        {searchQuery && !isSearching && searchResults.length === 0 && (
          <p className="text-sm text-zinc-500">
            No documents found. Upload documents from the Documents page.
          </p>
        )}
        <div className="flex gap-2">
          <button
            type="button"
            onClick={() => {
              setShowLinkForm(false);
              setSearchQuery("");
              setSearchResults([]);
            }}
            className="px-3 py-1 text-sm bg-zinc-200 dark:bg-zinc-700 text-zinc-700 dark:text-zinc-300 rounded hover:bg-zinc-300 dark:hover:bg-zinc-600"
          >
            Cancel
          </button>
        </div>
      </div>
    </div>
  );

  const displayDocuments = isCreateMode ? pendingDocuments : documents;

  return (
    <div className="space-y-2">
      {/* Header */}
      <div className="flex items-center justify-between mb-3">
        <span className="text-sm font-medium text-zinc-700 dark:text-zinc-300">
          Documents
        </span>
        {!showLinkForm && !showUploadForm && (
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => setShowUploadForm(true)}
              className="flex items-center gap-1 text-sm text-lime-600 dark:text-lime-400 hover:text-lime-700 dark:hover:text-lime-300"
            >
              <Upload size={16} />
              Upload
            </button>
            <span className="text-zinc-400 dark:text-zinc-600">|</span>
            <button
              type="button"
              onClick={() => setShowLinkForm(true)}
              className="flex items-center gap-1 text-sm text-lime-600 dark:text-lime-400 hover:text-lime-700 dark:hover:text-lime-300"
            >
              <Plus size={16} />
              Link Document
            </button>
          </div>
        )}
      </div>

      {/* Error */}
      {error && (
        <div className="p-2 text-sm text-red-600 dark:text-red-400 bg-red-100 dark:bg-red-900/30 rounded">
          {error}
        </div>
      )}

      {/* Upload Form */}
      {showUploadForm && (
        <div className="mb-4">
          <Collapse in={showUploadForm}>
            <DocumentUpload
              onUploadComplete={handleUploadComplete}
              entityType={isCreateMode ? undefined : entityType}
              entityId={isCreateMode ? undefined : entityId || undefined}
            />
          </Collapse>
        </div>
      )}

      {/* Link Form */}
      {showLinkForm && renderLinkForm()}

      {/* Loading */}
      {isLoading && (
        <p className="text-sm text-zinc-500">Loading documents...</p>
      )}

      {/* Documents List */}
      <div className="space-y-2">
        {displayDocuments.map((doc) => renderDocumentCard(doc, isCreateMode))}
      </div>

      {/* Empty state */}
      {!isLoading &&
        displayDocuments.length === 0 &&
        !showLinkForm &&
        !showUploadForm && (
          <p className="text-sm text-zinc-500 dark:text-zinc-400">
            No documents linked yet.
          </p>
        )}
    </div>
  );
};

export default DocumentsSection;
