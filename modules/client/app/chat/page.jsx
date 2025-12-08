"use client";

import { useEffect, useCallback, useState } from "react";
import { env } from "next-runtime-env";
import { AlertCircle, X } from "lucide-react";
import Chat from "../../components/Chat/Chat";
import { usePageLoading } from "../../context/pageLoadingContext";
import { useConversationContext } from "../../context/conversationContext";
import styles from "./page.module.css";

const ChatPage = () => {
  const { dispatchPageLoading } = usePageLoading();
  const { conversationState } = useConversationContext();
  const { newChatKey } = conversationState;
  const [error, setError] = useState(null);
  const isVerbose = env("NEXT_PUBLIC_VERBOSE_ERRORS") === "true";

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: true });
  }, [dispatchPageLoading]);

  const handleConnectionChange = useCallback(
    (isConnected) => {
      dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: !isConnected });
    },
    [dispatchPageLoading]
  );

  const handleError = useCallback((err) => {
    setError(err);
  }, []);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  return (
    <div className={styles.chatPageContainer}>
      {error && (
        <div className={styles.errorBanner}>
          <div className={styles.errorContent}>
            <AlertCircle size={16} className={styles.errorIcon} />
            <div className={styles.errorText}>{error.message}</div>
            <button className={styles.errorDismiss} onClick={clearError}>
              <X size={14} />
            </button>
          </div>
          {isVerbose && error.details && (
            <pre className={styles.errorDetails}>{error.details}</pre>
          )}
        </div>
      )}
      <Chat
        key={newChatKey}
        variant="page"
        onConnectionChange={handleConnectionChange}
        error={error}
        onError={handleError}
        onClearError={clearError}
      />
    </div>
  );
};

export default ChatPage;
