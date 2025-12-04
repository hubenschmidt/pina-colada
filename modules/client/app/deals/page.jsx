"use client";

import { useEffect } from "react";
import { Text, Stack, Center } from "@mantine/core";
import { usePageLoading } from "../../context/pageLoadingContext";

const DealsPage = () => {
  const { dispatchPageLoading } = usePageLoading();

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  return (
    <Center mih={400}>
      <Stack align="center" gap="md">
        <Text size="xl" fw={600} c="dimmed">
          Deals
        </Text>
        <Text c="dimmed">This feature is coming soon.</Text>
      </Stack>
    </Center>
  );
};

export default DealsPage;
