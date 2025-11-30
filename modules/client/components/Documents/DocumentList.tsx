"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  Stack,
  Text,
  Badge,
  Group,
  ActionIcon,
  Menu,
  TagsInput,
  TextInput,
  Loader,
  Center,
  Anchor,
} from "@mantine/core";
import { Search, X } from "lucide-react";
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
import { DataTable, type PageData, type Column } from "../DataTable/DataTable";

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
  const router = useRouter();
  const [data, setData] = useState<PageData<Document> | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filterTags, setFilterTags] = useState<string[]>(
    externalFilterTags || []
  );
  const [availableTags, setAvailableTags] = useState<string[]>([]);
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(50);
  const [sortBy, setSortBy] = useState<string>("updated_at");
  const [sortDirection, setSortDirection] = useState<"ASC" | "DESC">("DESC");

  const loadDocuments = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const pageData = await getDocuments(
        page,
        limit,
        sortBy,
        sortDirection,
        searchQuery || undefined,
        filterTags.length > 0 ? filterTags : undefined,
        entityType,
        entityId
      );
      setData(pageData);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load documents");
    } finally {
      setLoading(false);
    }
  }, [page, limit, sortBy, sortDirection, searchQuery, filterTags, entityType, entityId]);

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
  }, [loadDocuments]);

  useEffect(() => {
    loadTags();
  }, [loadTags]);

  const handleDownload = async (doc: Document, e?: React.MouseEvent) => {
    if (e) {
      e.stopPropagation();
    }
    try {
      await downloadDocument(doc.id, doc.filename);
    } catch (err) {
      console.error("Download failed:", err);
    }
  };

  const handleDelete = async (doc: Document, e?: React.MouseEvent) => {
    if (e) {
      e.stopPropagation();
    }
    if (!confirm(`Delete "${doc.filename}"?`)) return;

    try {
      await deleteDocument(doc.id);
      onDocumentDeleted?.(doc.id);
      loadDocuments();
    } catch (err) {
      console.error("Delete failed:", err);
    }
  };

  const handleRowClick = (doc: Document) => {
    router.push(`/assets/documents/${doc.id}`);
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

  const columns: Column<Document>[] = [
    {
      header: "Name",
      accessor: "filename",
      sortable: true,
      sortKey: "filename",
      render: (doc) => (
        <Group gap="xs">
          <FileText className="h-4 w-4 text-lime-600" />
          <div>
            <Text size="sm" fw={500}>
              {doc.filename}
            </Text>
            {doc.description && (
              <Text size="xs" c="dimmed" lineClamp={1}>
                {doc.description}
              </Text>
            )}
          </div>
        </Group>
      ),
    },
    {
      header: "Linked To",
      accessor: "entities",
      render: (doc) => (
        <Group gap={4}>
          {(doc.entities || []).map((entity, idx) => (
            <Anchor
              key={`${entity.entity_type}-${entity.entity_id}-${idx}`}
              component={Link}
              href={getEntityUrl(entity.entity_type, entity.entity_id)}
              size="xs"
              onClick={(e) => e.stopPropagation()}
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
      ),
    },
    {
      header: "Size",
      accessor: "file_size",
      sortable: true,
      sortKey: "file_size",
      render: (doc) => (
        <Text size="sm" c="dimmed">
          {formatFileSize(doc.file_size)}
        </Text>
      ),
    },
    {
      header: "Tags",
      accessor: "tags",
      render: (doc) => (
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
      ),
    },
    {
      header: "Uploaded",
      accessor: "created_at",
      sortable: true,
      sortKey: "created_at",
      render: (doc) => (
        <Text size="sm" c="dimmed">
          {formatDate(doc.created_at)}
        </Text>
      ),
    },
    {
      header: "Actions",
      width: 80,
      render: (doc) => (
        <Menu position="bottom-end" withinPortal>
          <Menu.Target>
            <ActionIcon
              variant="subtle"
              color="gray"
              onClick={(e) => e.stopPropagation()}
            >
              <MoreVertical className="h-4 w-4" />
            </ActionIcon>
          </Menu.Target>
          <Menu.Dropdown>
            <Menu.Item
              leftSection={<Download className="h-4 w-4" />}
              onClick={(e) => handleDownload(doc, e)}
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
              onClick={(e) => handleDelete(doc, e)}
            >
              Delete
            </Menu.Item>
          </Menu.Dropdown>
        </Menu>
      ),
    },
  ];

  if (loading && !data) {
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
        <TextInput
          placeholder="Search documents by name..."
          value={searchQuery}
          onChange={(e) => {
            setSearchQuery(e.target.value);
            setPage(1);
          }}
          leftSection={<Search className="h-4 w-4" />}
          rightSection={
            searchQuery ? (
              <ActionIcon
                size="sm"
                variant="subtle"
                onClick={() => {
                  setSearchQuery("");
                  setPage(1);
                }}
              >
                <X className="h-4 w-4" />
              </ActionIcon>
            ) : null
          }
          size="sm"
          style={{ flex: 1 }}
        />
        <TagsInput
          placeholder="Filter by tags..."
          value={filterTags}
          onChange={(tags) => {
            setFilterTags(tags);
            setPage(1);
          }}
          data={availableTags}
          clearable
          size="sm"
          style={{ flex: 1 }}
        />
        {headerRight}
      </Group>

      <DataTable
        data={data}
        columns={columns}
        onPageChange={setPage}
        pageValue={page}
        onPageSizeChange={(size) => {
          setLimit(size);
          setPage(1);
        }}
        pageSizeValue={limit}
        sortBy={sortBy}
        sortDirection={sortDirection}
        onSortChange={({ sortBy: newSortBy, direction }) => {
          setSortBy(newSortBy);
          setSortDirection(direction);
          setPage(1);
        }}
        onRowClick={handleRowClick}
        rowKey={(doc) => doc.id}
        emptyText={
          filterTags.length > 0
            ? "No documents match the selected tags."
            : "No documents yet. Upload one above!"
        }
      />
    </Stack>
  );
};
