"use client";

import { useState, useEffect, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import {
  Stack,
  Title,
  Text,
  Paper,
  Group,
  Badge,
  Button,
  Loader,
  Center,
  Anchor,
  Divider,
  ActionIcon,
  Checkbox,
  SimpleGrid,
  TextInput,
} from "@mantine/core";
import { DataTable } from "../../../../components/DataTable/DataTable";
import {
  ArrowLeft,
  Download,
  Trash2,
  FileText,
  Link2,
  Calendar,
  HardDrive,
  History,
  Search,
} from "lucide-react";
import {
  getDocument,
  deleteDocument,
  downloadDocument,
  getDocumentPreviewUrl,
  getDocumentVersions,
  setCurrentDocumentVersion,
} from "../../../../api";
import { DeleteConfirmBanner } from "../../../../components/DeleteConfirmBanner";

const DocumentDetailPage = () => {
  const params = useParams();
  const router = useRouter();
  const documentId = Number(params.id);

  const [document, setDocument] = useState(null);
  const [versions, setVersions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [previewUrl, setPreviewUrl] = useState(null);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [versionsLoading, setVersionsLoading] = useState(false);
  const [versionPage, setVersionPage] = useState(1);
  const [versionPageSize, setVersionPageSize] = useState(10);

  // Linked entities state
  const [entitySearch, setEntitySearch] = useState("");
  const [entityPage, setEntityPage] = useState(1);
  const [entityPageSize, setEntityPageSize] = useState(10);
  const [entitySortBy, setEntitySortBy] = useState("entity_type");
  const [entitySortDir, setEntitySortDir] = useState("ASC");

  // Sorting state for version history
  const [versionSortBy, setVersionSortBy] = useState("version_number");
  const [versionSortDir, setVersionSortDir] = useState("DESC");

  // Delete confirmation state
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const loadDocument = useCallback(async () => {
    if (!documentId) return;
    setLoading(true);
    setError(null);
    try {
      const doc = await getDocument(documentId);
      setDocument(doc);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load document");
    } finally {
      setLoading(false);
    }
  }, [documentId]);

  const loadVersions = useCallback(async () => {
    if (!documentId) return;
    setVersionsLoading(true);
    try {
      const data = await getDocumentVersions(documentId);
      setVersions(data.versions);
    } catch {
      // Versions are optional, ignore errors
    } finally {
      setVersionsLoading(false);
    }
  }, [documentId]);

  useEffect(() => {
    loadDocument();
  }, [loadDocument]);

  useEffect(() => {
    if (document) {
      loadVersions();
    }
  }, [document?.id, loadVersions]);

  useEffect(() => {
    if (!document) return;

    const canPreview = [
      "application/pdf",
      "image/png",
      "image/jpeg",
      "image/gif",
      "image/webp",
    ].includes(document.content_type);
    if (!canPreview) return;

    let url = null;
    setPreviewLoading(true);
    getDocumentPreviewUrl(document.id)
      .then((blobUrl) => {
        url = blobUrl;
        setPreviewUrl(blobUrl);
      })
      .catch(() => setPreviewUrl(null))
      .finally(() => setPreviewLoading(false));

    return () => {
      if (url) window.URL.revokeObjectURL(url);
    };
  }, [document?.id, document?.content_type]);

  const handleDownload = async () => {
    if (!document) return;
    try {
      await downloadDocument(document.id, document.filename);
    } catch (err) {
      console.error("Download failed:", err);
    }
  };

  const handleDeleteClick = () => {
    setShowDeleteConfirm(true);
  };

  const handleDeleteConfirm = async () => {
    if (!document) return;

    setDeleting(true);
    try {
      await deleteDocument(document.id);
      router.push("/assets/documents");
    } catch (err) {
      console.error("Delete failed:", err);
      setDeleting(false);
      setShowDeleteConfirm(false);
    }
  };

  const handleSetCurrentVersion = async (versionId) => {
    try {
      await setCurrentDocumentVersion(versionId);
      await loadDocument();
      await loadVersions();
    } catch (err) {
      console.error("Failed to set current version:", err);
    }
  };

  const handleDownloadVersion = async (version) => {
    try {
      await downloadDocument(version.id, version.filename);
    } catch (err) {
      console.error("Download failed:", err);
    }
  };

  const formatFileSize = (bytes) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return "-";
    return new Date(dateStr).toLocaleString();
  };

  const getEntityUrl = (entityType, entityId) => {
    const typeMap = {
      Organization: `/accounts/organizations/${entityId}`,
      Individual: `/accounts/individuals/${entityId}`,
      Project: `/projects/${entityId}`,
      Contact: `/accounts/contacts/${entityId}`,
      Lead: `/leads/jobs/${entityId}`,
    };
    return typeMap[entityType] || "#";
  };

  const getEntityColor = (entityType) => {
    const colorMap = {
      Organization: "blue",
      Individual: "green",
      Project: "violet",
      Contact: "cyan",
      Lead: "orange",
    };
    return colorMap[entityType] || "gray";
  };

  // Combine entities from all versions with version info

  const allEntities = versions.flatMap((v) =>
    (v.entities || []).map((e) => ({
      ...e,
      version_number: v.version_number,
      version_id: v.id,
    }))
  );

  // Filter and sort entities client-side
  const filteredEntities = allEntities.filter((e) => {
    if (!entitySearch.trim()) return true;
    const search = entitySearch.toLowerCase();
    return (
      e.entity_name.toLowerCase().includes(search) || e.entity_type.toLowerCase().includes(search)
    );
  });

  const sortedEntities = [...filteredEntities].sort((a, b) => {
    const key = entitySortBy;
    if (key === "version_number") {
      const cmp = (a.version_number || 0) - (b.version_number || 0);
      return entitySortDir === "ASC" ? cmp : -cmp;
    }
    const aVal = String(a[key] || "").toLowerCase();
    const bVal = String(b[key] || "").toLowerCase();
    const cmp = aVal.localeCompare(bVal);
    return entitySortDir === "ASC" ? cmp : -cmp;
  });

  // Sort versions client-side
  const sortedVersions = [...versions].sort((a, b) => {
    const key = versionSortBy;
    let cmp = 0;
    if (key === "version_number") {
      cmp = (a.version_number || 0) - (b.version_number || 0);
    } else if (key === "created_at") {
      cmp = new Date(a.created_at || 0).getTime() - new Date(b.created_at || 0).getTime();
    }
    return versionSortDir === "ASC" ? cmp : -cmp;
  });

  if (loading) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">Loading document...</Text>
        </Stack>
      </Center>
    );
  }

  if (error || !document) {
    return (
      <Stack gap="lg">
        <Button
          variant="subtle"
          leftSection={<ArrowLeft className="h-4 w-4" />}
          onClick={() => router.push("/assets/documents")}>
          Back to Documents
        </Button>
        <Text c="red">{error || "Document not found"}</Text>
      </Stack>
    );
  }

  return (
    <Stack gap="lg">
      <Group justify="space-between">
        <Group gap="sm">
          <Button
            variant="subtle"
            size="sm"
            leftSection={<ArrowLeft className="h-4 w-4" />}
            onClick={() => router.push("/assets/documents")}>
            Back
          </Button>
          <Title order={2}>{document.filename}</Title>
          <Badge size="lg" variant="light" color="gray">
            v{document.version_number}
          </Badge>
        </Group>
        <Group gap="sm">
          <Button
            variant="light"
            color="lime"
            leftSection={<Download className="h-4 w-4" />}
            onClick={handleDownload}>
            Download
          </Button>
          <Button
            variant="light"
            color="red"
            leftSection={<Trash2 className="h-4 w-4" />}
            onClick={handleDeleteClick}
            disabled={showDeleteConfirm}>
            Delete
          </Button>
        </Group>
      </Group>

      {showDeleteConfirm && (
        <DeleteConfirmBanner
          itemName={document.filename}
          onConfirm={handleDeleteConfirm}
          onCancel={() => setShowDeleteConfirm(false)}
          loading={deleting}
        />
      )}

      <Paper p="lg" withBorder>
        <Stack gap="md">
          {document.description && (
            <>
              <Text size="sm" c="dimmed">
                {document.description}
              </Text>
              <Divider />
            </>
          )}

          <Group gap="xl" wrap="wrap">
            <Group gap="xs">
              <HardDrive className="h-4 w-4 text-zinc-400" />
              <Text size="sm" c="dimmed">
                Size:
              </Text>
              <Text size="sm">{formatFileSize(document.file_size)}</Text>
            </Group>
            <Group gap="xs">
              <FileText className="h-4 w-4 text-zinc-400" />
              <Text size="sm" c="dimmed">
                Type:
              </Text>
              <Text size="sm">{document.content_type}</Text>
            </Group>
            <Group gap="xs">
              <Calendar className="h-4 w-4 text-zinc-400" />
              <Text size="sm" c="dimmed">
                Uploaded:
              </Text>
              <Text size="sm">{formatDate(document.created_at)}</Text>
            </Group>
            {(document.tags || []).length > 0 && (
              <Group gap="xs">
                <Text size="sm" c="dimmed">
                  Tags:
                </Text>
                {document.tags.map((tag) => (
                  <Badge key={tag} size="sm" variant="light" color="lime">
                    {tag}
                  </Badge>
                ))}
              </Group>
            )}
          </Group>
        </Stack>
      </Paper>

      {/* Linked Entities and Version History side by side */}
      {(allEntities.length > 0 || versions.length > 0) && (
        <SimpleGrid cols={{ base: 1, md: 2 }} spacing="lg">
          {/* Linked Entities */}
          {allEntities.length > 0 && (
            <Paper p="lg" withBorder>
              <Group gap="xs" mb="md">
                <Link2 className="h-4 w-4 text-zinc-400" />
                <Text size="sm" c="dimmed">
                  Linked To ({allEntities.length})
                </Text>
              </Group>
              <TextInput
                placeholder="Search entities..."
                value={entitySearch}
                onChange={(e) => {
                  setEntitySearch(e.target.value);
                  setEntityPage(1);
                }}
                leftSection={<Search className="h-4 w-4" />}
                size="sm"
                mb="sm"
              />

              <DataTable
                data={{
                  items: sortedEntities.map((e, idx) => ({ ...e, _idx: idx })),
                  currentPage: entityPage,
                  totalPages: Math.ceil(sortedEntities.length / entityPageSize),
                  total: sortedEntities.length,
                  pageSize: entityPageSize,
                }}
                columns={[
                  {
                    header: "Type",
                    width: 120,
                    sortable: true,
                    sortKey: "entity_type",
                    render: (e) => (
                      <Badge size="sm" variant="light" color={getEntityColor(e.entity_type)}>
                        {e.entity_type}
                      </Badge>
                    ),
                  },
                  {
                    header: "Name",
                    sortable: true,
                    sortKey: "entity_name",
                    render: (e) => (
                      <Anchor
                        component={Link}
                        href={getEntityUrl(e.entity_type, e.entity_id)}
                        size="sm">
                        {e.entity_name}
                      </Anchor>
                    ),
                  },
                  {
                    header: "Version",
                    width: 80,
                    sortable: true,
                    sortKey: "version_number",
                    render: (e) => (
                      <Anchor
                        component={Link}
                        href={`/assets/documents/${e.version_id}?v=${e.version_number}`}
                        size="sm">
                        v{e.version_number}
                      </Anchor>
                    ),
                  },
                ]}
                rowKey={(e) => `${e.entity_type}-${e.entity_id}-${e.version_id}-${e._idx}`}
                onPageChange={setEntityPage}
                pageValue={entityPage}
                onPageSizeChange={(size) => {
                  setEntityPageSize(size);
                  setEntityPage(1);
                }}
                pageSizeValue={entityPageSize}
                sortBy={entitySortBy}
                sortDirection={entitySortDir}
                onSortChange={({ sortBy, direction }) => {
                  setEntitySortBy(sortBy);
                  setEntitySortDir(direction);
                }}
                emptyText={entitySearch ? "No matching entities" : "No linked entities"}
              />
            </Paper>
          )}

          {/* Version History */}
          {versions.length > 0 && (
            <Paper p="lg" withBorder>
              <Group gap="xs" mb="md">
                <History className="h-4 w-4 text-zinc-400" />
                <Text size="sm" c="dimmed">
                  Version History
                </Text>
              </Group>
              {versionsLoading ? (
                <Center mih={100}>
                  <Loader size="sm" color="lime" />
                </Center>
              ) : (
                <DataTable
                  data={{
                    items: sortedVersions,
                    currentPage: versionPage,
                    totalPages: Math.ceil(sortedVersions.length / versionPageSize),
                    total: sortedVersions.length,
                    pageSize: versionPageSize,
                  }}
                  columns={[
                    {
                      header: "Version",
                      accessor: "version_number",
                      sortable: true,
                      sortKey: "version_number",
                      render: (v) => (
                        <Group gap="xs">
                          <Text size="sm">v{v.version_number}</Text>
                          {v.id === document.id && (
                            <Badge size="xs" variant="outline" color="blue">
                              Viewing
                            </Badge>
                          )}
                        </Group>
                      ),
                    },
                    {
                      header: "Uploaded",
                      accessor: "created_at",
                      sortable: true,
                      sortKey: "created_at",
                      render: (v) => <Text size="sm">{formatDate(v.created_at)}</Text>,
                    },
                    {
                      header: "Current",
                      width: 70,
                      render: (v) => (
                        <Checkbox
                          checked={v.is_current_version}
                          onChange={() => !v.is_current_version && handleSetCurrentVersion(v.id)}
                          color="lime"
                          styles={{
                            input: {
                              cursor: v.is_current_version ? "default" : "pointer",
                            },
                          }}
                        />
                      ),
                    },
                    {
                      header: "",
                      width: 40,
                      render: (v) => (
                        <ActionIcon
                          variant="subtle"
                          color="gray"
                          size="sm"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleDownloadVersion(v);
                          }}
                          title="Download this version">
                          <Download className="h-3.5 w-3.5" />
                        </ActionIcon>
                      ),
                    },
                  ]}
                  onPageChange={setVersionPage}
                  pageValue={versionPage}
                  onPageSizeChange={(size) => {
                    setVersionPageSize(size);
                    setVersionPage(1);
                  }}
                  pageSizeValue={versionPageSize}
                  sortBy={versionSortBy}
                  sortDirection={versionSortDir}
                  onSortChange={({ sortBy, direction }) => {
                    setVersionSortBy(sortBy);
                    setVersionSortDir(direction);
                  }}
                  rowKey={(v) => v.id}
                  onRowClick={(v) =>
                    v.id !== document.id &&
                    router.push(`/assets/documents/${v.id}?v=${v.version_number}`)
                  }
                  emptyText="No versions found"
                />
              )}
            </Paper>
          )}
        </SimpleGrid>
      )}

      {/* Document Preview */}
      <Paper p="lg" withBorder>
        <Text size="sm" c="dimmed" mb="md">
          Preview
        </Text>
        {previewLoading ? (
          <Center mih={400}>
            <Loader size="md" color="lime" />
          </Center>
        ) : previewUrl ? (
          document.content_type === "application/pdf" ? (
            <iframe
              src={previewUrl}
              className="w-full h-[600px] border-0 rounded"
              title={document.filename}
            />
          ) : document.content_type.startsWith("image/") ? (
            <img src={previewUrl} alt={document.filename} className="max-w-full h-auto rounded" />
          ) : null
        ) : (
          <Center mih={200}>
            <Stack align="center" gap="sm">
              <FileText className="h-12 w-12 text-zinc-400" />
              <Text size="sm" c="dimmed">
                Preview not available for this file type
              </Text>
            </Stack>
          </Center>
        )}
      </Paper>

      {/* Version History */}
    </Stack>
  );
};

export default DocumentDetailPage;
