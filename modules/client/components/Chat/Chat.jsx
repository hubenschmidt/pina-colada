import React, { useEffect, useId, useState, useRef } from "react";
import { useWs } from "../../hooks/useWs";
import { useConversationContext } from "../../context/conversationContext";
import styles from "./Chat.module.css";
import { Copy, Check, Download, ChevronDown } from "lucide-react";
import { env } from "next-runtime-env";
import { Box } from "@mantine/core";

// ---------- User context (testing-only; no consent prompts) ----------

const parseUTM = (u) =>
  Array.from(u.searchParams.entries())
    .filter(([k]) => k.startsWith("utm_"))
    .reduce((acc, [k, v]) => ((acc[k] = v), acc), {});

const supports = {
  local: () => {
    try {
      localStorage.setItem("__t", "1");
      localStorage.removeItem("__t");
      return true;
    } catch {
      return false;
    }
  },
  session: () => {
    try {
      sessionStorage.setItem("__t", "1");
      sessionStorage.removeItem("__t");
      return true;
    } catch {
      return false;
    }
  },
};

const getPositionOnce = () => {
  return new Promise((resolve) => {
    if (!("geolocation" in navigator)) return resolve(null);
    navigator.geolocation.getCurrentPosition(
      (pos) => resolve(pos),
      () => resolve(null),
      { enableHighAccuracy: true, timeout: 8000, maximumAge: 0 }
    );
  });
};

const buildInitialContext = async () => {
  const ctx = await withOptionalBattery(baseContext());

  // Only enrich with geo if permission is ALREADY granted.
  try {
    // permissions API might not exist; treat as "unknown"
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const perm = navigator.permissions?.query
      ? await navigator.permissions.query({
          name: "geolocation",
        })
      : null;

    if (perm?.state === "granted") {
      const pos = await getPositionOnce();
      if (pos) {
        ctx.geo = {
          lat: pos.coords.latitude,
          lon: pos.coords.longitude,
          accuracy: pos.coords.accuracy,
          source: "geolocation",
        };
      }
    }
    // If state is "prompt" or "denied", do nothing (no second send later)
  } catch {
    // ignore permission errors; keep ctx as-is
  }
  return ctx;
};

const baseContext = () => {
  const nav = navigator;
  const url = new URL(window.location.href);

  const languages =
    Array.isArray(navigator.languages) && navigator.languages.length
      ? [...navigator.languages] // <- makes a mutable copy
      : [navigator.language];

  return {
    schema: "user_context@v1",
    ua: navigator.userAgent,
    platform: navigator.platform,
    vendor: nav.vendor,
    languages, // <- now string[]
    tz: Intl.DateTimeFormat().resolvedOptions().timeZone,
    dnt: typeof nav.doNotTrack === "string" ? nav.doNotTrack === "1" : null,
    gpc: typeof nav.globalPrivacyControl === "boolean" ? nav.globalPrivacyControl : null,
    deviceMemory: nav.deviceMemory,
    hardwareConcurrency: nav.hardwareConcurrency,
    maxTouchPoints: nav.maxTouchPoints,
    screen: {
      w: screen.width,
      h: screen.height,
      dpr: window.devicePixelRatio,
      colorDepth: screen.colorDepth,
    },
    viewport: { w: window.innerWidth, h: window.innerHeight },
    prefersColorScheme: window.matchMedia?.("(prefers-color-scheme: dark)").matches
      ? "dark"
      : window.matchMedia?.("(prefers-color-scheme: light)").matches
        ? "light"
        : "no-preference",
    prefersReducedMotion: !!window.matchMedia?.("(prefers-reduced-motion: reduce)").matches,
    connection: nav.connection
      ? {
          effectiveType: nav.connection.effectiveType,
          downlink: nav.connection.downlink,
          rtt: nav.connection.rtt,
          saveData: nav.connection.saveData,
        }
      : undefined,
    url: url.toString(),
    ref: document.referrer || null,
    utm: parseUTM(url),
    storage: {
      cookies: navigator.cookieEnabled,
      local: supports.local(),
      session: supports.session(),
    },
  };
};

const withOptionalBattery = async (ctx) => {
  try {
    if ("getBattery" in navigator) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const b = await navigator.getBattery();
      ctx.battery = {
        charging: !!b.charging,
        level: typeof b.level === "number" ? b.level : undefined,
      };
    }
  } catch {}
  return ctx;
};

// Turn bare URLs and [text](url) into clickable links; preserve newlines.
export const renderWithLinks = (text) => {
  const out = [];
  let last = 0;
  let key = 0;

  // Matches either [label](https://url) or bare https://url
  const LINK_RE = /(\[([^\]]+)\]\((https?:\/\/[^\s)]+)\))|(https?:\/\/[^\s]+)/g;

  // Error message pattern
  const ERROR_MSG = "Sorry, there was an error generating the response.";

  const pushLineBreak = (index, totalLines) => {
    if (index >= totalLines - 1) return;
    out.push(<br key={`br-${key++}`} />);
  };

  const pushLine = (line, isError) => {
    if (isError) {
      out.push(
        <span key={`error-${key++}`} className={styles.errorText}>
          {line}
        </span>
      );
      return;
    }
    out.push(line);
  };

  const pushText = (chunk, isError = false) => {
    if (!chunk) return;
    const lines = chunk.split("\n");
    lines.forEach((line, i) => {
      if (!line) {
        pushLineBreak(i, lines.length);
        return;
      }

      pushLine(line, isError);
      pushLineBreak(i, lines.length);
    });
  };

  // Check if text contains error message
  const errorIndex = text.indexOf(ERROR_MSG);

  // Process links normally (error message will be handled in trailing text)
  text.replace(LINK_RE, (match, _mdFull, mdLabel, mdUrl, bareUrl, offset) => {
    // preceding text
    pushText(text.slice(last, offset));

    const url = mdUrl || bareUrl;
    const label = mdLabel || bareUrl;

    out.push(
      <a
        key={`a-${key++}`}
        href={url}
        target="_blank"
        rel="noopener noreferrer"
        className={styles.link}>
        {label}
      </a>
    );

    last = offset + match.length;
    return match;
  });

  // trailing text (may include error message)
  const trailingText = text.slice(last);
  if (errorIndex === -1 || errorIndex < last) {
    pushText(trailingText);
    return out;
  }

  // Error message is in trailing text - split and style it
  const beforeError = trailingText.slice(0, errorIndex - last);
  const afterError = trailingText.slice(errorIndex - last + ERROR_MSG.length);

  if (beforeError) pushText(beforeError);
  pushText(ERROR_MSG, true);
  if (afterError) pushText(afterError);

  return out;
};

// ---------- Runtime configuration ----------

const API_URL = env("NEXT_PUBLIC_API_URL") || "http://localhost:8000";

const getWsUrl = () => {
  return API_URL.replace(/^http/, "ws") + "/ws";
};

const WS_URL = getWsUrl();

const Chat = ({ variant = "embedded", threadId: urlThreadId, onConnectionChange }) => {
  const { conversationState, loadConversations, selectConversation } = useConversationContext();
  const { activeConversation } = conversationState;

  const { isOpen, isThinking, tokenUsage, messages, sendMessage, sendControl, reset, loadMessages, threadId } =
    useWs(WS_URL, { threadId: urlThreadId });

  const [input, setInput] = useState("");
  const [composing, setComposing] = useState(false);
  const [copiedIndex, setCopiedIndex] = useState(null);
  const [toolsDropdownOpen, setToolsDropdownOpen] = useState(false);
  const [demoDropdownOpen, setDemoDropdownOpen] = useState(false);
  const [isDemoMode, setIsDemoMode] = useState(false);
  const [demoIframeUrl, setDemoIframeUrl] = useState(`${API_URL}/mocks/401k-rollover/`);
  const hasRefreshedRef = useRef(false);

  const listId = useId();
  const [hasSentContext, setHasSentContext] = useState(false);
  const toolsDropdownId = useId();
  const demoDropdownId = useId();

  // Notify parent of connection state changes
  useEffect(() => {
    onConnectionChange?.(isOpen);
  }, [isOpen, onConnectionChange]);

  // Load conversation from URL threadId on mount
  useEffect(() => {
    if (urlThreadId) {
      selectConversation(urlThreadId).catch(() => {
        // New conversation, no existing messages to load
      });
    }
  }, [urlThreadId, selectConversation]);

  // Load messages when a conversation is selected
  useEffect(() => {
    if (activeConversation?.messages) {
      loadMessages(activeConversation.messages);
      hasRefreshedRef.current = true; // Don't trigger refresh for loaded conversations
    }
  }, [activeConversation, loadMessages]);

  // --- Send rich user context as soon as WS opens (testing-only) ---
  useEffect(() => {
    if (!isOpen || hasSentContext) return;
    (async () => {
      const ctx = await buildInitialContext();
      sendControl({ type: "user_context", ctx }); // single, silent send
      setHasSentContext(true);
    })();
  }, [isOpen, hasSentContext, sendControl]);

  // scroll the message list, not the page
  useEffect(() => {
    const el = document.getElementById(listId);
    if (!el) return;
    el.scrollTop = el.scrollHeight;
  }, [messages, listId]);

  // After first bot response: refresh conversation list so new chat appears in sidebar
  useEffect(() => {
    const hasUserMessage = messages.some((m) => m.user === "User");
    const hasBotResponse = messages.some((m) => m.user === "PinaColada" && !m.streaming);
    const botResponseCount = messages.filter((m) => m.user === "PinaColada" && !m.streaming).length;

    // After first complete bot response to a user message
    if (hasUserMessage && hasBotResponse && botResponseCount === 2 && !hasRefreshedRef.current) {
      hasRefreshedRef.current = true;
      // Refresh conversation list so new chat appears in sidebar
      setTimeout(() => loadConversations(), 500);
    }
  }, [messages, loadConversations]);

  // Close tools dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event) => {
      const dropdown = document.getElementById(toolsDropdownId);
      if (dropdown && !dropdown.contains(event.target)) {
        setToolsDropdownOpen(false);
      }
    };

    if (toolsDropdownOpen) {
      document.addEventListener("mousedown", handleClickOutside);
      return () => document.removeEventListener("mousedown", handleClickOutside);
    }
  }, [toolsDropdownOpen, toolsDropdownId]);

  // Close demo dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event) => {
      const dropdown = document.getElementById(demoDropdownId);
      if (dropdown && !dropdown.contains(event.target)) {
        setDemoDropdownOpen(false);
      }
    };

    if (demoDropdownOpen) {
      document.addEventListener("mousedown", handleClickOutside);
      return () => document.removeEventListener("mousedown", handleClickOutside);
    }
  }, [demoDropdownOpen, demoDropdownId]);

  const onSubmit = (e) => {
    e.preventDefault();
    const text = input.trim();
    if (!text) return;
    sendMessage(text);
    setInput("");
  };

  const onKeyDown = (e) => {
    if (composing) return;
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      onSubmit(e);
    }
  };

  const handleCopy = async (text, index) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedIndex(index);
      setTimeout(() => setCopiedIndex(null), 2000);
    } catch (err) {
      console.error("Failed to copy:", err);
    }
  };

  const exportChat = () => {
    if (!messages.length) return;

    const timestamp = new Date().toISOString().replace(/[:.]/g, "-");
    const content = messages
      .map((msg) => {
        const msgTimestamp = new Date().toISOString();
        return `[${msgTimestamp}] ${msg.user}: ${msg.msg}`;
      })
      .join("\n\n");

    const blob = new Blob([content], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `pinacolada-chat-${timestamp}.txt`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <Box
      className={`${styles.chatRoot} ${variant === "page" ? styles.chatRootPage : ""} w-full ${
        isDemoMode ? "max-w-full" : variant === "embedded" ? "max-w-5xl" : "max-w-4xl"
      } mx-auto min-h-[80svh] ${
        isDemoMode ? "flex flex-row gap-4 items-start" : "flex items-center"
      } ${variant === "embedded" ? "px-4 py-6" : ""}`}>
      {isDemoMode && (
        <div className={styles.demoIframePanel}>
          <div className={styles.demoIframeHeader}>
            <span className={styles.demoIframeTitle}>Mock 401k Provider</span>
            <button className={styles.exitDemoButton} onClick={() => setIsDemoMode(false)}>
              Exit Demo
            </button>
          </div>
          <iframe src={demoIframeUrl} className={styles.demoIframe} title="401k Provider Demo" />
        </div>
      )}
      <section
        className={`${styles.shellCard} ${variant === "page" ? styles.shellCardFlat : ""} w-full`}>
        {/* header - hidden on page variant */}
        {variant !== "page" && (
          <header className={styles.header}>
            <div className={styles.headerLeft}>
              <div
                className={`${styles.status} ${isOpen ? styles.statusOnline : styles.statusOffline}`}
                title={isOpen ? "Connected" : "Disconnected"}
              />

              <b className={styles.title}>Chat</b>
            </div>
            <div className={styles.headerRight}>
              <div className={styles.toolsDropdown} id={toolsDropdownId}>
                <button
                  type="button"
                  className={styles.toolsButton}
                  onClick={() => setToolsDropdownOpen(!toolsDropdownOpen)}
                  title="Tools">
                  <span>Tools</span>
                  <ChevronDown size={16} className={toolsDropdownOpen ? styles.chevronOpen : ""} />
                </button>
                {toolsDropdownOpen && (
                  <div className={styles.toolsMenu}>
                    <div className="px-4 py-3 text-sm text-zinc-500 italic">Big things coming!</div>
                  </div>
                )}
              </div>
              <button
                type="button"
                className={styles.exportButton}
                onClick={exportChat}
                disabled={!messages.length}
                title="Export chat to .txt file">
                <Download size={16} />
              </button>
            </div>
          </header>
        )}

        {/* chat panel - always rendered the same */}
        <main className={styles.chatPanel}>
          {/* message list */}
          <section
            id={listId}
            className={styles.msgList}
            onWheel={(e) => e.stopPropagation()}
            onTouchMove={(e) => e.stopPropagation()}>
            {messages.map((m, i) => {
              const isUser = m.user === "User";
              return (
                <div
                  key={i}
                  className={`${styles.msgRow} ${isUser ? styles.msgRowUser : styles.msgRowBot}`}>
                  <div
                    className={`${styles.msgContainer} ${
                      isUser ? styles.msgContainerUser : styles.msgContainerBot
                    }`}>
                    <div
                      className={`${styles.bubble} ${
                        isUser ? styles.bubbleUser : styles.bubbleBot
                      }`}>
                      <div className={styles.bubbleText}>{renderWithLinks(m.msg.trim())}</div>
                    </div>
                    <div
                      className={`${styles.copyButtonWrapper} ${
                        isUser ? styles.copyButtonWrapperUser : styles.copyButtonWrapperBot
                      }`}>
                      <button
                        className={`${styles.copyButton} ${copiedIndex === i ? styles.copied : ""}`}
                        onClick={() => handleCopy(m.msg.trim(), i)}
                        title="Copy to clipboard">
                        {copiedIndex === i ? (
                          <>
                            <Check size={14} />
                            <span>Copied!</span>
                          </>
                        ) : (
                          <>
                            <Copy size={14} />
                          </>
                        )}
                      </button>
                    </div>
                  </div>
                </div>
              );
            })}
          </section>

          {/* input */}
          <form className={styles.inputForm} onSubmit={onSubmit}>
            {/* Typing indicator above input */}
            {(isThinking || tokenUsage || variant === "page") && (
              <div className={styles.thinkingIndicator}>
                <span className={styles.thinkingText}>{isThinking ? "thinking" : ""}</span>
                <div className={styles.indicatorRight}>
                  {tokenUsage && (
                    <span className={styles.tokenUsage}>
                      turn:{" "}
                      {tokenUsage.current.total >= 1000
                        ? `${(tokenUsage.current.total / 1000).toFixed(1)}k`
                        : tokenUsage.current.total}
                      {" · "}
                      total:{" "}
                      {tokenUsage.cumulative.total >= 1000
                        ? `${(tokenUsage.cumulative.total / 1000).toFixed(1)}k`
                        : tokenUsage.cumulative.total}
                    </span>
                  )}
                  {variant === "page" && (
                    <div
                      className={`${styles.statusSmall} ${isOpen ? styles.statusOnline : styles.statusOffline}`}
                      title={isOpen ? "Connected" : "Disconnected"}
                    />
                  )}
                </div>
              </div>
            )}
            <div className={styles.inputRow}>
              <input
                type="text"
                autoFocus
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={onKeyDown}
                onCompositionStart={() => setComposing(true)}
                onCompositionEnd={() => setComposing(false)}
                placeholder="Type or paste your message..."
                spellCheck
                className={styles.inputText}
              />
              <button type="submit" className={styles.btn} disabled={!isOpen || !input.trim()}>
                Send
              </button>
              {isThinking && (
                <button
                  type="button"
                  className={`${styles.btn} ${styles.btnStop}`}
                  onClick={() => sendControl({ type: "cancel" })}>
                  Stop
                </button>
              )}
              <button
                type="button"
                className={`${styles.btn} ${styles.btnGhost}`}
                onClick={() => {
                  reset();
                  hasRefreshedRef.current = false;
                }}>
                Reset
              </button>
              {variant === "page" && (
                <button
                  type="button"
                  className={styles.exportButton}
                  onClick={exportChat}
                  disabled={!messages.length}
                  title="Export chat to .txt file">
                  <Download size={16} />
                </button>
              )}
            </div>
          </form>
        </main>
      </section>
    </Box>
  );
};

export default Chat;
