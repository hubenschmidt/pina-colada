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
  Loader,
  Center,
  Anchor,
} from "@mantine/core";
import SearchBox from "../SearchBox/SearchBox";
import { Download, Trash2, MoreVertical, FileText, Link2 } from "lucide-react";
import { getDocuments, deleteDocument, downloadDocument, getTags } from "../../api";
import { DataTable } from "../DataTable/DataTable";
import { DeleteConfirmBanner } from "../DeleteConfirmBanner/DeleteConfirmBanner";

export const DocumentList = ({
  filterTags: externalFilterTags,
  entityType,
  entityId,
  onDocumentDeleted,
  headerRight,
}) => {
  const router = useRouter();
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [filterTags, setFilterTags] = useState(externalFilterTags || []);
  const [availableTags, setAvailableTags] = useState([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [page, setPage] = useState(1);
  const [deleteConfirm, setDeleteConfirm] = useState(null);
  const [deleting, setDeleting] = useState(false);
  const [limit, setLimit] = useState(50);
  const [sortBy, setSortBy] = useState("updated_at");
  const [sortDirection, setSortDirection] = useState("DESC");

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

  const handleDownload = async (doc, e) => {
    if (e) {
      e.stopPropagation();
    }
    try {
      await downloadDocument(doc.id, doc.filename);
    } catch (err) {
      console.error("Download failed:", err);
    }
  };

  const handleDeleteClick = (doc, e) => {
    if (e) {
      e.stopPropagation();
    }
    setDeleteConfirm(doc);
  };

  const handleDeleteConfirm = async () => {
    if (!deleteConfirm) return;

    setDeleting(true);
    try {
      await deleteDocument(deleteConfirm.id);
      onDocumentDeleted?.(deleteConfirm.id);
      setDeleteConfirm(null);
      loadDocuments();
    } catch (err) {
      console.error("Delete failed:", err);
    } finally {
      setDeleting(false);
    }
  };

  const handleDeleteCancel = () => {
    setDeleteConfirm(null);
  };

  const handleRowClick = (doc) => {
    router.push(`/assets/documents/${doc.id}?v=${doc.version_number}`);
  };

  const handleSearch = (query) => {
    setSearchQuery(query);
    setPage(1);
  };

  const formatFileSize = (bytes) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return "-";
    return new Date(dateStr).toLocaleDateString();
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

  const columns = [
    {
      header: "Name",
      accessor: "filename",
      sortable: true,
      sortKey: "filename",
      render: (doc) => (
        <Group gap="xs">
          <FileText className="h-4 w-4 text-lime-600" />
          <div>
            <Group gap="xs">
              <Text size="sm" fw={500}>
                {doc.filename}
              </Text>
              <Badge size="xs" variant="light" color="gray">
                v{doc.version_number}
              </Badge>
            </Group>
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
      render: (doc) => {
        // Filter out Lead entities to avoid clutter (documents like resumes link to many jobs)
        const filteredEntities = (doc.entities || []).filter((e) => e.entity_type !== "Lead");
        return (
          <Group gap={4}>
            {filteredEntities.map((entity, idx) => (
              <Anchor
                key={`${entity.entity_type}-${entity.entity_id}-${idx}`}
                component={Link}
                href={getEntityUrl(entity.entity_type, entity.entity_id)}
                size="xs"
                onClick={(e) => e.stopPropagation()}>
                <Badge
                  size="sm"
                  variant="light"
                  color={getEntityColor(entity.entity_type)}
                  style={{ cursor: "pointer" }}>
                  {entity.entity_name}
                </Badge>
              </Anchor>
            ))}
            {filteredEntities.length === 0 && (
              <Text size="xs" c="dimmed">
                -
              </Text>
            )}
          </Group>
        );
      },
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
            <ActionIcon variant="subtle" color="gray" onClick={(e) => e.stopPropagation()}>
              <MoreVertical className="h-4 w-4" />
            </ActionIcon>
          </Menu.Target>
          <Menu.Dropdown>
            <Menu.Item
              leftSection={<Download className="h-4 w-4" />}
              onClick={(e) => handleDownload(doc, e)}>
              Download
            </Menu.Item>
            <Menu.Item leftSection={<Link2 className="h-4 w-4" />} disabled>
              Link to entity
            </Menu.Item>
            <Menu.Divider />
            <Menu.Item
              leftSection={<Trash2 className="h-4 w-4" />}
              color="red"
              onClick={(e) => handleDeleteClick(doc, e)}>
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
        <SearchBox placeholder="Search documents... (Enter to search)" onSearch={handleSearch} />

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

      {deleteConfirm && (
        <DeleteConfirmBanner
          itemName={deleteConfirm.filename}
          onConfirm={handleDeleteConfirm}
          onCancel={handleDeleteCancel}
          loading={deleting}
        />
      )}

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
