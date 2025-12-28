"use client";
import { createContext, useReducer, useContext, useCallback, useEffect } from "react";
import {
  getConversations,
  getConversation as getConversationApi,
  archiveConversation as archiveConversationApi,
  deleteConversationPermanent,
} from "../api";
import { useUserContext } from "./userContext";

const initialState = {
  conversations: [],
  activeConversation: null,
  activeThreadId: null,
  isLoading: false,
  error: null,
  newChatKey: 0,
};

const conversationReducer = (state, action) => {
  switch (action.type) {
    case "SET_CONVERSATIONS":
      return { ...state, conversations: action.payload, isLoading: false };
    case "SET_ACTIVE_CONVERSATION":
      return { ...state, activeConversation: action.payload };
    case "SET_ACTIVE_THREAD_ID":
      return { ...state, activeThreadId: action.payload };
    case "SET_LOADING":
      return { ...state, isLoading: action.payload };
    case "SET_ERROR":
      return { ...state, error: action.payload, isLoading: false };
    case "ADD_CONVERSATION":
      return {
        ...state,
        conversations: [action.payload, ...state.conversations],
      };
    case "UPDATE_CONVERSATION":
      return {
        ...state,
        conversations: state.conversations.map((c) =>
          c.thread_id === action.payload.thread_id ? { ...c, ...action.payload } : c
        ),
      };
    case "REMOVE_CONVERSATION":
      return {
        ...state,
        conversations: state.conversations.filter((c) => c.thread_id !== action.payload),
      };
    case "CLEAR_ACTIVE":
      return { ...state, activeConversation: null, activeThreadId: null };
    case "NEW_CHAT":
      return { ...state, activeConversation: null, activeThreadId: null, newChatKey: state.newChatKey + 1 };
    default:
      return state;
  }
};

export const ConversationContext = createContext({
  conversationState: initialState,
  loadConversations: async () => {},
  selectConversation: async () => {},
  createNewConversation: () => {},
  startNewChat: () => {},
  archiveConversation: async () => {},
  clearActiveConversation: () => {},
  updateConversationTitle: () => {},
});

export const useConversationContext = () => useContext(ConversationContext);

export const ConversationProvider = ({ children }) => {
  const [conversationState, dispatch] = useReducer(conversationReducer, initialState);
  const { userState } = useUserContext();

  const loadConversations = useCallback(async () => {
    if (!userState.isAuthed) return;

    dispatch({ type: "SET_LOADING", payload: true });
    try {
      const data = await getConversations();
      dispatch({ type: "SET_CONVERSATIONS", payload: data });
    } catch (error) {
      console.error("Failed to load conversations:", error);
      dispatch({ type: "SET_ERROR", payload: error.message });
    }
  }, [userState.isAuthed]);

  const selectConversation = useCallback(async (threadId) => {
    if (!threadId) {
      dispatch({ type: "CLEAR_ACTIVE" });
      return null;
    }

    dispatch({ type: "SET_LOADING", payload: true });
    try {
      const data = await getConversationApi(threadId);
      dispatch({ type: "SET_ACTIVE_CONVERSATION", payload: data });
      dispatch({ type: "SET_ACTIVE_THREAD_ID", payload: threadId });
      dispatch({ type: "SET_LOADING", payload: false });
      return data;
    } catch (error) {
      console.error("Failed to load conversation:", error);
      dispatch({ type: "SET_ERROR", payload: error.message });
      return null;
    }
  }, []);

  const createNewConversation = useCallback(() => {
    const newThreadId = crypto.randomUUID();
    dispatch({ type: "CLEAR_ACTIVE" });
    dispatch({ type: "SET_ACTIVE_THREAD_ID", payload: newThreadId });
    return newThreadId;
  }, []);

  const archiveConversation = useCallback(async (threadId) => {
    try {
      await archiveConversationApi(threadId);
      dispatch({ type: "REMOVE_CONVERSATION", payload: threadId });
      if (conversationState.activeThreadId === threadId) {
        dispatch({ type: "CLEAR_ACTIVE" });
      }
      return true;
    } catch (error) {
      console.error("Failed to archive conversation:", error);
      return false;
    }
  }, [conversationState.activeThreadId]);

  const clearActiveConversation = useCallback(() => {
    dispatch({ type: "CLEAR_ACTIVE" });
  }, []);

  const startNewChat = useCallback(() => {
    dispatch({ type: "NEW_CHAT" });
  }, []);

  const updateConversationTitle = useCallback((threadId, title) => {
    dispatch({ type: "UPDATE_CONVERSATION", payload: { thread_id: threadId, title } });
  }, []);

  // Load conversations on auth
  useEffect(() => {
    if (userState.isAuthed) {
      loadConversations();
    }
  }, [userState.isAuthed, loadConversations]);

  return (
    <ConversationContext.Provider
      value={{
        conversationState,
        loadConversations,
        selectConversation,
        createNewConversation,
        startNewChat,
        archiveConversation,
        clearActiveConversation,
        updateConversationTitle,
      }}>
      {children}
    </ConversationContext.Provider>
  );
};
