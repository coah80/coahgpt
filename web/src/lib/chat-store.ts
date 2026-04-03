import type { Conversation, Message } from './types.js';

const STORAGE_KEY = 'coahgpt-conversations';

function loadLocal(): ReadonlyArray<Conversation> {
  if (typeof window === 'undefined') return [];
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed;
  } catch {
    return [];
  }
}

function saveLocal(conversations: ReadonlyArray<Conversation>): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(conversations));
  } catch {
    // storage full or unavailable
  }
}

export function loadConversations(): ReadonlyArray<Conversation> {
  return loadLocal();
}

export function createConversation(id: string): Conversation {
  return {
    id,
    messages: [],
    preview: 'New conversation',
    createdAt: Date.now(),
    updatedAt: Date.now(),
  };
}

export function addMessage(
  conversation: Conversation,
  message: Message,
): Conversation {
  const preview =
    message.role === 'user'
      ? message.content.slice(0, 50)
      : conversation.preview;

  return {
    ...conversation,
    messages: [...conversation.messages, message],
    preview,
    updatedAt: Date.now(),
  };
}

export function updateLastAssistantMessage(
  conversation: Conversation,
  content: string,
): Conversation {
  const messages = [...conversation.messages];
  const lastIndex = messages.length - 1;

  if (lastIndex >= 0 && messages[lastIndex].role === 'assistant') {
    messages[lastIndex] = { ...messages[lastIndex], content };
  }

  return {
    ...conversation,
    messages,
    updatedAt: Date.now(),
  };
}

export function removeConversation(
  conversations: ReadonlyArray<Conversation>,
  id: string,
): ReadonlyArray<Conversation> {
  return conversations.filter((c) => c.id !== id);
}

export function upsertConversation(
  conversations: ReadonlyArray<Conversation>,
  updated: Conversation,
): ReadonlyArray<Conversation> {
  const exists = conversations.some((c) => c.id === updated.id);
  if (exists) {
    return conversations.map((c) => (c.id === updated.id ? updated : c));
  }
  return [updated, ...conversations];
}

export { saveLocal as saveConversations, loadLocal as loadLocalConversations };
