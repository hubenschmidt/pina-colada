"use client";

import Chat from "../../components/Chat/Chat";
import { Container } from "@mantine/core";

const ChatPage = () => {
  return (
    <Container size="xl" p={0}>
      <Chat variant="page" />
    </Container>
  );
};

export default ChatPage;
