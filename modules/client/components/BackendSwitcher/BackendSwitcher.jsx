"use client";
import { useState, useEffect } from "react";
import { Select } from "@mantine/core";
import { Server } from "lucide-react";
import { checkDeveloperAccess } from "../../api";

const BACKENDS = [
  { value: "agent", label: "Python (8000)" },
  { value: "agent-go", label: "Go (8080)" },
];

const STORAGE_KEY = "pina-colada-backend";
const isDev = process.env.NODE_ENV === "development";

const BackendSwitcher = () => {
  const [isDeveloper, setIsDeveloper] = useState(isDev);
  const [backend, setBackend] = useState("agent");
  const [loading, setLoading] = useState(!isDev);

  useEffect(() => {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      setBackend(stored);
    }

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

  const handleChange = (value) => {
    setBackend(value);
    localStorage.setItem(STORAGE_KEY, value);
    window.location.reload();
  };

  if (loading || !isDeveloper) return null;

  return (
    <Select
      size="xs"
      w={130}
      leftSection={<Server size={14} />}
      data={BACKENDS}
      value={backend}
      onChange={handleChange}
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
