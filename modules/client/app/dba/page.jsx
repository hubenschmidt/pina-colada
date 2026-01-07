"use client";

import { useEffect } from "react";
import { Stack, Title, Text, Tabs } from "@mantine/core";
import { ListChecks, Settings2 } from "lucide-react";
import { usePageLoading } from "../../context/pageLoadingContext";
import { useUserContext } from "../../context/userContext";
import ProposalQueue from "../../components/ProposalQueue/ProposalQueue";
import ApprovalConfigPanel from "../../components/ApprovalConfig/ApprovalConfigPanel";

const DbaPage = () => {
  const { dispatchPageLoading } = usePageLoading();
  const { userState } = useUserContext();

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  if (!userState.isAuthed) {
    return (
      <Stack gap="lg" p="lg">
        <Text>Please log in to access this page.</Text>
      </Stack>
    );
  }

  return (
    <Stack gap="lg">
      <div>
        <Title order={1}>Data Management</Title>
        <Text c="dimmed">Review agent proposals and configure entity settings.</Text>
      </div>

      <Tabs defaultValue="proposals">
        <Tabs.List>
          <Tabs.Tab value="proposals" leftSection={<ListChecks size={16} />}>
            Pending Proposals
          </Tabs.Tab>
          <Tabs.Tab value="config" leftSection={<Settings2 size={16} />}>
            Entity Settings
          </Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="proposals" pt="xl">
          <ProposalQueue />
        </Tabs.Panel>

        <Tabs.Panel value="config" pt="xl">
          <ApprovalConfigPanel />
        </Tabs.Panel>
      </Tabs>
    </Stack>
  );
};

export default DbaPage;
