"use client";

import { useEffect, useCallback } from "react";
import { useParams } from "next/navigation";
import Chat from "../../../components/Chat/Chat";
import { usePageLoading } from "../../../context/pageLoadingContext";

const ChatThreadPage = () => {
  const params = useParams();
  const threadId = params.threadId;
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

  return <Chat variant="page" threadId={threadId} onConnectionChange={handleConnectionChange} />;
};

export default ChatThreadPage;
