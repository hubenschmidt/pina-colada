import React, { useEffect, useRef, useState } from "react";
import { useWs, ChatMsg } from "../../hooks/useWs";
import styles from "./Chat.module.css";

const WS_URL = "ws://localhost:8000/ws";

const Chat = () => {
  const { isOpen, messages, sendMessage, reset } = useWs(WS_URL);
  const [input, setInput] = useState("");
  const [composing, setComposing] = useState(false);

  const listRef = useRef<HTMLDivElement | null>(null);
  const inputRef = useRef<HTMLTextAreaElement | null>(null);

  // Scroll only the message list, not the whole window
  useEffect(() => {
    const el = listRef.current;
    if (!el) return;
    el.scrollTop = el.scrollHeight;
  }, [messages]);

  // Focus textarea without jumping page
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
    <div className={styles.chatRoot}>
      <div className={styles.app}>
        <header className={styles.header}>
          <div
            className={`${styles.status} ${
              isOpen ? styles["status--online"] : styles["status--offline"]
            }`}
            title={isOpen ? "Connected" : "Disconnected"}
          />
          <b className={styles.title}>LangGraph-React</b>
        </header>

        <main className={styles["chat-card"]}>
          <section
            ref={listRef}
            className={styles["msg-list"]}
            onWheel={(e) => e.stopPropagation()}
            onTouchMove={(e) => e.stopPropagation()}
          >
            {messages.map((m: ChatMsg, i: number) => {
              const isUser = m.user === "User";
              return (
                <div
                  key={i}
                  className={`${styles["msg-row"]} ${
                    isUser ? styles["msg-row--user"] : styles["msg-row--bot"]
                  }`}
                >
                  <div
                    className={`${styles.bubble} ${
                      isUser ? styles["bubble--user"] : styles["bubble--bot"]
                    }`}
                  >
                    <div
                      className={`${styles["bubble__author"]} ${
                        isUser
                          ? styles["bubble__author--right"]
                          : styles["bubble__author--left"]
                      }`}
                    >
                      <strong>{m.user}</strong>
                    </div>
                    <div className={styles["bubble__text"]}>{m.msg.trim()}</div>
                  </div>
                </div>
              );
            })}
          </section>

          <form className={styles["input-form"]} onSubmit={onSubmit}>
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
              className={styles["input-form__textarea"]}
            />
            <div className={styles["input-form__actions"]}>
              <button
                type="submit"
                className={styles.btn}
                disabled={!isOpen || !input.trim()}
              >
                Send
              </button>
              <button
                type="button"
                className={`${styles.btn} ${styles["btn--ghost"]}`}
                onClick={reset}
              >
                Reset
              </button>
            </div>
          </form>
        </main>
      </div>
    </div>
  );
};

export default Chat;
