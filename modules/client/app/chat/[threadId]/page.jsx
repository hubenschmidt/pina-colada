"use client";

import { useEffect, useCallback, useState, useId } from "react";
import { useParams } from "next/navigation";
import { ChevronDown } from "lucide-react";
import Chat from "../../../components/Chat/Chat";
import AgentConfigMenu from "../../../components/Chat/AgentConfigMenu";
import { usePageLoading } from "../../../context/pageLoadingContext";
import styles from "../page.module.css";

const ChatThreadPage = () => {
  const params = useParams();
  const threadId = params.threadId;
  const { dispatchPageLoading } = usePageLoading();
  const [useEvaluator, setUseEvaluator] = useState(false);
  const [toolsDropdownOpen, setToolsDropdownOpen] = useState(false);
  const toolsDropdownId = useId();

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: true });
  }, [dispatchPageLoading]);

  const handleConnectionChange = useCallback(
    (isConnected) => {
      dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: !isConnected });
    },
    [dispatchPageLoading]
  );

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
          variant="page"
          threadId={threadId}
          onConnectionChange={handleConnectionChange}
          useEvaluator={useEvaluator}
        />
      </div>
    </div>
  );
};

export default ChatThreadPage;
