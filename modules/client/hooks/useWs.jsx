import { useEffect, useMemo, useState, useCallback, useRef, useContext } from "react";
import { UserContext } from "../context/userContext";

const GREETING = "Welcome to your CRM assistant. I can help you look up contacts, organizations, accounts, and documents. What would you like to find?";

/** Apply a streaming chunk to chat state (guard-clause style). */
const applyStreamChunk = (prev, chunk) => {
  const last = prev.at(-1);
  const isStreamingBot = !!(last && last.user === "PinaColada" && last.streaming);

  // if there isn't an active streaming bot bubble, start one
  if (!isStreamingBot) {
    return [...prev, { user: "PinaColada", msg: chunk, streaming: true }];
  }

  // otherwise append to the current streaming bubble
  const next = prev.slice();
  next[next.length - 1] = { ...last, msg: last.msg + chunk };
  return next;
};

/** Close the active streaming bubble if one exists (guard-clause style). */
const applyEndOfTurn = (prev) => {
  const last = prev.at(-1);
  // nothing to do if last isn't a streaming bot bubble
  if (!(last && last.user === "PinaColada" && last.streaming)) return prev;

  const next = prev.slice();
  next[next.length - 1] = { ...last, streaming: false };
  return next;
};

/** Handle new response start - close current bubble if streaming */
const applyStartOfTurn = (prev) => {
  const last = prev.at(-1);
  // If there's an active streaming bubble, close it before starting new one
  if (last && last.user === "PinaColada" && last.streaming) {
    const next = prev.slice();
    next[next.length - 1] = { ...last, streaming: false };
    return next;
  }
  return prev;
};

const parseTokenUsage = (obj) => {
  if (!obj.on_token_usage || typeof obj.on_token_usage !== "object") return null;
  const current = obj.on_token_usage;
  const cumulative = obj.on_token_cumulative || {};
  return {
    current: {
      input: current.input || 0,
      output: current.output || 0,
      total: current.total || 0,
    },
    cumulative: {
      input: cumulative.input || 0,
      output: cumulative.output || 0,
      total: cumulative.total || 0,
    },
  };
};

const handleParsedMessage = (obj, ctx) => {
  // Start of new assistant turn
  if (obj.on_chat_model_start === true) {
    ctx.setIsThinking(true);
    ctx.setTokenUsage(null);
    ctx.setMessages((prev) => applyStartOfTurn(prev));
    return true;
  }

  // Token usage update
  const tokenData = parseTokenUsage(obj);
  if (tokenData) {
    ctx.setTokenUsage(tokenData);
    return true;
  }

  // Streaming chunk
  const chunk = obj.on_chat_model_stream;
  if (typeof chunk === "string" && chunk.length > 0) {
    ctx.setMessages((prev) => applyStreamChunk(prev, chunk));
    return true;
  }

  // End of assistant turn
  if (obj.on_chat_model_end === true) {
    ctx.setIsThinking(false);
    ctx.setMessages((prev) => applyEndOfTurn(prev));
    return true;
  }

  // Cancelled by user
  if (obj.type === "cancelled") {
    ctx.setIsThinking(false);
    ctx.setMessages((prev) => {
      const closed = applyEndOfTurn(prev);
      return [...closed, { user: "PinaColada", msg: "⏹️ Generation stopped." }];
    });
    return true;
  }

  // Error from backend
  if (obj.type === "error") {
    ctx.setIsThinking(false);
    ctx.setMessages((prev) => applyEndOfTurn(prev));
    if (ctx.onError) {
      ctx.onError(obj.message || "An error occurred", obj.details || obj.stack);
    }
    return true;
  }

  return false;
};

const handleUiEvents = (obj, ctx) => {
  const uiKeys = Object.keys(obj).filter(
    (k) => k.startsWith("on_ui_") && k !== "on_chat_model_stream" && k !== "on_chat_model_end"
  );
  if (uiKeys.length === 0) return;
  const k = uiKeys[0];
  ctx.setMessages((prev) => [
    ...prev,
    { user: "PinaColada", msg: `🔔 ${k}: ${JSON.stringify(obj[k])}` },
  ]);
};

export const useWs = (url, { threadId: initialThreadId = null, onError = null } = {}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [isThinking, setIsThinking] = useState(false);
  const [tokenUsage, setTokenUsage] = useState(null);
  const [messages, setMessages] = useState([
    {
      user: "PinaColada",
      msg: GREETING,
    },
  ]);
  const wsRef = useRef(null);

  // Get user context for user_id and tenant_id
  const { userState } = useContext(UserContext);
  const userId = userState?.user?.id;
  const tenantId = userState?.user?.tenant_id;

  // Thread ID - use provided or generate new one on mount
  const [uuid] = useState(() => initialThreadId || crypto.randomUUID());

  useEffect(() => {
    const socket = new WebSocket(url);
    wsRef.current = socket;

    socket.onopen = () => {
      // Restore wsRef in case it was cleared by a premature close
      wsRef.current = socket;
      setIsOpen(true);
      socket.send(JSON.stringify({ uuid, init: true }));
    };

    socket.onclose = () => {
      setIsOpen(false);
      // Only clear the ref if this is the current socket
      if (wsRef.current === socket) {
        wsRef.current = null;
      }
    };

    socket.onerror = () => setIsOpen(false);

    socket.onmessage = (event) => {
      if (typeof event.data !== "string") return;

      let parsed;
      try {
        parsed = JSON.parse(event.data);
      } catch (e) {
        // Plain text from server -> bot bubble
        setMessages((prev) => [...prev, { user: "PinaColada", msg: event.data }]);
        return;
      }

      if (parsed === null || typeof parsed !== "object") return;
      const obj = parsed;
      const ctx = { setIsThinking, setTokenUsage, setMessages, onError };

      if (handleParsedMessage(obj, ctx)) return;
      handleUiEvents(obj, ctx);
    };

    return () => {
      try {
        socket.close();
      } catch (e2) {}
      wsRef.current = null;
    };
  }, [url, uuid]);

  const sendMessage = useCallback(
    (text) => {
      const t = text.trim();
      if (!t) return;
      const s = wsRef.current;
      if (!s || s.readyState !== WebSocket.OPEN) return;

      setMessages((prev) => [...prev, { user: "User", msg: t }]);
      setIsThinking(true); // Start thinking when user sends message

      // Include user_id and tenant_id for message persistence
      const payload = { uuid, message: t };
      if (userId) payload.user_id = userId;
      if (tenantId) payload.tenant_id = tenantId;

      s.send(JSON.stringify(payload));
    },
    [uuid, userId, tenantId]
  );

  // NEW: silent JSON sender (no UI echo)
  const sendControl = useCallback(
    (payload) => {
      const s = wsRef.current;
      if (!s || s.readyState !== WebSocket.OPEN) return;
      s.send(typeof payload === "string" ? payload : JSON.stringify({ uuid, ...(payload || {}) }));
    },
    [uuid]
  );

  const reset = useCallback(() => {
    setMessages([
      {
        user: "PinaColada",
        msg: GREETING,
      },
    ]);
  }, []);

  const loadMessages = useCallback((conversationMessages) => {
    if (!conversationMessages?.length) {
      setMessages([{ user: "PinaColada", msg: GREETING }]);
      return;
    }
    const formatted = conversationMessages.map((m) => ({
      user: m.role === "user" ? "User" : "PinaColada",
      msg: m.content,
    }));
    setMessages(formatted);
  }, []);

  return {
    isOpen,
    isThinking,
    tokenUsage,
    messages,
    sendMessage,
    sendControl,
    reset,
    loadMessages,
    threadId: uuid,
  };
};
