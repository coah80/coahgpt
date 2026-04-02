<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { goto } from '$app/navigation';
  import MessageBubble from '$lib/components/MessageBubble.svelte';
  import Sidebar from '$lib/components/Sidebar.svelte';
  import CatLogo from '$lib/components/CatLogo.svelte';
  import {
    loadConversations,
    saveConversations,
    loadLocalConversations,
    createConversation,
    createServerConversation,
    loadMessages,
    saveMessage,
    deleteServerConversation,
    addMessage,
    updateLastAssistantMessage,
    removeConversation,
    upsertConversation,
  } from '$lib/chat-store.js';
  import { getMe, getUser, logout, getToken, isLoggedIn } from '$lib/auth.js';
  import type { Conversation, Message } from '$lib/types.js';

  let conversations = $state<ReadonlyArray<Conversation>>([]);
  let activeId = $state<string | null>(null);
  let inputText = $state('');
  let streaming = $state(false);
  let sidebarCollapsed = $state(true);
  let deepResearch = $state(false);
  let webSearch = $state(false);
  let messagesContainer: HTMLElement | undefined = $state();
  let textareaEl: HTMLTextAreaElement | undefined = $state();
  let abortController: AbortController | null = $state(null);
  let authChecked = $state(false);

  const user = $derived(getUser());

  const activeConversation = $derived(
    conversations.find((c) => c.id === activeId) ?? null
  );

  const messages = $derived(activeConversation?.messages ?? []);
  const hasMessages = $derived(messages.length > 0);

  const API_BASE = import.meta.env.DEV ? 'http://localhost:8095' : '';

  const suggestions = [
    { label: 'write some code', icon: 'code' },
    { label: 'explain a concept', icon: 'learn' },
    { label: 'debug my shit', icon: 'debug' },
    { label: 'roast my code', icon: 'roast' },
  ] as const;

  onMount(async () => {
    const me = await getMe();
    if (!me) {
      goto('/login');
      return;
    }
    authChecked = true;

    if (isLoggedIn()) {
      conversations = await loadConversations();
    } else {
      conversations = loadLocalConversations();
    }
    if (conversations.length > 0) {
      activeId = conversations[0].id;
      // load messages for the active conversation from the server
      if (isLoggedIn() && activeId) {
        const msgs = await loadMessages(activeId);
        if (msgs.length > 0) {
          const active = conversations.find((c) => c.id === activeId);
          if (active) {
            const updated = { ...active, messages: msgs };
            conversations = upsertConversation(conversations, updated);
          }
        }
      }
    }

    sidebarCollapsed = true;

    tick().then(() => {
      textareaEl?.focus();
    });
  });

  async function handleSignOut() {
    await logout();
    goto('/');
  }

  function scrollToBottom() {
    tick().then(() => {
      if (messagesContainer) {
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
      }
    });
  }

  function newChat() {
    const id = crypto.randomUUID();
    const conv = createConversation(id);
    conversations = [conv, ...conversations];
    activeId = id;
    if (isLoggedIn()) {
      createServerConversation(id);
    } else {
      saveConversations(conversations);
    }
    if (window.innerWidth < 768) sidebarCollapsed = true;
    tick().then(() => textareaEl?.focus());
  }

  async function selectChat(id: string) {
    activeId = id;
    if (window.innerWidth < 768) sidebarCollapsed = true;

    // load messages from server if authed and conversation has no messages loaded yet
    const active = conversations.find((c) => c.id === id);
    if (isLoggedIn() && active && active.messages.length === 0) {
      const msgs = await loadMessages(id);
      if (msgs.length > 0) {
        const updated = { ...active, messages: msgs };
        conversations = upsertConversation(conversations, updated);
      }
    }

    scrollToBottom();
  }

  function deleteChat(id: string) {
    conversations = removeConversation(conversations, id);
    if (activeId === id) {
      activeId = conversations.length > 0 ? conversations[0].id : null;
    }
    if (isLoggedIn()) {
      deleteServerConversation(id);
    } else {
      saveConversations(conversations);
    }
  }

  function autoResize() {
    if (!textareaEl) return;
    textareaEl.style.height = 'auto';
    const maxHeight = 6 * 24;
    textareaEl.style.height = Math.min(textareaEl.scrollHeight, maxHeight) + 'px';
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  }

  function stopStreaming() {
    if (abortController) {
      abortController.abort();
      abortController = null;
    }
  }

  function populateInput(text: string) {
    inputText = text;
    tick().then(() => {
      textareaEl?.focus();
      autoResize();
    });
  }

  function getGreetingName(): string {
    if (!user) return 'there';
    const firstName = user.name.split(' ')[0].toLowerCase();
    return firstName || 'there';
  }

  async function sendMessage() {
    const text = inputText.trim();
    if (!text || streaming) return;

    let convId = activeId;
    if (!convId) {
      const id = crypto.randomUUID();
      const conv = createConversation(id);
      conversations = [conv, ...conversations];
      convId = id;
      activeId = id;
    }

    const userMessage: Message = {
      id: crypto.randomUUID(),
      role: 'user',
      content: text,
      timestamp: Date.now(),
    };

    let conv = conversations.find((c) => c.id === convId)!;
    conv = addMessage(conv, userMessage);
    conversations = upsertConversation(conversations, conv);
    if (!isLoggedIn()) {
      saveConversations(conversations);
    }

    inputText = '';
    if (textareaEl) {
      textareaEl.style.height = 'auto';
    }
    scrollToBottom();

    const assistantMessage: Message = {
      id: crypto.randomUUID(),
      role: 'assistant',
      content: '',
      timestamp: Date.now(),
    };

    conv = addMessage(conv, assistantMessage);
    conversations = upsertConversation(conversations, conv);
    streaming = true;
    scrollToBottom();

    const controller = new AbortController();
    abortController = controller;

    try {
      const modeLabel = deepResearch ? ' [Deep Research]' : webSearch ? ' [Web Search]' : '';
      const authed = isLoggedIn();
      const response = await fetch(`${API_BASE}/api/chat`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...(getToken() ? { Authorization: `Bearer ${getToken()}` } : {}),
        },
        body: JSON.stringify({
          session_id: convId,
          message: text + modeLabel,
          ...(authed ? { conversation_id: convId } : {}),
        }),
        signal: controller.signal,
      });

      if (!response.ok || !response.body) {
        throw new Error(`API error: ${response.status}`);
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';
      let fullContent = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() ?? '';

        for (const line of lines) {
          if (!line.startsWith('data: ')) continue;
          const jsonStr = line.slice(6).trim();
          if (!jsonStr) continue;

          try {
            const parsed = JSON.parse(jsonStr);
            if (parsed.token) {
              fullContent += parsed.token;
              conv = conversations.find((c) => c.id === convId)!;
              conv = updateLastAssistantMessage(conv, fullContent);
              conversations = upsertConversation(conversations, conv);
              scrollToBottom();
            }
            if (parsed.done) {
              break;
            }
          } catch {
            // skip malformed JSON
          }
        }
      }
    } catch (err: unknown) {
      if (err instanceof DOMException && err.name === 'AbortError') {
        // user cancelled
      } else {
        conv = conversations.find((c) => c.id === convId)!;
        conv = updateLastAssistantMessage(
          conv,
          "lol the server isn't running right now. try again later or yell at coah on discord."
        );
        conversations = upsertConversation(conversations, conv);
      }
    } finally {
      streaming = false;
      abortController = null;
      if (!isLoggedIn()) {
        saveConversations(conversations);
      }
      scrollToBottom();
    }
  }
</script>

<svelte:head>
  <title>Chat - coahGPT</title>
</svelte:head>

{#if !authChecked}
  <div class="min-h-screen bg-base flex items-center justify-center">
    <div class="w-6 h-6 border-2 border-mauve/30 border-t-mauve rounded-full animate-spin"></div>
  </div>
{:else}
<div class="chat-layout">
  <Sidebar
    {conversations}
    {activeId}
    collapsed={sidebarCollapsed}
    onNewChat={newChat}
    onSelectChat={selectChat}
    onDeleteChat={deleteChat}
    onToggleCollapse={() => sidebarCollapsed = !sidebarCollapsed}
  />

  <div class="main-area">
    <!-- Top bar -->
    <header class="top-bar">
      <div class="top-bar-left">
        <button
          onclick={() => sidebarCollapsed = !sidebarCollapsed}
          class="icon-btn"
          aria-label="Toggle sidebar"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
      </div>

      <div class="top-bar-center">
        {#if hasMessages && activeConversation}
          <span class="top-bar-title">{activeConversation.preview}</span>
        {:else}
          <span class="top-bar-brand"><CatLogo size={22} clickable={false} /> coahGPT</span>
        {/if}
        {#if deepResearch}
          <span class="mode-badge mode-badge-research">Deep Research</span>
        {/if}
        {#if webSearch}
          <span class="mode-badge mode-badge-web">Web Search</span>
        {/if}
      </div>

      <div class="top-bar-right">
        {#if user}
          <div class="user-badge">
            <div class="user-avatar-sm">{user.name.charAt(0).toUpperCase()}</div>
            <span class="user-name-sm">{user.name}</span>
          </div>
        {/if}
        <button onclick={handleSignOut} class="sign-out-btn">
          Sign out
        </button>
      </div>
    </header>

    <!-- Content area -->
    {#if !hasMessages}
      <div class="empty-state">
        <div class="empty-state-content">
          <CatLogo size={48} />

          <div class="greeting">
            <p class="greeting-hey">hey, {getGreetingName()}</p>
            <h1 class="greeting-main">What do you wanna build?</h1>
          </div>

          <div class="input-container-centered">
            <div class="input-box">
              <textarea
                bind:this={textareaEl}
                bind:value={inputText}
                oninput={autoResize}
                onkeydown={handleKeydown}
                placeholder="ask coahGPT anything..."
                disabled={streaming}
                rows="1"
                class="input-textarea"
              ></textarea>
              <div class="input-actions">
                <span class="model-badge">CoahGPT One</span>
                <button
                  onclick={sendMessage}
                  disabled={streaming || !inputText.trim()}
                  class="send-btn {streaming || !inputText.trim() ? 'send-btn-disabled' : 'send-btn-active'}"
                  aria-label="Send message"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 19V5m0 0l-6 6m6-6l6 6" />
                  </svg>
                </button>
              </div>
            </div>

            <div class="toggles-row">
              <button
                onclick={() => { deepResearch = !deepResearch; if (deepResearch) webSearch = false; }}
                class="feature-toggle {deepResearch ? 'feature-toggle-active-research' : ''}"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
                </svg>
                Deep Research
              </button>
              <button
                onclick={() => { webSearch = !webSearch; if (webSearch) deepResearch = false; }}
                class="feature-toggle {webSearch ? 'feature-toggle-active-web' : ''}"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
                </svg>
                Web Search
              </button>
            </div>

            <div class="suggestion-pills">
              {#each suggestions as s}
                <button
                  onclick={() => populateInput(s.label)}
                  class="suggestion-pill"
                >
                  {#if s.icon === 'code'}
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                    </svg>
                  {:else if s.icon === 'learn'}
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
                    </svg>
                  {:else if s.icon === 'debug'}
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  {:else}
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 18.657A8 8 0 016.343 7.343S7 9 9 10c0-2 .5-5 2.986-7C14 5 16.09 5.777 17.656 7.343A7.975 7.975 0 0120 13a7.975 7.975 0 01-2.343 5.657z" />
                    </svg>
                  {/if}
                  {s.label}
                </button>
              {/each}
            </div>
          </div>

          <p class="disclaimer">coahGPT can make mistakes. please don't sue us.</p>
        </div>
      </div>
    {:else}
      <div class="conversation-area">
        <div bind:this={messagesContainer} class="messages-scroll">
          <div class="messages-inner">
            {#each messages as msg, i (msg.id)}
              <MessageBubble
                role={msg.role}
                content={msg.content}
                streaming={streaming && i === messages.length - 1 && msg.role === 'assistant'}
              />
            {/each}
          </div>
        </div>

        <div class="bottom-input-area">
          <div class="bottom-input-inner">
            {#if streaming}
              <div class="stop-container">
                <button onclick={stopStreaming} class="stop-btn">
                  <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
                    <rect x="6" y="6" width="12" height="12" rx="2" />
                  </svg>
                  Stop generating
                </button>
              </div>
            {/if}

            <div class="toggles-row">
              <button
                onclick={() => { deepResearch = !deepResearch; if (deepResearch) webSearch = false; }}
                class="feature-toggle {deepResearch ? 'feature-toggle-active-research' : ''}"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
                </svg>
                Deep Research
              </button>
              <button
                onclick={() => { webSearch = !webSearch; if (webSearch) deepResearch = false; }}
                class="feature-toggle {webSearch ? 'feature-toggle-active-web' : ''}"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
                </svg>
                Web Search
              </button>
            </div>

            <div class="input-box">
              <textarea
                bind:this={textareaEl}
                bind:value={inputText}
                oninput={autoResize}
                onkeydown={handleKeydown}
                placeholder={streaming ? 'thinking...' : 'ask coahGPT anything...'}
                disabled={streaming}
                rows="1"
                class="input-textarea"
              ></textarea>
              <div class="input-actions">
                <span class="model-badge">CoahGPT One</span>
                <button
                  onclick={sendMessage}
                  disabled={streaming || !inputText.trim()}
                  class="send-btn {streaming || !inputText.trim() ? 'send-btn-disabled' : 'send-btn-active'}"
                  aria-label="Send message"
                >
                  {#if streaming}
                    <div class="spinner"></div>
                  {:else}
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M12 19V5m0 0l-6 6m6-6l6 6" />
                    </svg>
                  {/if}
                </button>
              </div>
            </div>

            <p class="disclaimer-bottom">coahGPT can make mistakes. please don't sue us.</p>
          </div>
        </div>
      </div>
    {/if}
  </div>
</div>
{/if}

<style>
  /* ---- Layout ---- */
  .chat-layout {
    height: 100vh;
    display: flex;
    background-color: var(--color-base);
    overflow: hidden;
  }

  .main-area {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  /* ---- Top bar ---- */
  .top-bar {
    display: flex;
    align-items: center;
    height: 3.25rem;
    padding: 0 1rem;
    border-bottom: 1px solid color-mix(in srgb, var(--color-surface0) 30%, transparent);
    background: var(--color-base);
    flex-shrink: 0;
  }

  .top-bar-left {
    flex-shrink: 0;
  }

  .top-bar-center {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    min-width: 0;
  }

  .top-bar-brand {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
    font-weight: 800;
    color: var(--color-mauve);
    letter-spacing: -0.01em;
  }

  .top-bar-title {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--color-text);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 360px;
  }

  .top-bar-right {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .icon-btn {
    padding: 0.5rem;
    border-radius: 0.5rem;
    color: var(--color-subtext0);
    background: none;
    border: none;
    cursor: pointer;
    transition: all 0.15s;
  }

  .icon-btn:hover {
    color: var(--color-text);
    background: color-mix(in srgb, var(--color-surface0) 50%, transparent);
  }

  .mode-badge {
    padding: 0.125rem 0.5rem;
    border-radius: 9999px;
    font-size: 0.6875rem;
    font-weight: 600;
    letter-spacing: 0.01em;
  }

  .mode-badge-research {
    background: color-mix(in srgb, var(--color-lavender) 12%, transparent);
    color: var(--color-lavender);
  }

  .mode-badge-web {
    background: color-mix(in srgb, var(--color-green) 12%, transparent);
    color: var(--color-green);
  }

  .user-badge {
    display: none;
    align-items: center;
    gap: 0.5rem;
  }

  @media (min-width: 640px) {
    .user-badge {
      display: flex;
    }
  }

  .user-avatar-sm {
    width: 1.75rem;
    height: 1.75rem;
    border-radius: 9999px;
    background: var(--color-mauve);
    color: var(--color-crust);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.75rem;
    font-weight: 700;
  }

  .user-name-sm {
    font-size: 0.75rem;
    color: var(--color-subtext0);
    max-width: 120px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .sign-out-btn {
    padding: 0.25rem 0.625rem;
    border-radius: 0.5rem;
    font-size: 0.75rem;
    color: var(--color-subtext0);
    background: none;
    border: none;
    cursor: pointer;
    transition: all 0.15s;
  }

  .sign-out-btn:hover {
    color: var(--color-red);
    background: color-mix(in srgb, var(--color-red) 10%, transparent);
  }

  /* ---- Empty state ---- */
  .empty-state {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 2rem;
    overflow-y: auto;
  }

  .empty-state-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1.75rem;
    width: 100%;
    max-width: 640px;
  }

  .greeting {
    text-align: center;
  }

  .greeting-hey {
    font-size: 1rem;
    color: var(--color-subtext0);
    margin: 0 0 0.375rem 0;
  }

  .greeting-main {
    font-size: 1.75rem;
    font-weight: 700;
    color: var(--color-text);
    margin: 0;
    letter-spacing: -0.02em;
  }

  @media (min-width: 640px) {
    .greeting-main {
      font-size: 2.125rem;
    }
  }

  /* ---- Input box (shared) ---- */
  .input-container-centered {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .input-box {
    display: flex;
    align-items: flex-end;
    gap: 0.5rem;
    background: var(--color-surface0);
    border: 1px solid var(--color-surface1);
    border-radius: 1.5rem;
    padding: 0.625rem 0.625rem 0.625rem 1.125rem;
    transition: border-color 0.2s, box-shadow 0.2s;
  }

  .input-box:focus-within {
    border-color: color-mix(in srgb, var(--color-mauve) 40%, transparent);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-mauve) 8%, transparent);
  }

  .input-textarea {
    flex: 1;
    background: transparent;
    color: var(--color-text);
    border: none;
    outline: none;
    resize: none;
    font-size: 1rem;
    line-height: 1.5rem;
    padding: 0.375rem 0;
    max-height: 9rem;
    font-family: inherit;
  }

  .input-textarea::placeholder {
    color: var(--color-overlay0);
  }

  .input-textarea:disabled {
    opacity: 0.5;
  }

  .input-actions {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-shrink: 0;
  }

  .model-badge {
    display: none;
    padding: 0.25rem 0.625rem;
    border-radius: 9999px;
    font-size: 0.6875rem;
    font-weight: 600;
    background: color-mix(in srgb, var(--color-mauve) 12%, transparent);
    color: var(--color-mauve);
    white-space: nowrap;
    letter-spacing: 0.01em;
  }

  @media (min-width: 480px) {
    .model-badge {
      display: inline-flex;
    }
  }

  .send-btn {
    width: 2.25rem;
    height: 2.25rem;
    border-radius: 9999px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    cursor: pointer;
    transition: all 0.15s;
    flex-shrink: 0;
  }

  .send-btn-active {
    background: var(--color-mauve);
    color: var(--color-crust);
  }

  .send-btn-active:hover {
    opacity: 0.85;
    transform: scale(1.05);
  }

  .send-btn-disabled {
    background: color-mix(in srgb, var(--color-surface1) 40%, transparent);
    color: var(--color-overlay0);
    cursor: not-allowed;
  }

  .spinner {
    width: 1rem;
    height: 1rem;
    border: 2px solid color-mix(in srgb, var(--color-overlay0) 30%, transparent);
    border-top-color: var(--color-overlay0);
    border-radius: 9999px;
    animation: spin 0.6s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  /* ---- Feature toggles ---- */
  .toggles-row {
    display: flex;
    gap: 0.5rem;
  }

  .feature-toggle {
    display: flex;
    align-items: center;
    gap: 0.375rem;
    padding: 0.375rem 0.75rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 500;
    background: color-mix(in srgb, var(--color-surface0) 30%, transparent);
    color: var(--color-subtext0);
    border: 1px solid color-mix(in srgb, var(--color-surface1) 20%, transparent);
    cursor: pointer;
    transition: all 0.15s;
  }

  .feature-toggle:hover {
    color: var(--color-subtext1);
    background: color-mix(in srgb, var(--color-surface0) 50%, transparent);
  }

  .feature-toggle-active-research {
    background: color-mix(in srgb, var(--color-lavender) 20%, transparent);
    color: var(--color-lavender);
    border-color: color-mix(in srgb, var(--color-lavender) 30%, transparent);
  }

  .feature-toggle-active-web {
    background: color-mix(in srgb, var(--color-green) 20%, transparent);
    color: var(--color-green);
    border-color: color-mix(in srgb, var(--color-green) 30%, transparent);
  }

  /* ---- Suggestion pills ---- */
  .suggestion-pills {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    justify-content: center;
  }

  .suggestion-pill {
    display: flex;
    align-items: center;
    gap: 0.375rem;
    padding: 0.5rem 0.875rem;
    border-radius: 9999px;
    font-size: 0.8125rem;
    color: var(--color-subtext0);
    background: color-mix(in srgb, var(--color-surface0) 40%, transparent);
    border: 1px solid color-mix(in srgb, var(--color-surface1) 25%, transparent);
    cursor: pointer;
    transition: all 0.15s;
  }

  .suggestion-pill:hover {
    color: var(--color-text);
    background: color-mix(in srgb, var(--color-surface0) 70%, transparent);
    border-color: color-mix(in srgb, var(--color-surface1) 50%, transparent);
  }

  .disclaimer {
    font-size: 0.75rem;
    color: var(--color-overlay0);
    text-align: center;
    margin: 0;
  }

  /* ---- Conversation area ---- */
  .conversation-area {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-height: 0;
  }

  .messages-scroll {
    flex: 1;
    overflow-y: auto;
    padding: 2rem 1rem;
  }

  .messages-inner {
    max-width: 42.5rem;
    margin: 0 auto;
  }

  /* ---- Bottom input ---- */
  .bottom-input-area {
    flex-shrink: 0;
    background: linear-gradient(to bottom, transparent, var(--color-base) 12px);
    padding: 1rem 1rem 1.25rem;
  }

  .bottom-input-inner {
    max-width: 42.5rem;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .stop-container {
    display: flex;
    justify-content: center;
    margin-bottom: 0.25rem;
  }

  .stop-btn {
    display: flex;
    align-items: center;
    gap: 0.375rem;
    padding: 0.375rem 0.875rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--color-subtext0);
    background: var(--color-surface0);
    border: 1px solid var(--color-surface1);
    cursor: pointer;
    transition: all 0.15s;
  }

  .stop-btn:hover {
    color: var(--color-text);
    background: color-mix(in srgb, var(--color-surface0) 80%, var(--color-surface1));
  }

  .disclaimer-bottom {
    font-size: 0.6875rem;
    color: var(--color-overlay0);
    text-align: center;
    margin: 0;
  }
</style>
