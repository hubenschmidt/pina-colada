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
} from "@mantine/core";
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
} from "lucide-react";
import {
  getDocument,
  deleteDocument,
  downloadDocument,
  Document,
} from "../../../../api";

const DocumentDetailPage = () => {
  const params = useParams();
  const router = useRouter();
  const documentId = Number(params.id);

  const [document, setDocument] = useState<Document | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

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

  useEffect(() => {
    loadDocument();
  }, [loadDocument]);

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
    };
    return typeMap[entityType] || "#";
  };

  const getEntityIcon = (entityType: string) => {
    const iconMap: Record<string, React.ReactNode> = {
      Organization: <Building2 className="h-4 w-4" />,
      Individual: <User className="h-4 w-4" />,
      Project: <FolderKanban className="h-4 w-4" />,
      Contact: <User className="h-4 w-4" />,
    };
    return iconMap[entityType] || <Link2 className="h-4 w-4" />;
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

          <Group gap="xl">
            <Group gap="xs">
              <HardDrive className="h-4 w-4 text-zinc-400" />
              <Text size="sm" c="dimmed">Size:</Text>
              <Text size="sm">{formatFileSize(document.file_size)}</Text>
            </Group>
            <Group gap="xs">
              <FileText className="h-4 w-4 text-zinc-400" />
              <Text size="sm" c="dimmed">Type:</Text>
              <Text size="sm">{document.content_type}</Text>
            </Group>
            <Group gap="xs">
              <Calendar className="h-4 w-4 text-zinc-400" />
              <Text size="sm" c="dimmed">Uploaded:</Text>
              <Text size="sm">{formatDate(document.created_at)}</Text>
            </Group>
          </Group>

          {(document.tags || []).length > 0 && (
            <>
              <Divider />
              <div>
                <Text size="sm" c="dimmed" mb="xs">Tags</Text>
                <Group gap="xs">
                  {document.tags.map((tag) => (
                    <Badge key={tag} size="sm" variant="light" color="lime">
                      {tag}
                    </Badge>
                  ))}
                </Group>
              </div>
            </>
          )}

          {(document.entities || []).length > 0 && (
            <>
              <Divider />
              <div>
                <Text size="sm" c="dimmed" mb="xs">Linked To</Text>
                <Stack gap="xs">
                  {document.entities.map((entity, idx) => (
                    <Anchor
                      key={`${entity.entity_type}-${entity.entity_id}-${idx}`}
                      component={Link}
                      href={getEntityUrl(entity.entity_type, entity.entity_id)}
                    >
                      <Group gap="xs">
                        {getEntityIcon(entity.entity_type)}
                        <Text size="sm">
                          {entity.entity_type} #{entity.entity_id}
                        </Text>
                      </Group>
                    </Anchor>
                  ))}
                </Stack>
              </div>
            </>
          )}
        </Stack>
      </Paper>
    </Stack>
  );
};

export default DocumentDetailPage;
