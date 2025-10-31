import React, { useEffect, useRef, useState } from "react";
import { useWs, ChatMsg } from "../../hooks/useWs";
import styles from "./Chat.module.css";

// Runtime configuration - detects environment based on hostname
const getWsUrl = () => {
  // Guard against SSR - return localhost during build
  if (typeof window === "undefined") {
    return "ws://localhost:8000/ws";
  }

  // Production: detect pinacolada.co domain
  if (window.location.hostname.includes("pinacolada.co")) {
    // TODO: Replace with your actual DigitalOcean agent URL
    return "wss://pina-colada-43do6.ondigitalocean.app/ws";
  }

  // Development: localhost
  return "ws://localhost:8000/ws";
};

const WS_URL = getWsUrl();

const Chat = () => {
  const { isOpen, messages, sendMessage, reset } = useWs(WS_URL);
  const [input, setInput] = useState("");
  const [composing, setComposing] = useState(false);

  const listRef = useRef<HTMLDivElement | null>(null);
  const inputRef = useRef<HTMLTextAreaElement | null>(null);

  // scroll the message list, not the page
  useEffect(() => {
    const el = listRef.current;
    if (!el) return;
    el.scrollTop = el.scrollHeight;
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
      className={
        // outer wrapper spacing uses Tailwind so it plays nice w/ the rest of the layout
        `${styles.chatRoot} w-full max-w-5xl mx-auto -mt-4`
      }
    >
      <section
        // shell card matches global aesthetic (soft bg, subtle border)
        className={`${styles.shellCard} w-full`}
      >
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
                    <div className={styles.bubbleText}>{m.msg.trim()}</div>
                  </div>
                </div>
              );
            })}
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
