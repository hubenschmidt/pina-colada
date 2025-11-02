import React, { useEffect, useRef, useState } from "react";
import { useWs, ChatMsg } from "../../hooks/useWs";
import styles from "./Chat.module.css";

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

function getPositionOnce(): Promise<GeolocationPosition | null> {
  return new Promise((resolve) => {
    if (!("geolocation" in navigator)) return resolve(null);
    navigator.geolocation.getCurrentPosition(
      (pos) => resolve(pos),
      () => resolve(null),
      { enableHighAccuracy: true, timeout: 8000, maximumAge: 0 }
    );
  });
}

async function buildInitialContext(): Promise<UserContextV1> {
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
}

function baseContext(): UserContextV1 {
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
}

async function withOptionalBattery(ctx: UserContextV1): Promise<UserContextV1> {
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
}

// Turn bare URLs and [text](url) into clickable links; preserve newlines.
function renderWithLinks(text: string): React.ReactNode[] {
  const out: React.ReactNode[] = [];
  let last = 0;
  let key = 0;

  // Matches either [label](https://url) or bare https://url
  const LINK_RE = /(\[([^\]]+)\]\((https?:\/\/[^\s)]+)\))|(https?:\/\/[^\s]+)/g;

  const pushText = (chunk: string) => {
    if (!chunk) return;
    const lines = chunk.split("\n");
    lines.forEach((line, i) => {
      if (line) out.push(line);
      if (i < lines.length - 1) out.push(<br key={`br-${key++}`} />);
    });
  };

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

  // trailing text
  pushText(text.slice(last));

  return out;
}

// ---------- Runtime configuration ----------

const getWsUrl = () => {
  if (typeof window === "undefined") return "ws://localhost:8000/ws";
  if (window.location.hostname.includes("pinacolada.co"))
    return "wss://api.pinacolada.co/ws";
  return "ws://localhost:8000/ws";
};

const WS_URL = getWsUrl();

const Chat = () => {
  // Log the WebSocket URL on mount for debugging
  useEffect(() => {
    console.log("WebSocket URL:", WS_URL);
  }, []);

  const { isOpen, messages, sendMessage, sendControl, reset } = useWs(WS_URL);
  const [input, setInput] = useState("");
  const [composing, setComposing] = useState(false);
  const [waitForWs, setWaitForWs] = useState(false);

  const listRef = useRef<HTMLDivElement | null>(null);
  const inputRef = useRef<HTMLTextAreaElement | null>(null);
  const sentCtxRef = useRef(false);

  // --- Send rich user context as soon as WS opens (testing-only) ---
  useEffect(() => {
    if (!isOpen || sentCtxRef.current) return;
    (async () => {
      const ctx = await buildInitialContext();
      sendControl({ type: "user_context", ctx }); // single, silent send
      sentCtxRef.current = true;
    })();
  }, [isOpen, sendControl]);

  // scroll the message list, not the page
  useEffect(() => {
    const el = listRef.current;
    if (!el) return;
    el.scrollTop = el.scrollHeight;
  }, [messages]);

  // hide loader when the next assistant message arrives
  useEffect(() => {
    if (!messages.length) return;
    const last = messages[messages.length - 1];
    if (last.user !== "User") setWaitForWs(false);
  }, [messages]);

  // autofocus textarea on mount
  useEffect(() => {
    inputRef.current?.focus({ preventScroll: true });
  }, []);

  const onSubmit: React.FormEventHandler<HTMLFormElement> = (e) => {
    e.preventDefault();
    const text = input.trim();
    if (!text) return;
    sendMessage(text);
    setInput("");
    setWaitForWs(true); // NEW: show typing indicator until next bot msg
  };

  const onKeyDown: React.KeyboardEventHandler<HTMLTextAreaElement> = (e) => {
    if (composing) return;
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      onSubmit(e as unknown as React.FormEvent<HTMLFormElement>);
    }
  };

  return (
    <div
      className={`${styles.chatRoot} w-full max-w-5xl mx-auto min-h-[80svh] flex items-center px-4 py-6`}
    >
      <section className={`${styles.shellCard} w-full`}>
        {/* header */}
        <header className={styles.header}>
          <div
            className={`${styles.status} ${
              isOpen ? styles.statusOnline : styles.statusOffline
            }`}
            title={isOpen ? "Connected" : "Disconnected"}
          />
          <b className={styles.title}>Chat</b>
        </header>

        {/* dark "terminal" panel */}
        <main className={styles.chatPanel}>
          {/* message list */}
          <section
            ref={listRef}
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
                    className={`${styles.bubble} ${
                      isUser ? styles.bubbleUser : styles.bubbleBot
                    }`}
                  >
                    <div
                      className={`${styles.bubbleAuthor} ${
                        isUser
                          ? styles.bubbleAuthorRight
                          : styles.bubbleAuthorLeft
                      }`}
                    >
                      <strong>{m.user}</strong>
                    </div>
                    <div className={styles.bubbleText}>
                      {renderWithLinks(m.msg.trim())}
                    </div>
                  </div>
                </div>
              );
            })}
            {/* Typing indicator (shows while waiting for next WS assistant message) */}
            {waitForWs && (
              <div className={styles.typingRow}>
                <div className={styles.typingBubble}>
                  <span className={styles.typingDots}>
                    <span className={styles.dot} />
                    <span className={styles.dot} />
                    <span className={styles.dot} />
                  </span>
                </div>
              </div>
            )}
          </section>

          {/* input */}
          <form className={styles.inputForm} onSubmit={onSubmit}>
            <textarea
              ref={inputRef}
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
    </div>
  );
};

export default Chat;
