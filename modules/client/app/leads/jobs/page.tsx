"use client";

import LeadTracker from "../../../components/LeadTracker/LeadTracker";
import { useLeadConfig } from "../../../components/config";
import { Container } from "@mantine/core";

const JobsPage = () => {
  const jobConfig = useLeadConfig("job");

<<<<<<< HEAD
  return (
    <Container size="xl" p={0}>
      <LeadTracker config={jobConfig} />
    </Container>
  );
=======
  return <LeadTracker config={jobConfig} />;
>>>>>>> 6ad842a46d25e1578737259ee70a54dccfbf1861
};

export default JobsPage;
