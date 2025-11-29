"use client";

import { useState } from "react";
import { Stack, TextInput, Textarea, Button, Group, Text, Select } from "@mantine/core";
import { Project, ProjectInput } from "../../api";

type ProjectFormProps = {
  project?: Project | null;
  onAdd?: (data: ProjectInput) => Promise<Project>;
  onUpdate?: (data: Partial<ProjectInput>) => Promise<Project>;
  onDelete?: () => Promise<void>;
  onClose: () => void;
};

const ProjectForm = ({
  project,
  onAdd,
  onUpdate,
  onDelete,
  onClose,
}: ProjectFormProps) => {
  const isEdit = !!project;
  const [name, setName] = useState(project?.name || "");
  const [description, setDescription] = useState(project?.description || "");
  const [status, setStatus] = useState(project?.status || "Active");
  const [startDate, setStartDate] = useState(project?.start_date || "");
  const [endDate, setEndDate] = useState(project?.end_date || "");
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!name.trim()) {
      setError("Name is required");
      return;
    }

    setSaving(true);
    try {
      const data: ProjectInput = {
        name: name.trim(),
        description: description.trim() || null,
        status: status.trim() || null,
        start_date: startDate || null,
        end_date: endDate || null,
      };

      if (isEdit && onUpdate) {
        await onUpdate(data);
      } else if (onAdd) {
        await onAdd(data);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save project");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!onDelete) return;
    if (!confirm("Are you sure you want to delete this project?")) return;

    setDeleting(true);
    try {
      await onDelete();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete project");
      setDeleting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <Stack gap="lg">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
          {isEdit ? "Edit Project" : "New Project"}
        </h1>

        {error && (
          <Text c="red" size="sm">
            {error}
          </Text>
        )}

        <TextInput
          label="Name"
          placeholder="Project name"
          value={name}
          onChange={(e) => setName(e.currentTarget.value)}
          required
        />

        <Textarea
          label="Description"
          placeholder="Project description"
          value={description}
          onChange={(e) => setDescription(e.currentTarget.value)}
          rows={3}
        />

        <Select
          label="Status"
          data={[
            { value: "Active", label: "Active" },
            { value: "Inactive", label: "Inactive" },
          ]}
          value={status}
          onChange={(value) => setStatus(value || "Active")}
        />

        <Group grow>
          <TextInput
            label="Start Date"
            type="date"
            value={startDate}
            onChange={(e) => setStartDate(e.currentTarget.value)}
          />
          <TextInput
            label="End Date"
            type="date"
            value={endDate}
            onChange={(e) => setEndDate(e.currentTarget.value)}
          />
        </Group>

        <Group justify="space-between">
          <Group>
            <Button variant="subtle" onClick={onClose} disabled={saving || deleting}>
              Cancel
            </Button>
            {isEdit && onDelete && (
              <Button
                color="red"
                variant="subtle"
                onClick={handleDelete}
                loading={deleting}
                disabled={saving}
              >
                Delete
              </Button>
            )}
          </Group>
          <Button type="submit" loading={saving} disabled={deleting}>
            {isEdit ? "Save Changes" : "Create Project"}
          </Button>
        </Group>
      </Stack>
    </form>
  );
};

export default ProjectForm;
