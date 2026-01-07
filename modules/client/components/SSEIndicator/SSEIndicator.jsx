"use client";

import { Zap } from "lucide-react";

/**
 * SSE connection status indicator
 * Shows a lightning bolt that:
 * - Green: connected and idle
 * - Green + pulsing: connected and received keep-alive
 * - Yellow: reconnecting
 * - Hidden: not connected
 */
const SSEIndicator = ({ status, size = 14 }) => {
  if (!status?.connected && !status?.reconnecting) return null;

  const isReconnecting = status?.reconnecting;
  const isPulsing = status?.pulsing;

  return (
    <Zap
      size={size}
      color={isReconnecting ? "#fab005" : "lime"}
      style={{
        opacity: isPulsing ? 1 : 0.5,
        transform: isPulsing ? "scale(1.2)" : "scale(1)",
        transition: "all 0.2s ease-out",
      }}
    />
  );
};

export default SSEIndicator;
