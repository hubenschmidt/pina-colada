import React, { useEffect, useId, useState } from "react";
import { useWs, ChatMsg } from "../../hooks/useWs";
import styles from "./Chat.module.css";
import { Copy, Check, Download, ChevronDown } from "lucide-react";
import { env } from "next-runtime-env";
import { Box } from "@mantine/core";

// ---------- User context (testing-only; no consent prompts) ----------

type UserContextV1 = {
  schema: "user_context@v1";
  ua: string;
  platform?: string;
  vendor?: string;
  languages: string[];
  tz: string;
  dnt: boolean | null;
  gpc: boolean | null;
  deviceMemory?: number;
  hardwareConcurrency?: number;
  maxTouchPoints?: number;
  screen: { w: number; h: number; dpr: number; colorDepth: number };
  viewport: { w: number; h: number };
  prefersColorScheme?: "light" | "dark" | "no-preference";
  prefersReducedMotion?: boolean;
  connection?: {
    effectiveType?: string;
    downlink?: number;
    rtt?: number;
    saveData?: boolean;
  };
  url: string;
  ref: string | null;
  utm?: Record<string, string>;
  storage: { cookies: boolean; local: boolean; session: boolean };
  // Optional extras we may add later
  battery?: { charging: boolean; level?: number };
  geo?: {
    lat: number;
    lon: number;
    accuracy?: number;
    source: "geolocation" | "ip";
  };
};

const parseUTM = (u: URL) =>
  Array.from(u.searchParams.entries())
    .filter(([k]) => k.startsWith("utm_"))
    .reduce((acc, [k, v]) => ((acc[k] = v), acc), {} as Record<string, string>);

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

const getPositionOnce = (): Promise<GeolocationPosition | null> => {
  return new Promise((resolve) => {
    if (!("geolocation" in navigator)) return resolve(null);
    navigator.geolocation.getCurrentPosition(
      (pos) => resolve(pos),
      () => resolve(null),
      { enableHighAccuracy: true, timeout: 8000, maximumAge: 0 }
    );
  });
};

const buildInitialContext = async (): Promise<UserContextV1> => {
  const ctx = await withOptionalBattery(baseContext());

  // Only enrich with geo if permission is ALREADY granted.
  try {
    // permissions API might not exist; treat as "unknown"
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const perm: any = (navigator as any).permissions?.query
      ? await (navigator as any).permissions.query({
          name: "geolocation" as PermissionName,
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

const baseContext = (): UserContextV1 => {
  const nav = navigator as any;
  const url = new URL(window.location.href);

  const languages: string[] =
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
    gpc:
      typeof (nav as any).globalPrivacyControl === "boolean"
        ? (nav as any).globalPrivacyControl
        : null,
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
    prefersColorScheme: window.matchMedia?.("(prefers-color-scheme: dark)")
      .matches
      ? "dark"
      : window.matchMedia?.("(prefers-color-scheme: light)").matches
      ? "light"
      : "no-preference",
    prefersReducedMotion: !!window.matchMedia?.(
      "(prefers-reduced-motion: reduce)"
    ).matches,
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

const withOptionalBattery = async (
  ctx: UserContextV1
): Promise<UserContextV1> => {
  try {
    if ("getBattery" in navigator) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const b: any = await (navigator as any).getBattery();
      ctx.battery = {
        charging: !!b.charging,
        level: typeof b.level === "number" ? b.level : undefined,
      };
    }
  } catch {}
  return ctx;
};

// Turn bare URLs and [text](url) into clickable links; preserve newlines.
export const renderWithLinks = (text: string): React.ReactNode[] => {
  const out: React.ReactNode[] = [];
  let last = 0;
  let key = 0;

  // Matches either [label](https://url) or bare https://url
  const LINK_RE = /(\[([^\]]+)\]\((https?:\/\/[^\s)]+)\))|(https?:\/\/[^\s]+)/g;

  // Error message pattern
  const ERROR_MSG = "Sorry, there was an error generating the response.";

  const pushLineBreak = (index: number, totalLines: number) => {
    if (index >= totalLines - 1) return;
    out.push(<br key={`br-${key++}`} />);
  };

  const pushLine = (line: string, isError: boolean) => {
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

  const pushText = (chunk: string, isError: boolean = false) => {
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

    const url = (mdUrl as string) || (bareUrl as string);
    const label = (mdLabel as string) || (bareUrl as string);

    out.push(
      <a
        key={`a-${key++}`}
        href={url}
        target="_blank"
        rel="noopener noreferrer"
        className={styles.link}
      >
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

type ChatProps = {
  variant?: "page" | "embedded";
  onConnectionChange?: (isConnected: boolean) => void;
};

const Chat = ({ variant = "embedded", onConnectionChange }: ChatProps) => {
  // Log the WebSocket URL on mount for debugging
  useEffect(() => {
    console.log("WebSocket URL:", WS_URL);
  }, []);

  const { isOpen, isThinking, tokenUsage, messages, sendMessage, sendControl, reset } = useWs(WS_URL);
  const [input, setInput] = useState("");
  const [composing, setComposing] = useState(false);
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null);
  const [toolsDropdownOpen, setToolsDropdownOpen] = useState(false);
  const [demoDropdownOpen, setDemoDropdownOpen] = useState(false);
  const [isDemoMode, setIsDemoMode] = useState(false);
  const [demoIframeUrl, setDemoIframeUrl] = useState(
    `${API_URL}/mocks/401k-rollover/`
  );

  const listId = useId();
  const [hasSentContext, setHasSentContext] = useState(false);
  const toolsDropdownId = useId();
  const demoDropdownId = useId();

  // Notify parent of connection state changes
  useEffect(() => {
    onConnectionChange?.(isOpen);
  }, [isOpen, onConnectionChange]);

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

  // Close tools dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const dropdown = document.getElementById(toolsDropdownId);
      if (dropdown && !dropdown.contains(event.target as Node)) {
        setToolsDropdownOpen(false);
      }
    };

    if (toolsDropdownOpen) {
      document.addEventListener("mousedown", handleClickOutside);
      return () =>
        document.removeEventListener("mousedown", handleClickOutside);
    }
  }, [toolsDropdownOpen, toolsDropdownId]);

  // Close demo dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const dropdown = document.getElementById(demoDropdownId);
      if (dropdown && !dropdown.contains(event.target as Node)) {
        setDemoDropdownOpen(false);
      }
    };

    if (demoDropdownOpen) {
      document.addEventListener("mousedown", handleClickOutside);
      return () =>
        document.removeEventListener("mousedown", handleClickOutside);
    }
  }, [demoDropdownOpen, demoDropdownId]);

  const onSubmit: React.FormEventHandler<HTMLFormElement> = (e) => {
    e.preventDefault();
    const text = input.trim();
    if (!text) return;
    sendMessage(text);
    setInput("");
  };

  const onKeyDown: React.KeyboardEventHandler<HTMLTextAreaElement> = (e) => {
    if (composing) return;
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      onSubmit(e as unknown as React.FormEvent<HTMLFormElement>);
    }
  };

  const handleCopy = async (text: string, index: number) => {
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
      className={`${styles.chatRoot} ${
        variant === "page" ? styles.chatRootPage : ""
      } w-full ${
        isDemoMode
          ? "max-w-full"
          : variant === "embedded"
          ? "max-w-5xl"
          : "max-w-4xl"
      } mx-auto min-h-[80svh] ${
        isDemoMode ? "flex flex-row gap-4 items-start" : "flex items-center"
      } ${variant === "embedded" ? "px-4 py-6" : ""}`}
    >
      {isDemoMode && (
        <div className={styles.demoIframePanel}>
          <div className={styles.demoIframeHeader}>
            <span className={styles.demoIframeTitle}>Mock 401k Provider</span>
            <button
              className={styles.exitDemoButton}
              onClick={() => setIsDemoMode(false)}
            >
              Exit Demo
            </button>
          </div>
          <iframe
            src={demoIframeUrl}
            className={styles.demoIframe}
            title="401k Provider Demo"
          />
        </div>
      )}
      <section
        className={`${styles.shellCard} ${
          variant === "page" ? styles.shellCardFlat : ""
        } w-full`}
      >
        {/* header */}
        <header className={styles.header}>
          <div className={styles.headerLeft}>
            <div
              className={`${styles.status} ${
                isOpen ? styles.statusOnline : styles.statusOffline
              }`}
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
                title="Tools"
              >
                <span>Tools</span>
                <ChevronDown
                  size={16}
                  className={toolsDropdownOpen ? styles.chevronOpen : ""}
                />
              </button>
              {toolsDropdownOpen && (
                <div className={styles.toolsMenu}>
                  <div className="px-4 py-3 text-sm text-zinc-500 italic">
                    Big things coming!
                  </div>
                </div>
              )}
            </div>
            <button
              type="button"
              className={styles.exportButton}
              onClick={exportChat}
              disabled={!messages.length}
              title="Export chat to .txt file"
            >
              <Download size={16} />
            </button>
          </div>
        </header>

        {/* chat panel - always rendered the same */}
        <main className={styles.chatPanel}>
          {/* message list */}
          <section
            id={listId}
            className={styles.msgList}
            onWheel={(e) => e.stopPropagation()}
            onTouchMove={(e) => e.stopPropagation()}
          >
            {messages.map((m: ChatMsg, i: number) => {
              const isUser = m.user === "User";
              return (
                <div
                  key={i}
                  className={`${styles.msgRow} ${
                    isUser ? styles.msgRowUser : styles.msgRowBot
                  }`}
                >
                  <div
                    className={`${styles.msgContainer} ${
                      isUser ? styles.msgContainerUser : styles.msgContainerBot
                    }`}
                  >
                    <div
                      className={`${styles.bubble} ${
                        isUser ? styles.bubbleUser : styles.bubbleBot
                      }`}
                    >
                      <div className={styles.bubbleText}>
                        {renderWithLinks(m.msg.trim())}
                      </div>
                    </div>
                    <div
                      className={`${styles.copyButtonWrapper} ${
                        isUser
                          ? styles.copyButtonWrapperUser
                          : styles.copyButtonWrapperBot
                      }`}
                    >
                      <button
                        className={`${styles.copyButton} ${
                          copiedIndex === i ? styles.copied : ""
                        }`}
                        onClick={() => handleCopy(m.msg.trim(), i)}
                        title="Copy to clipboard"
                      >
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
            {(isThinking || tokenUsage) && (
              <div className={styles.thinkingIndicator}>
                <span className={styles.thinkingText}>
                  {isThinking ? "thinking" : ""}
                </span>
                {tokenUsage && (
                  <span className={styles.tokenUsage}>
                    {tokenUsage.current.total >= 1000
                      ? `${(tokenUsage.current.total / 1000).toFixed(1)}k`
                      : tokenUsage.current.total}
                    {" / "}
                    {tokenUsage.cumulative.total >= 1000
                      ? `${(tokenUsage.cumulative.total / 1000).toFixed(1)}k`
                      : tokenUsage.cumulative.total} tokens
                  </span>
                )}
              </div>
            )}
            <textarea
              autoFocus
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={onKeyDown}
              onCompositionStart={() => setComposing(true)}
              onCompositionEnd={() => setComposing(false)}
              placeholder="Type or paste your message..."
              rows={3}
              spellCheck
              className={styles.textarea}
            />
            <div className={styles.actions}>
              <button
                type="submit"
                className={styles.btn}
                disabled={!isOpen || !input.trim()}
              >
                Send
              </button>
              <button
                type="button"
                className={`${styles.btn} ${styles.btnGhost}`}
                onClick={reset}
              >
                Reset
              </button>
            </div>
          </form>
        </main>
      </section>
    </Box>
  );
};

export default Chat;
