<script lang="ts">
  import { onMount, tick } from 'svelte';
  import type { Conversation, Message, StreamToken } from '$lib/types.js';
  import {
    loadConversations,
    createConversation,
    addMessage,
    updateLastAssistantMessage,
    removeConversation,
    upsertConversation,
    saveConversations,
  } from '$lib/chat-store.js';
  import Sidebar from '$lib/components/Sidebar.svelte';
  import MessageBubble from '$lib/components/MessageBubble.svelte';
  import CatLogo from '$lib/components/CatLogo.svelte';
  import PurpleCat from '$lib/components/PurpleCat.svelte';
  import CyclingText from '$lib/components/CyclingText.svelte';

  let conversations = $state<ReadonlyArray<Conversation>>([]);
  let activeId = $state<string | null>(null);
  let input = $state('');
  let streaming = $state(false);
  let sessionId = $state<string | null>(null);
  let sidebarCollapsed = $state(true);
  let messagesEnd: HTMLDivElement | undefined = $state();
  let textareaEl: HTMLTextAreaElement | undefined = $state();

  const activeConversation = $derived(
    conversations.find((c) => c.id === activeId) ?? null
  );
  const messages = $derived(activeConversation?.messages ?? []);

  onMount(() => {
    conversations = loadConversations();
  });

  function generateId(): string {
    return Math.random().toString(36).slice(2) + Date.now().toString(36);
  }

  function scrollToBottom() {
    tick().then(() => {
      messagesEnd?.scrollIntoView({ behavior: 'smooth' });
    });
  }

  function newChat() {
    const conv = createConversation(generateId());
    conversations = [conv, ...conversations];
    activeId = conv.id;
    sessionId = null;
    input = '';
    sidebarCollapsed = true;
    saveConversations(conversations);
  }

  function selectChat(id: string) {
    activeId = id;
    sessionId = null;
    sidebarCollapsed = true;
    scrollToBottom();
  }

  function deleteChat(id: string) {
    conversations = removeConversation(conversations, id);
    if (activeId === id) {
      activeId = conversations.length > 0 ? conversations[0].id : null;
      sessionId = null;
    }
    saveConversations(conversations);
  }

  async function sendMessage() {
    const text = input.trim();
    if (!text || streaming) return;

    if (!activeId) {
      newChat();
    }

    const userMsg: Message = {
      id: generateId(),
      role: 'user',
      content: text,
      timestamp: Date.now(),
    };

    let conv = activeConversation!;
    conv = addMessage(conv, userMsg);
    conversations = upsertConversation(conversations, conv);
    saveConversations(conversations);
    input = '';
    resizeTextarea();
    scrollToBottom();

    const assistantMsg: Message = {
      id: generateId(),
      role: 'assistant',
      content: '',
      timestamp: Date.now(),
    };
    conv = addMessage(conv, assistantMsg);
    conversations = upsertConversation(conversations, conv);

    streaming = true;

    try {
      const resp = await fetch('/api/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          session_id: sessionId ?? '',
          message: text,
        }),
      });

      if (!resp.ok || !resp.body) {
        conv = updateLastAssistantMessage(conv, '[error: failed to connect]');
        conversations = upsertConversation(conversations, conv);
        streaming = false;
        return;
      }

      const reader = resp.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';
      let content = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() ?? '';

        for (const line of lines) {
          if (!line.startsWith('data: ')) continue;
          const json = line.slice(6).trim();
          if (!json) continue;

          try {
            const event: StreamToken & { thinking?: string } = JSON.parse(json);

            if (event.token) {
              content += event.token;
              conv = updateLastAssistantMessage(conv, content);
              conversations = upsertConversation(conversations, conv);
              scrollToBottom();
            }

            if (event.session_id) {
              sessionId = event.session_id;
            }

            if (event.done) break;
          } catch {
            // malformed chunk, skip
          }
        }
      }
    } catch {
      conv = updateLastAssistantMessage(conv, conv.messages.at(-1)?.content + '\n\n[connection lost]');
      conversations = upsertConversation(conversations, conv);
    }

    streaming = false;
    saveConversations(conversations);
    scrollToBottom();
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  }

  function resizeTextarea() {
    if (!textareaEl) return;
    textareaEl.style.height = 'auto';
    textareaEl.style.height = Math.min(textareaEl.scrollHeight, 200) + 'px';
  }
</script>

<svelte:head>
  <link
    rel="stylesheet"
    href="https://fonts.googleapis.com/css2?family=Playfair+Display:wght@700;800;900&display=swap"
  />
</svelte:head>

<div class="app">
  <Sidebar
    {conversations}
    {activeId}
    collapsed={sidebarCollapsed}
    onNewChat={newChat}
    onSelectChat={selectChat}
    onDeleteChat={deleteChat}
    onToggleCollapse={() => { sidebarCollapsed = !sidebarCollapsed; }}
  />

  <main class="main">
    <!-- Header -->
    <header class="header">
      <button
        class="menu-btn"
        onclick={() => { sidebarCollapsed = !sidebarCollapsed; }}
        aria-label="Toggle sidebar"
      >
        <svg width="20" height="20" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
        </svg>
      </button>
      <CatLogo size={28} tooltipPosition="bottom" />
      <h1 class="title">coahGPT</h1>
    </header>

    <!-- Messages -->
    <div class="messages">
      {#if messages.length === 0}
        <div class="empty-state">
          <PurpleCat size={100} glow={true} />
          <h2 class="empty-title">
            built for <CyclingText />
          </h2>
          <p class="empty-sub">ask anything, no account needed</p>
        </div>
      {:else}
        <div class="messages-inner">
          {#each messages as msg (msg.id)}
            <MessageBubble
              role={msg.role}
              content={msg.content}
              streaming={streaming && msg.id === messages.at(-1)?.id && msg.role === 'assistant'}
            />
          {/each}
          <div bind:this={messagesEnd}></div>
        </div>
      {/if}
    </div>

    <!-- Input -->
    <div class="input-area">
      <div class="input-wrapper">
        <textarea
          bind:this={textareaEl}
          bind:value={input}
          oninput={resizeTextarea}
          onkeydown={handleKeydown}
          placeholder="message coahGPT..."
          rows="1"
          disabled={streaming}
          class="input-field"
        ></textarea>
        <button
          onclick={sendMessage}
          disabled={streaming || !input.trim()}
          class="send-btn"
          aria-label="Send"
        >
          {#if streaming}
            <svg class="spin" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path stroke-linecap="round" stroke-width="2" d="M12 2v4m0 12v4m-7.07-3.93l2.83-2.83m8.48-8.48l2.83-2.83M2 12h4m12 0h4M4.93 4.93l2.83 2.83m8.48 8.48l2.83 2.83" />
            </svg>
          {:else}
            <svg width="18" height="18" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M12 5l7 7-7 7" />
            </svg>
          {/if}
        </button>
      </div>
      <p class="disclaimer">runs on ollama locally. no data leaves your machine.</p>
    </div>
  </main>
</div>

<style>
  .app {
    display: flex;
    height: 100vh;
    overflow: hidden;
  }

  .main {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
    background: var(--color-base);
  }

  .header {
    display: flex;
    align-items: center;
    gap: 0.625rem;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid color-mix(in srgb, var(--color-surface0) 40%, transparent);
    background: var(--color-base);
    flex-shrink: 0;
  }

  .menu-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0.375rem;
    border-radius: 0.375rem;
    color: var(--color-subtext0);
    background: none;
    border: none;
    cursor: pointer;
    transition: all 0.15s;
  }

  .menu-btn:hover {
    color: var(--color-text);
    background: color-mix(in srgb, var(--color-surface0) 40%, transparent);
  }

  .title {
    font-family: 'Playfair Display', serif;
    font-size: 1.125rem;
    font-weight: 800;
    color: var(--color-mauve);
    margin: 0;
  }

  .messages {
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
  }

  .messages-inner {
    max-width: 768px;
    margin: 0 auto;
    padding: 1.5rem 1rem;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    gap: 1.25rem;
    padding: 2rem;
    opacity: 0.9;
  }

  .empty-title {
    font-family: 'Playfair Display', serif;
    font-size: 1.5rem;
    font-weight: 800;
    color: var(--color-text);
    margin: 0;
    text-align: center;
  }

  .empty-sub {
    color: var(--color-overlay0);
    font-size: 0.875rem;
    margin: 0;
  }

  .input-area {
    padding: 0.75rem 1rem 1rem;
    background: var(--color-base);
    flex-shrink: 0;
  }

  .input-wrapper {
    max-width: 768px;
    margin: 0 auto;
    display: flex;
    align-items: flex-end;
    gap: 0.5rem;
    background: color-mix(in srgb, var(--color-surface0) 50%, transparent);
    border: 1px solid color-mix(in srgb, var(--color-surface1) 40%, transparent);
    border-radius: 1rem;
    padding: 0.5rem 0.5rem 0.5rem 1rem;
    transition: border-color 0.15s;
  }

  .input-wrapper:focus-within {
    border-color: color-mix(in srgb, var(--color-mauve) 40%, transparent);
  }

  .input-field {
    flex: 1;
    resize: none;
    background: none;
    border: none;
    outline: none;
    color: var(--color-text);
    font-family: var(--font-sans);
    font-size: 0.9375rem;
    line-height: 1.5;
    padding: 0.25rem 0;
    max-height: 200px;
  }

  .input-field::placeholder {
    color: var(--color-overlay0);
  }

  .input-field:disabled {
    opacity: 0.5;
  }

  .send-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 2.25rem;
    height: 2.25rem;
    border-radius: 0.625rem;
    background: var(--color-mauve);
    color: var(--color-crust);
    border: none;
    cursor: pointer;
    transition: all 0.15s;
    flex-shrink: 0;
  }

  .send-btn:hover:not(:disabled) {
    filter: brightness(1.1);
  }

  .send-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .spin {
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }

  .disclaimer {
    max-width: 768px;
    margin: 0.5rem auto 0;
    text-align: center;
    font-size: 0.7rem;
    color: var(--color-surface2);
  }

  @media (max-width: 640px) {
    .messages-inner {
      padding: 1rem 0.75rem;
    }

    .empty-title {
      font-size: 1.25rem;
    }
  }
</style>
