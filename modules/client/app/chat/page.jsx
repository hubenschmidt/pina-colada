"use client";

import { useEffect, useCallback, useState, useId } from "react";
import { env } from "next-runtime-env";
import { AlertCircle, AlertTriangle, X, ChevronDown } from "lucide-react";
import Chat from "../../components/Chat/Chat";
import AgentConfigMenu from "../../components/Chat/AgentConfigMenu";
import { usePageLoading } from "../../context/pageLoadingContext";
import { useConversationContext } from "../../context/conversationContext";
import styles from "./page.module.css";

const ChatPage = () => {
  const { dispatchPageLoading } = usePageLoading();
  const { conversationState } = useConversationContext();
  const { newChatKey } = conversationState;
  const [error, setError] = useState(null);
  const [useEvaluator, setUseEvaluator] = useState(false);
  const [toolsDropdownOpen, setToolsDropdownOpen] = useState(false);
  const toolsDropdownId = useId();
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

  useEffect(() => {
    const handleClickOutside = (event) => {
      const menu = document.getElementById(toolsDropdownId);
      if (menu && !menu.contains(event.target)) {
        setToolsDropdownOpen(false);
      }
    };

    if (toolsDropdownOpen) {
      document.addEventListener("mousedown", handleClickOutside);
      return () => document.removeEventListener("mousedown", handleClickOutside);
    }
  }, [toolsDropdownOpen, toolsDropdownId]);

  return (
    <div className={styles.chatPageContainer}>
      {error && (
        <div className={error.isWarning ? styles.warningBanner : styles.errorBanner}>
          <div className={error.isWarning ? styles.warningContent : styles.errorContent}>
            {error.isWarning ? (
              <AlertTriangle size={16} className={styles.warningIcon} />
            ) : (
              <AlertCircle size={16} className={styles.errorIcon} />
            )}
            <div className={error.isWarning ? styles.warningText : styles.errorText}>{error.message}</div>
            <button className={error.isWarning ? styles.warningDismiss : styles.errorDismiss} onClick={clearError}>
              <X size={14} />
            </button>
          </div>
          {isVerbose && error.details && (
            <pre className={styles.errorDetails}>{error.details}</pre>
          )}
        </div>
      )}
      <div className={styles.chatWrapper}>
        <div className={styles.pageToolbar}>
          <div className={styles.toolsDropdown} id={toolsDropdownId}>
            <button
              type="button"
              className={styles.toolsButton}
              onClick={() => setToolsDropdownOpen(!toolsDropdownOpen)}>
              <span>Tools</span>
              <ChevronDown size={16} className={toolsDropdownOpen ? styles.chevronOpen : ""} />
            </button>
            {toolsDropdownOpen && (
              <div className={styles.toolsMenu}>
                <label className={styles.evaluatorToggle}>
                  <input
                    type="checkbox"
                    checked={useEvaluator}
                    onChange={(e) => setUseEvaluator(e.target.checked)}
                  />
                  <span>Use evaluator</span>
                </label>
              </div>
            )}
          </div>
          <AgentConfigMenu />
        </div>
        <Chat
          key={newChatKey}
          variant="page"
          onConnectionChange={handleConnectionChange}
          error={error}
          onError={handleError}
          onClearError={clearError}
          useEvaluator={useEvaluator}
        />
      </div>
    </div>
  );
};

export default ChatPage;
