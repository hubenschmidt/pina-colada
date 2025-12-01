"use client";

import { Paper, Group, Text, Button } from "@mantine/core";

type DeleteConfirmBannerProps = {
  itemName: string;
  onConfirm: () => void;
  onCancel: () => void;
  loading?: boolean;
  message?: string;
};

export const DeleteConfirmBanner = ({
  itemName,
  onConfirm,
  onCancel,
  loading = false,
  message,
}: DeleteConfirmBannerProps) => {
  return (
    <Paper
      p="md"
      withBorder
      radius="md"
      className="bg-red-50 dark:bg-red-950 border-red-200 dark:border-red-800"
    >
      <Group justify="space-between">
        <Text size="sm">
          {message || (
            <>
              Delete <strong>{itemName}</strong>? This cannot be undone.
            </>
          )}
        </Text>
        <Group gap="xs">
          <Button
            size="xs"
            variant="default"
            onClick={onCancel}
            disabled={loading}
          >
            Cancel
          </Button>
          <Button
            size="xs"
            color="red"
            onClick={onConfirm}
            loading={loading}
          >
            Delete
          </Button>
        </Group>
      </Group>
    </Paper>
  );
};
