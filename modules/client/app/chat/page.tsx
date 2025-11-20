"use client";

import { useEffect, useCallback } from "react";
import Chat from "../../components/Chat/Chat";
<<<<<<< HEAD
import { Container } from "@mantine/core";

const ChatPage = () => {
  return (
    <Container size="xl" p={0}>
      <Chat variant="page" />
    </Container>
=======
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
>>>>>>> 6ad842a46d25e1578737259ee70a54dccfbf1861
  );
};

export default ChatPage;
