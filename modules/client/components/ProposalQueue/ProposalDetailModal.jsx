"use client";

import { useState, useEffect, useCallback } from "react";
import { Modal, Stack, Group, Badge, Button, Alert, Text, List } from "@mantine/core";
import { Check, X, Save } from "lucide-react";
import JsonEditor from "../JsonEditor/JsonEditor";
import { updateProposalPayload } from "../../api";

const operationColors = {
  create: "green",
  update: "blue",
  delete: "red",
};

const ProposalDetailModal = ({ proposal, onClose, onApprove, onReject, onPayloadUpdated }) => {
  const [payload, setPayload] = useState(null);
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState(null);
  const [executeError, setExecuteError] = useState(null);
  const [hasChanges, setHasChanges] = useState(false);
  const [approving, setApproving] = useState(false);
  const [rejecting, setRejecting] = useState(false);

  useEffect(() => {
    if (proposal?.payload) {
      try {
        const parsed =
          typeof proposal.payload === "string" ? JSON.parse(proposal.payload) : proposal.payload;
        setPayload(parsed);
      } catch {
        setPayload(proposal.payload);
      }
    }
    setHasChanges(false);
    setSaveError(null);
    setExecuteError(null);
  }, [proposal]);

  const handleApprove = async () => {
    setApproving(true);
    setExecuteError(null);
    try {
      await onApprove(proposal.id);
    } catch (e) {
      setExecuteError(e.message);
    } finally {
      setApproving(false);
    }
  };

  const handleReject = async () => {
    setRejecting(true);
    setExecuteError(null);
    try {
      await onReject(proposal.id);
    } catch (e) {
      setExecuteError(e.message);
    } finally {
      setRejecting(false);
    }
  };

  const handlePayloadChange = useCallback((newPayload) => {
    setPayload(newPayload);
    setHasChanges(true);
  }, []);

  const handleSave = async () => {
    if (!proposal || !hasChanges) return;
    setSaving(true);
    setSaveError(null);
    try {
      await updateProposalPayload(proposal.id, payload);
      setHasChanges(false);
      onPayloadUpdated?.();
    } catch (e) {
      setSaveError(e.message);
    } finally {
      setSaving(false);
    }
  };

  const getValidationErrors = () => {
    if (!proposal?.validation_errors) return [];
    if (typeof proposal.validation_errors === "string") {
      try {
        const parsed = JSON.parse(proposal.validation_errors);
        return Array.isArray(parsed) ? parsed : [];
      } catch {
        return [];
      }
    }
    return Array.isArray(proposal.validation_errors) ? proposal.validation_errors : [];
  };

  const validationErrors = getValidationErrors();
  const hasValidationErrors = validationErrors.length > 0;

  if (!proposal) return null;

  return (
    <Modal opened={!!proposal} onClose={onClose} title="Review Proposal" size="lg">
      <Stack gap="md">
        <Group>
          <Badge variant="light">{proposal.entity_type}</Badge>
          <Badge color={operationColors[proposal.operation] || "gray"}>{proposal.operation}</Badge>
          {proposal.entity_id && (
            <Text size="sm" c="dimmed">
              ID: {proposal.entity_id}
            </Text>
          )}
        </Group>

        {hasValidationErrors && (
          <Alert color="orange" title="Validation Errors">
            <List size="sm">
              {validationErrors.map((err, i) => (
                <List.Item key={i}>
                  <Text size="sm">
                    <strong>{err.field}</strong>: {err.message}
                  </Text>
                </List.Item>
              ))}
            </List>
          </Alert>
        )}

        {saveError && (
          <Alert color="red" onClose={() => setSaveError(null)} withCloseButton>
            {saveError}
          </Alert>
        )}

        {executeError && (
          <Alert color="red" title="Execution Failed" onClose={() => setExecuteError(null)} withCloseButton>
            {executeError}
          </Alert>
        )}

        <JsonEditor value={payload} onChange={handlePayloadChange} minRows={12} />

        <Group justify="space-between">
          <Button
            variant="outline"
            leftSection={<Save size={16} />}
            onClick={handleSave}
            loading={saving}
            disabled={!hasChanges}
          >
            Save Changes
          </Button>

          <Group gap="xs">
            <Button color="red" leftSection={<X size={16} />} onClick={handleReject} loading={rejecting}>
              Reject
            </Button>
            <Button
              color="green"
              leftSection={<Check size={16} />}
              onClick={handleApprove}
              disabled={hasValidationErrors}
              loading={approving}
            >
              Approve
            </Button>
          </Group>
        </Group>

        {hasValidationErrors && (
          <Text size="xs" c="dimmed" ta="center">
            Fix validation errors before approving
          </Text>
        )}
      </Stack>
    </Modal>
  );
};

export default ProposalDetailModal;
