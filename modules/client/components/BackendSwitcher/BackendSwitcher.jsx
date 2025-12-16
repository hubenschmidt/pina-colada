"use client";

import { Select } from "@mantine/core";
import { Server } from "lucide-react";
import { useBackendContext } from "../../context/backendContext";
import DeveloperFeature from "../DeveloperFeature/DeveloperFeature";

const BACKENDS = [
  { value: "agent", label: "Python (8000)" },
  { value: "agent-go", label: "Go (8080)" },
];

const BackendSwitcher = () => {
  const { backend, setBackend } = useBackendContext();

  return (
    <DeveloperFeature inline>
      <Select
        size="xs"
        w={130}
        leftSection={<Server size={14} />}
        data={BACKENDS}
        value={backend}
        onChange={setBackend}
        allowDeselect={false}
        styles={{
          input: {
            fontSize: "11px",
            minHeight: "28px",
            height: "28px",
          },
        }}
      />
    </DeveloperFeature>
  );
};

export default BackendSwitcher;
