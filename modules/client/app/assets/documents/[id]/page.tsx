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
} from "@mantine/core";
import { DataTable, type Column } from "../../../../components/DataTable/DataTable";
import {
  ArrowLeft,
  Download,
  Trash2,
  FileText,
  Building2,
  User,
  FolderKanban,
  Link2,
  Calendar,
  HardDrive,
  History,
  Check,
} from "lucide-react";
import {
  getDocument,
  deleteDocument,
  downloadDocument,
  getDocumentPreviewUrl,
  getDocumentVersions,
  setCurrentDocumentVersion,
  Document,
} from "../../../../api";

const DocumentDetailPage = () => {
  const params = useParams();
  const router = useRouter();
  const documentId = Number(params.id);

  const [document, setDocument] = useState<Document | null>(null);
  const [versions, setVersions] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [versionsLoading, setVersionsLoading] = useState(false);
  const [versionPage, setVersionPage] = useState(1);
  const [versionPageSize, setVersionPageSize] = useState(10);

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

    let url: string | null = null;
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

  const handleDelete = async () => {
    if (!document) return;
    if (!confirm(`Delete "${document.filename}"?`)) return;

    try {
      await deleteDocument(document.id);
      router.push("/assets/documents");
    } catch (err) {
      console.error("Delete failed:", err);
    }
  };

  const handleSetCurrentVersion = async (versionId: number) => {
    try {
      await setCurrentDocumentVersion(versionId);
      await loadDocument();
      await loadVersions();
    } catch (err) {
      console.error("Failed to set current version:", err);
    }
  };

  const handleDownloadVersion = async (version: Document) => {
    try {
      await downloadDocument(version.id, version.filename);
    } catch (err) {
      console.error("Download failed:", err);
    }
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  const formatDate = (dateStr: string | null): string => {
    if (!dateStr) return "-";
    return new Date(dateStr).toLocaleString();
  };

  const getEntityUrl = (entityType: string, entityId: number): string => {
    const typeMap: Record<string, string> = {
      Organization: `/accounts/organizations/${entityId}`,
      Individual: `/accounts/individuals/${entityId}`,
      Project: `/projects/${entityId}`,
      Contact: `/accounts/contacts/${entityId}`,
      Lead: `/leads/jobs/${entityId}`,
    };
    return typeMap[entityType] || "#";
  };

  const getEntityColor = (entityType: string): string => {
    const colorMap: Record<string, string> = {
      Organization: "blue",
      Individual: "green",
      Project: "violet",
      Contact: "cyan",
      Lead: "orange",
    };
    return colorMap[entityType] || "gray";
  };

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
          onClick={() => router.push("/assets/documents")}
        >
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
            onClick={() => router.push("/assets/documents")}
          >
            Back
          </Button>
          <Title order={2}>{document.filename}</Title>
        </Group>
        <Group gap="sm">
          <Button
            variant="light"
            color="lime"
            leftSection={<Download className="h-4 w-4" />}
            onClick={handleDownload}
          >
            Download
          </Button>
          <Button
            variant="light"
            color="red"
            leftSection={<Trash2 className="h-4 w-4" />}
            onClick={handleDelete}
          >
            Delete
          </Button>
        </Group>
      </Group>

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

          {(document.entities || []).length > 0 && (
            <Group gap="xs">
              <Text size="sm" c="dimmed">
                Linked To:
              </Text>
              {document.entities.map((entity, idx) => (
                <Anchor
                  key={`${entity.entity_type}-${entity.entity_id}-${idx}`}
                  component={Link}
                  href={getEntityUrl(entity.entity_type, entity.entity_id)}
                  size="sm"
                >
                  <Badge
                    size="sm"
                    variant="light"
                    color={getEntityColor(entity.entity_type)}
                    style={{ cursor: "pointer" }}
                  >
                    {entity.entity_name}
                  </Badge>
                </Anchor>
              ))}
            </Group>
          )}
        </Stack>
      </Paper>

      {versions.length > 1 && (
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
                items: versions,
                currentPage: versionPage,
                totalPages: Math.ceil(versions.length / versionPageSize),
                total: versions.length,
                pageSize: versionPageSize,
              }}
              columns={[
                {
                  header: "Version",
                  accessor: "version_number",
                  render: (v) => <Text size="sm">v{v.version_number}</Text>,
                },
                {
                  header: "Uploaded",
                  accessor: "created_at",
                  render: (v) => <Text size="sm">{formatDate(v.created_at)}</Text>,
                },
                {
                  header: "Size",
                  accessor: "file_size",
                  render: (v) => <Text size="sm">{formatFileSize(v.file_size)}</Text>,
                },
                {
                  header: "Current",
                  width: 80,
                  render: (v) => (
                    <Checkbox
                      checked={v.is_current_version}
                      onChange={() => !v.is_current_version && handleSetCurrentVersion(v.id)}
                      disabled={v.is_current_version}
                      color="lime"
                    />
                  ),
                },
                {
                  header: "Download",
                  width: 80,
                  render: (v) => (
                    <ActionIcon
                      variant="subtle"
                      color="gray"
                      size="sm"
                      onClick={() => handleDownloadVersion(v)}
                      title="Download this version"
                    >
                      <Download className="h-3.5 w-3.5" />
                    </ActionIcon>
                  ),
                },
              ] as Column<Document>[]}
              onPageChange={setVersionPage}
              pageValue={versionPage}
              onPageSizeChange={(size) => {
                setVersionPageSize(size);
                setVersionPage(1);
              }}
              pageSizeValue={versionPageSize}
              rowKey={(v) => v.id}
              emptyText="No versions found"
            />
          )}
        </Paper>
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
            <img
              src={previewUrl}
              alt={document.filename}
              className="max-w-full h-auto rounded"
            />
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
