"use client";

import { useEffect, useCallback } from "react";
import Chat from "../../components/Chat/Chat";
import { usePageLoading } from "../../context/pageLoadingContext";

const ChatPage = () => {
  const { dispatchPageLoading } = usePageLoading();

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: true });
  }, [dispatchPageLoading]);

  const handleConnectionChange = useCallback(
    (isConnected) => {
      dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: !isConnected });
    },
    [dispatchPageLoading]
  );

  return <Chat variant="page" onConnectionChange={handleConnectionChange} />;
};

export default ChatPage;