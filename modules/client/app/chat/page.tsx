"use client";

import Chat from "../../components/Chat/Chat";
import { Container } from "@mantine/core";

export default function ChatPage() {
  return (
    <Container size="xl" p={0}>
      <Chat variant="page" />
    </Container>
  );
}
