import type { Conversation, Message } from './types.js';
import { getToken } from './auth.js';

const STORAGE_KEY = 'coahgpt-conversations';

async function authFetch(path: string, options?: RequestInit): Promise<Response> {
  const token = getToken();
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = `Bearer ${token}`;
  return fetch(path, { ...options, headers: { ...headers, ...options?.headers } });
}

function isAuthed(): boolean {
  return getToken() !== null;
}

// localStorage fallback helpers
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

export async function loadConversations(): Promise<ReadonlyArray<Conversation>> {
  if (!isAuthed()) return loadLocal();

  try {
    const res = await authFetch('/api/conversations');
    if (!res.ok) return loadLocal();
    const data = await res.json();
    const convos: ReadonlyArray<Conversation> = (data.conversations ?? []).map(
      (c: { id: string; title: string; preview: string; updated_at: string }) => ({
        id: c.id,
        messages: [],
        preview: c.title || c.preview || 'New conversation',
        createdAt: new Date(c.updated_at).getTime(),
        updatedAt: new Date(c.updated_at).getTime(),
      }),
    );
    return convos;
  } catch {
    return loadLocal();
  }
}

export async function createServerConversation(id: string): Promise<void> {
  if (!isAuthed()) return;
  try {
    await authFetch('/api/conversations', {
      method: 'POST',
      body: JSON.stringify({ id }),
    });
  } catch {
    // best effort
  }
}

export async function loadMessages(conversationId: string): Promise<ReadonlyArray<Message>> {
  if (!isAuthed()) return [];

  try {
    const res = await authFetch(`/api/conversations/${conversationId}`);
    if (!res.ok) return [];
    const data = await res.json();
    return (data.messages ?? []).map(
      (m: { role: string; content: string }, i: number) => ({
        id: `${conversationId}-${i}`,
        role: m.role as 'user' | 'assistant',
        content: m.content,
        timestamp: Date.now(),
      }),
    );
  } catch {
    return [];
  }
}

export async function saveMessage(
  conversationId: string,
  role: string,
  content: string,
): Promise<void> {
  if (!isAuthed()) return;
  try {
    await authFetch(`/api/conversations/${conversationId}/messages`, {
      method: 'POST',
      body: JSON.stringify({ role, content }),
    });
  } catch {
    // best effort
  }
}

export async function deleteServerConversation(id: string): Promise<void> {
  if (!isAuthed()) return;
  try {
    await authFetch(`/api/conversations/${id}`, { method: 'DELETE' });
  } catch {
    // best effort
  }
}

// pure functions for local state management (no mutation)
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

// keep for localStorage fallback
export { saveLocal as saveConversations, loadLocal as loadLocalConversations };
