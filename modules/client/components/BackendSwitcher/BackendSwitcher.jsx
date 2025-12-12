"use client";

import { useState, useEffect } from "react";
import { Select } from "@mantine/core";
import { Server } from "lucide-react";
import { checkDeveloperAccess } from "../../api";
import { useBackendContext } from "../../context/backendContext";

const BACKENDS = [
  { value: "agent", label: "Python (8000)" },
  { value: "agent-go", label: "Go (8080)" },
];

const isDev = process.env.NODE_ENV === "development";

const BackendSwitcher = () => {
  const { backend, setBackend } = useBackendContext();
  const [isDeveloper, setIsDeveloper] = useState(isDev);
  const [loading, setLoading] = useState(!isDev);

  useEffect(() => {
    if (!isDev) {
      checkDeveloperAccess()
        .then((res) => {
          setIsDeveloper(res.has_access === true);
        })
        .catch(() => {
          setIsDeveloper(false);
        })
        .finally(() => {
          setLoading(false);
        });
    }
  }, []);

  if (loading || !isDeveloper) return null;

  return (
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
  );
};

export default BackendSwitcher;
