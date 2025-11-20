"use client";

import { useEffect, useCallback } from "react";
import Chat from "../../components/Chat/Chat";
import Header from "../../components/Header";
import { usePageLoading } from "../../context/pageLoadingContext";

const ChatPage = () => {
  const { dispatchPageLoading } = usePageLoading();

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: true });
  }, [dispatchPageLoading]);

  const handleConnectionChange = useCallback(
    (isConnected: boolean) => {
      dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: !isConnected });
    },
    [dispatchPageLoading]
  );

  return (
    <>
      <Header />
      <Chat variant="page" onConnectionChange={handleConnectionChange} />
    </>
  );
};

export default ChatPage;
