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
  Text,
  Table,
  Loader,
  Alert,
  Pagination,
} from "@mantine/core";
import { Check, X, AlertTriangle } from "lucide-react";
import {
  getProposals,
  approveProposal,
  rejectProposal,
  bulkApproveProposals,
  bulkRejectProposals,
} from "../../api";
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
  const [totalPages, setTotalPages] = useState(1);
  const [bulkLoading, setBulkLoading] = useState(false);

  const loadProposals = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await getProposals(page, 20);
      setProposals(data.items || []);
      setTotalPages(data.total_pages || 1);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }, [page]);

  useEffect(() => {
    loadProposals();
  }, [loadProposals]);

  const toggleSelect = (id) => {
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
    try {
      await approveProposal(id);
      loadProposals();
      setActiveProposal(null);
    } catch (e) {
      setError(e.message);
    }
  };

  const handleSingleReject = async (id) => {
    try {
      await rejectProposal(id);
      loadProposals();
      setActiveProposal(null);
    } catch (e) {
      setError(e.message);
    }
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

  if (loading && proposals.length === 0) {
    return (
      <Paper withBorder p="lg">
        <Group justify="center">
          <Loader size="sm" />
          <Text>Loading proposals...</Text>
        </Group>
      </Paper>
    );
  }

  return (
    <Paper withBorder p="lg">
      <Stack gap="md">
        <Group justify="space-between">
          <Title order={4}>Pending Proposals</Title>
          {selectedIds.size > 0 && (
            <Group gap="xs">
              <Text size="sm" c="dimmed">
                {selectedIds.size} selected
              </Text>
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

        {proposals.length === 0 ? (
          <Text c="dimmed" ta="center" py="xl">
            No pending proposals
          </Text>
        ) : (
          <>
            <Table highlightOnHover>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th style={{ width: 40 }}>
                    <Checkbox
                      checked={selectedIds.size === proposals.length && proposals.length > 0}
                      indeterminate={selectedIds.size > 0 && selectedIds.size < proposals.length}
                      onChange={toggleSelectAll}
                    />
                  </Table.Th>
                  <Table.Th>Entity Type</Table.Th>
                  <Table.Th>Operation</Table.Th>
                  <Table.Th>Entity ID</Table.Th>
                  <Table.Th>Created</Table.Th>
                  <Table.Th>Status</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {proposals.map((proposal) => (
                  <Table.Tr
                    key={proposal.id}
                    style={{ cursor: "pointer" }}
                    onClick={() => setActiveProposal(proposal)}
                  >
                    <Table.Td onClick={(e) => e.stopPropagation()}>
                      <Checkbox
                        checked={selectedIds.has(proposal.id)}
                        onChange={() => toggleSelect(proposal.id)}
                      />
                    </Table.Td>
                    <Table.Td>
                      <Badge variant="light">{proposal.entity_type}</Badge>
                    </Table.Td>
                    <Table.Td>
                      <Badge color={operationColors[proposal.operation] || "gray"}>
                        {proposal.operation}
                      </Badge>
                    </Table.Td>
                    <Table.Td>{proposal.entity_id || "-"}</Table.Td>
                    <Table.Td>{formatDate(proposal.created_at)}</Table.Td>
                    <Table.Td>
                      {hasValidationErrors(proposal) && (
                        <Badge color="orange" leftSection={<AlertTriangle size={12} />}>
                          Validation Errors
                        </Badge>
                      )}
                    </Table.Td>
                  </Table.Tr>
                ))}
              </Table.Tbody>
            </Table>

            {totalPages > 1 && (
              <Group justify="center">
                <Pagination value={page} onChange={setPage} total={totalPages} />
              </Group>
            )}
          </>
        )}
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
