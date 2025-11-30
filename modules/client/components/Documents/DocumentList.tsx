"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  Stack,
  Text,
  Table,
  Badge,
  Group,
  ActionIcon,
  Menu,
  TagsInput,
  Loader,
  Center,
  Paper,
  Anchor,
} from "@mantine/core";
import {
  Download,
  Trash2,
  MoreVertical,
  FileText,
  Link2,
  Building2,
  User,
  FolderKanban,
} from "lucide-react";
import {
  getDocuments,
  deleteDocument,
  downloadDocument,
  getTags,
  Document,
} from "../../api";

type DocumentListProps = {
  filterTags?: string[];
  entityType?: string;
  entityId?: number;
  onDocumentDeleted?: (id: number) => void;
  headerRight?: React.ReactNode;
};

export const DocumentList = ({
  filterTags: externalFilterTags,
  entityType,
  entityId,
  onDocumentDeleted,
  headerRight,
}: DocumentListProps) => {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filterTags, setFilterTags] = useState<string[]>(
    externalFilterTags || []
  );
  const [availableTags, setAvailableTags] = useState<string[]>([]);

  const loadDocuments = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const docs = await getDocuments(
        filterTags.length > 0 ? filterTags : undefined,
        entityType,
        entityId
      );
      setDocuments(docs);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load documents");
    } finally {
      setLoading(false);
    }
  }, [filterTags, entityType, entityId]);

  const loadTags = useCallback(async () => {
    try {
      const tagList = await getTags();
      setAvailableTags(tagList.map((t) => t.name));
    } catch {
      // Ignore tag loading errors
    }
  }, []);

  useEffect(() => {
    loadDocuments();
    loadTags();
  }, [loadDocuments, loadTags]);

  const handleDownload = async (doc: Document) => {
    try {
      await downloadDocument(doc.id, doc.filename);
    } catch (err) {
      console.error("Download failed:", err);
    }
  };

  const handleDelete = async (doc: Document) => {
    if (!confirm(`Delete "${doc.filename}"?`)) return;

    try {
      await deleteDocument(doc.id);
      setDocuments((prev) => prev.filter((d) => d.id !== doc.id));
      onDocumentDeleted?.(doc.id);
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
    return new Date(dateStr).toLocaleDateString();
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
      Organization: <Building2 className="h-3 w-3" />,
      Individual: <User className="h-3 w-3" />,
      Project: <FolderKanban className="h-3 w-3" />,
      Contact: <User className="h-3 w-3" />,
    };
    return iconMap[entityType] || <Link2 className="h-3 w-3" />;
  };

  if (loading) {
    return (
      <Center mih={200}>
        <Stack align="center" gap="sm">
          <Loader size="md" color="lime" />
          <Text size="sm" c="dimmed">
            Loading documents...
          </Text>
        </Stack>
      </Center>
    );
  }

  if (error) {
    return (
      <Text c="red" size="sm">
        {error}
      </Text>
    );
  }

  return (
    <Stack gap="md">
      <Group gap="md">
        <TagsInput
          placeholder="Filter by tags..."
          value={filterTags}
          onChange={setFilterTags}
          data={availableTags}
          clearable
          size="sm"
          style={{ flex: 1 }}
        />
        {headerRight}
      </Group>

      {documents.length === 0 ? (
        <Paper p="xl" withBorder>
          <Stack align="center" gap="sm">
            <FileText className="h-8 w-8 text-zinc-400" />
            <Text size="sm" c="dimmed">
              {filterTags.length > 0
                ? "No documents match the selected tags."
                : "No documents yet. Upload one above!"}
            </Text>
          </Stack>
        </Paper>
      ) : (
        <Table striped highlightOnHover>
          <Table.Thead>
            <Table.Tr>
              <Table.Th>Name</Table.Th>
              <Table.Th>Linked To</Table.Th>
              <Table.Th>Size</Table.Th>
              <Table.Th>Tags</Table.Th>
              <Table.Th>Uploaded</Table.Th>
              <Table.Th style={{ width: 80 }}>Actions</Table.Th>
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {documents.map((doc) => (
              <Table.Tr key={doc.id}>
                <Table.Td>
                  <Group gap="xs">
                    <FileText className="h-4 w-4 text-lime-600" />
                    <div>
                      <Anchor
                        component={Link}
                        href={`/assets/documents/${doc.id}`}
                        size="sm"
                        fw={500}
                        underline="hover"
                        className="hover:font-semibold transition-all"
                      >
                        {doc.filename}
                      </Anchor>
                      {doc.description && (
                        <Text size="xs" c="dimmed" lineClamp={1}>
                          {doc.description}
                        </Text>
                      )}
                    </div>
                  </Group>
                </Table.Td>
                <Table.Td>
                  <Group gap={4}>
                    {(doc.entities || []).map((entity, idx) => (
                      <Anchor
                        key={`${entity.entity_type}-${entity.entity_id}-${idx}`}
                        component={Link}
                        href={getEntityUrl(
                          entity.entity_type,
                          entity.entity_id
                        )}
                        size="xs"
                      >
                        <Badge
                          size="sm"
                          variant="light"
                          color="blue"
                          style={{ cursor: "pointer" }}
                        >
                          {entity.entity_type}
                        </Badge>
                      </Anchor>
                    ))}
                    {(!doc.entities || doc.entities.length === 0) && (
                      <Text size="xs" c="dimmed">
                        -
                      </Text>
                    )}
                  </Group>
                </Table.Td>
                <Table.Td>
                  <Text size="sm" c="dimmed">
                    {formatFileSize(doc.file_size)}
                  </Text>
                </Table.Td>
                <Table.Td>
                  <Group gap={4}>
                    {(doc.tags || []).slice(0, 3).map((tag) => (
                      <Badge key={tag} size="xs" variant="light" color="lime">
                        {tag}
                      </Badge>
                    ))}
                    {(doc.tags || []).length > 3 && (
                      <Badge size="xs" variant="light" color="gray">
                        +{doc.tags.length - 3}
                      </Badge>
                    )}
                  </Group>
                </Table.Td>
                <Table.Td>
                  <Text size="sm" c="dimmed">
                    {formatDate(doc.created_at)}
                  </Text>
                </Table.Td>
                <Table.Td>
                  <Menu position="bottom-end" withinPortal>
                    <Menu.Target>
                      <ActionIcon variant="subtle" color="gray">
                        <MoreVertical className="h-4 w-4" />
                      </ActionIcon>
                    </Menu.Target>
                    <Menu.Dropdown>
                      <Menu.Item
                        leftSection={<Download className="h-4 w-4" />}
                        onClick={() => handleDownload(doc)}
                      >
                        Download
                      </Menu.Item>
                      <Menu.Item
                        leftSection={<Link2 className="h-4 w-4" />}
                        disabled
                      >
                        Link to entity
                      </Menu.Item>
                      <Menu.Divider />
                      <Menu.Item
                        leftSection={<Trash2 className="h-4 w-4" />}
                        color="red"
                        onClick={() => handleDelete(doc)}
                      >
                        Delete
                      </Menu.Item>
                    </Menu.Dropdown>
                  </Menu>
                </Table.Td>
              </Table.Tr>
            ))}
          </Table.Tbody>
        </Table>
      )}
    </Stack>
  );
};
