"use client";

import { useState, useEffect, useCallback } from "react";
import {
  Paper,
  Title,
  Stack,
  Group,
  Button,
  Checkbox,
  Badge,
  Alert,
} from "@mantine/core";
import { Check, X, AlertTriangle } from "lucide-react";
import {
  getProposals,
  approveProposal,
  rejectProposal,
  bulkApproveProposals,
  bulkRejectProposals,
} from "../../api";
import { DataTable } from "../DataTable/DataTable";
import ProposalDetailModal from "./ProposalDetailModal";

const operationColors = {
  create: "green",
  update: "blue",
  delete: "red",
};

const ProposalQueue = () => {
  const [proposals, setProposals] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [selectedIds, setSelectedIds] = useState(new Set());
  const [activeProposal, setActiveProposal] = useState(null);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(25);
  const [totalPages, setTotalPages] = useState(1);
  const [totalCount, setTotalCount] = useState(0);
  const [sortBy, setSortBy] = useState("created_at");
  const [sortDirection, setSortDirection] = useState("DESC");
  const [bulkLoading, setBulkLoading] = useState(false);

  const loadProposals = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await getProposals(page, pageSize, sortBy, sortDirection);
      setProposals(data.items || []);
      setTotalPages(data.totalPages || 1);
      setTotalCount(data.totalCount || 0);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, sortBy, sortDirection]);

  useEffect(() => {
    loadProposals();
  }, [loadProposals]);

  const toggleSelect = (id, e) => {
    e?.stopPropagation?.();
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const toggleSelectAll = () => {
    if (selectedIds.size === proposals.length) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(proposals.map((p) => p.id)));
    }
  };

  const handleBulkApprove = async () => {
    if (selectedIds.size === 0) return;
    setBulkLoading(true);
    try {
      await bulkApproveProposals(Array.from(selectedIds));
      setSelectedIds(new Set());
      loadProposals();
    } catch (e) {
      setError(e.message);
    } finally {
      setBulkLoading(false);
    }
  };

  const handleBulkReject = async () => {
    if (selectedIds.size === 0) return;
    setBulkLoading(true);
    try {
      await bulkRejectProposals(Array.from(selectedIds));
      setSelectedIds(new Set());
      loadProposals();
    } catch (e) {
      setError(e.message);
    } finally {
      setBulkLoading(false);
    }
  };

  const handleSingleApprove = async (id) => {
    await approveProposal(id);
    loadProposals();
    setActiveProposal(null);
  };

  const handleSingleReject = async (id) => {
    await rejectProposal(id);
    loadProposals();
    setActiveProposal(null);
  };

  const hasValidationErrors = (proposal) => {
    if (!proposal.validation_errors) return false;
    if (typeof proposal.validation_errors === "string") {
      return proposal.validation_errors !== "null" && proposal.validation_errors !== "[]";
    }
    return Array.isArray(proposal.validation_errors) && proposal.validation_errors.length > 0;
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return "-";
    const date = new Date(dateStr);
    return date.toLocaleDateString() + " " + date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  };

  const handleSortChange = ({ sortBy: newSortBy, direction }) => {
    setSortBy(newSortBy);
    setSortDirection(direction);
    setPage(1);
  };

  const getSummary = (proposal) => {
    const payload = typeof proposal.payload === "string"
      ? JSON.parse(proposal.payload || "{}")
      : proposal.payload || {};

    const entityType = proposal.entity_type?.toLowerCase();

    if (entityType === "job") {
      return payload.job_title || payload.title || "-";
    }
    if (entityType === "contact") {
      const name = [payload.first_name, payload.last_name].filter(Boolean).join(" ");
      return name || payload.email || "-";
    }
    if (entityType === "individual") {
      const name = [payload.first_name, payload.last_name].filter(Boolean).join(" ");
      return name || payload.email || "-";
    }
    if (entityType === "organization") {
      return payload.name || payload.organization_name || "-";
    }
    if (entityType === "opportunity") {
      return payload.opportunity_name || payload.name || "-";
    }
    if (entityType === "partnership") {
      return payload.partnership_name || payload.name || "-";
    }
    if (entityType === "note") {
      const content = payload.content || "";
      return content.length > 50 ? content.substring(0, 50) + "..." : content || "-";
    }
    if (entityType === "task") {
      return payload.title || payload.description?.substring(0, 50) || "-";
    }

    // Fallback: try common field names
    return payload.title || payload.name || payload.job_title || "-";
  };

  const columns = [
    {
      header: (
        <Checkbox
          checked={selectedIds.size === proposals.length && proposals.length > 0}
          indeterminate={selectedIds.size > 0 && selectedIds.size < proposals.length}
          onChange={toggleSelectAll}
        />
      ),
      width: 40,
      render: (row) => (
        <div onClick={(e) => e.stopPropagation()}>
          <Checkbox
            checked={selectedIds.has(row.id)}
            onChange={(e) => toggleSelect(row.id, e)}
          />
        </div>
      ),
    },
    {
      header: "Entity Type",
      accessor: "entity_type",
      sortKey: "entity_type",
      sortable: true,
      render: (row) => <Badge variant="light">{row.entity_type}</Badge>,
    },
    {
      header: "Summary",
      render: (row) => getSummary(row),
    },
    {
      header: "Operation",
      accessor: "operation",
      sortKey: "operation",
      sortable: true,
      render: (row) => (
        <Badge color={operationColors[row.operation] || "gray"}>
          {row.operation}
        </Badge>
      ),
    },
    {
      header: "Entity ID",
      accessor: "entity_id",
      render: (row) => row.entity_id || "-",
    },
    {
      header: "Created",
      accessor: "created_at",
      sortKey: "created_at",
      sortable: true,
      render: (row) => formatDate(row.created_at),
    },
    {
      header: "Status",
      render: (row) =>
        hasValidationErrors(row) ? (
          <Badge color="orange" leftSection={<AlertTriangle size={12} />}>
            Validation Errors
          </Badge>
        ) : null,
    },
  ];

  const tableData = {
    items: proposals,
    currentPage: page,
    totalPages,
    pageSize,
    totalCount,
  };

  return (
    <Paper withBorder p="lg">
      <Stack gap="md">
        <Group justify="space-between">
          <Title order={4}>Pending Proposals</Title>
          {selectedIds.size > 0 && (
            <Group gap="xs">
              <span className="text-sm text-zinc-500">{selectedIds.size} selected</span>
              <Button
                size="xs"
                color="green"
                leftSection={<Check size={14} />}
                onClick={handleBulkApprove}
                loading={bulkLoading}
              >
                Approve
              </Button>
              <Button
                size="xs"
                color="red"
                leftSection={<X size={14} />}
                onClick={handleBulkReject}
                loading={bulkLoading}
              >
                Reject
              </Button>
            </Group>
          )}
        </Group>

        {error && (
          <Alert color="red" onClose={() => setError(null)} withCloseButton>
            {error}
          </Alert>
        )}

        <DataTable
          data={tableData}
          columns={columns}
          rowKey={(row) => row.id}
          onRowClick={(row) => setActiveProposal(row)}
          onPageChange={setPage}
          pageValue={page}
          onPageSizeChange={(size) => {
            setPageSize(size);
            setPage(1);
          }}
          pageSizeValue={pageSize}
          pageSizeOptions={[5, 10, 25, 50, 100, 250, 500]}
          sortBy={sortBy}
          sortDirection={sortDirection}
          onSortChange={handleSortChange}
          emptyText="No pending proposals"
          highlightOnHover
        />
      </Stack>

      <ProposalDetailModal
        proposal={activeProposal}
        onClose={() => setActiveProposal(null)}
        onApprove={handleSingleApprove}
        onReject={handleSingleReject}
        onPayloadUpdated={loadProposals}
      />
    </Paper>
  );
};

export default ProposalQueue;
