"use client";

import LeadTracker from "../../../components/LeadTracker/LeadTracker";
import { useLeadConfig } from "../../../components/config";
import { Container } from "@mantine/core";

const JobsPage = () => {
  const jobConfig = useLeadConfig("job");

  return (
    <Container size="xl" p={0}>
      <LeadTracker config={jobConfig} />
    </Container>
  );
};

export default JobsPage;
