-- Create Conversation tables for chat history persistence

CREATE TABLE "Conversation" (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES "Tenant"(id),
    user_id BIGINT NOT NULL REFERENCES "User"(id),
    thread_id UUID NOT NULL UNIQUE,
    title TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    archived_at TIMESTAMPTZ
);

CREATE INDEX idx_conversation_user ON "Conversation"(user_id, archived_at);
CREATE INDEX idx_conversation_thread ON "Conversation"(thread_id);
CREATE INDEX idx_conversation_tenant ON "Conversation"(tenant_id);

CREATE TABLE "Conversation_Message" (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES "Conversation"(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('user', 'assistant')),
    content TEXT NOT NULL,
    token_usage JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE INDEX idx_conversation_message_conversation ON "Conversation_Message"(conversation_id, created_at);
