<script lang="ts">
  import type { Conversation } from '$lib/types.js';
  import CatLogo from './CatLogo.svelte';

  let {
    conversations,
    activeId,
    collapsed = true,
    onNewChat,
    onSelectChat,
    onDeleteChat,
    onToggleCollapse,
  }: {
    conversations: ReadonlyArray<Conversation>;
    activeId: string | null;
    collapsed?: boolean;
    onNewChat: () => void;
    onSelectChat: (id: string) => void;
    onDeleteChat: (id: string) => void;
    onToggleCollapse: () => void;
  } = $props();

  let hovered = $state(false);
  const expanded = $derived(!collapsed || hovered);

  function formatTime(ts: number): string {
    const d = new Date(ts);
    const now = new Date();
    const diffMs = now.getTime() - d.getTime();
    const diffHours = diffMs / (1000 * 60 * 60);

    if (diffHours < 1) return 'just now';
    if (diffHours < 24) return `${Math.floor(diffHours)}h ago`;
    if (diffHours < 48) return 'yesterday';
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  }

  function truncate(str: string, len: number): string {
    return str.length > len ? str.slice(0, len) + '...' : str;
  }
</script>

<!-- Mobile overlay -->
{#if !collapsed}
  <button
    class="mobile-overlay"
    onclick={onToggleCollapse}
    aria-label="Close sidebar"
  ></button>
{/if}

<aside
  class="sidebar"
  class:sidebar-expanded={expanded}
  class:sidebar-collapsed={!expanded}
  class:sidebar-mobile-open={!collapsed}
  class:sidebar-mobile-closed={collapsed}
  onmouseenter={() => { hovered = true; }}
  onmouseleave={() => { hovered = false; }}
>
  <!-- Top section: logo + new chat -->
  <div class="sidebar-top">
    <button
      onclick={onNewChat}
      class="new-chat-btn"
      aria-label="New Chat"
    >
      <svg class="icon-sm" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
      </svg>
      {#if expanded}
        <span class="btn-label">New Chat</span>
      {/if}
    </button>

    {#if !collapsed}
      <button
        onclick={onToggleCollapse}
        class="close-btn"
        aria-label="Close sidebar"
      >
        <svg class="icon-sm" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    {/if}
  </div>

  <!-- Chat list -->
  <nav class="chat-list">
    {#each conversations as conv (conv.id)}
      <div
        role="button"
        tabindex="0"
        onclick={() => onSelectChat(conv.id)}
        onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') onSelectChat(conv.id); }}
        class="chat-item"
        class:chat-item-active={conv.id === activeId}
      >
        <svg class="icon-chat" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
        </svg>
        {#if expanded}
          <div class="chat-item-text">
            <p class="chat-item-title">{truncate(conv.preview, 28)}</p>
            <p class="chat-item-time" class:chat-item-time-active={conv.id === activeId}>
              {formatTime(conv.updatedAt)}
            </p>
          </div>
          <button
            onclick={(e) => { e.stopPropagation(); onDeleteChat(conv.id); }}
            class="delete-btn"
            aria-label="Delete conversation"
          >
            <svg class="icon-xs" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </button>
        {/if}
      </div>
    {/each}
    {#if conversations.length === 0 && expanded}
      <p class="empty-label">No conversations yet</p>
    {/if}
  </nav>

  <!-- Bottom: back to home -->
  <div class="sidebar-bottom">
    <a href="/" class="home-link">
      <svg class="icon-sm" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-4 0a1 1 0 01-1-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 01-1 1" />
      </svg>
      {#if expanded}
        <span class="link-label">Back to home</span>
      {/if}
    </a>
  </div>
</aside>

<style>
  .mobile-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    z-index: 30;
    border: none;
    cursor: pointer;
  }

  @media (min-width: 768px) {
    .mobile-overlay {
      display: none;
    }
  }

  .sidebar {
    position: fixed;
    z-index: 40;
    height: 100%;
    display: flex;
    flex-direction: column;
    background: var(--color-mantle);
    border-right: 1px solid color-mix(in srgb, var(--color-surface0) 40%, transparent);
    transition: width 0.2s ease;
    overflow: hidden;
  }

  @media (min-width: 768px) {
    .sidebar {
      position: relative;
    }
  }

  .sidebar-expanded {
    width: 260px;
  }

  .sidebar-collapsed {
    width: 52px;
  }

  /* Mobile: off-screen when collapsed */
  @media (max-width: 767px) {
    .sidebar-mobile-closed {
      transform: translateX(-100%);
    }

    .sidebar-mobile-open {
      transform: translateX(0);
      width: 280px;
    }
  }

  .sidebar-top {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem;
    border-bottom: 1px solid color-mix(in srgb, var(--color-surface0) 40%, transparent);
  }

  .new-chat-btn {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding: 0.5rem;
    border-radius: 0.5rem;
    background: color-mix(in srgb, var(--color-mauve) 10%, transparent);
    color: var(--color-mauve);
    border: none;
    cursor: pointer;
    font-size: 0.8125rem;
    font-weight: 600;
    transition: background 0.15s;
    white-space: nowrap;
    overflow: hidden;
  }

  .new-chat-btn:hover {
    background: color-mix(in srgb, var(--color-mauve) 18%, transparent);
  }

  .close-btn {
    padding: 0.5rem;
    border-radius: 0.5rem;
    color: var(--color-subtext0);
    background: none;
    border: none;
    cursor: pointer;
    transition: all 0.15s;
    flex-shrink: 0;
  }

  @media (min-width: 768px) {
    .close-btn {
      display: none;
    }
  }

  .close-btn:hover {
    color: var(--color-text);
    background: color-mix(in srgb, var(--color-surface0) 50%, transparent);
  }

  .btn-label,
  .link-label {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .chat-list {
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
    padding: 0.375rem;
  }

  .chat-item {
    display: flex;
    align-items: center;
    gap: 0.625rem;
    padding: 0.5rem;
    border-radius: 0.5rem;
    cursor: pointer;
    transition: background 0.15s, color 0.15s;
    color: var(--color-subtext0);
    min-height: 2.25rem;
  }

  .chat-item:hover {
    background: color-mix(in srgb, var(--color-surface0) 40%, transparent);
    color: var(--color-subtext1);
  }

  .chat-item-active {
    background: color-mix(in srgb, var(--color-surface0) 60%, transparent);
    color: var(--color-text);
  }

  .icon-chat {
    width: 1.125rem;
    height: 1.125rem;
    flex-shrink: 0;
  }

  .chat-item-text {
    flex: 1;
    min-width: 0;
    overflow: hidden;
  }

  .chat-item-title {
    font-size: 0.8125rem;
    font-weight: 500;
    margin: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .chat-item-time {
    font-size: 0.6875rem;
    margin: 0.125rem 0 0 0;
    color: var(--color-overlay0);
  }

  .chat-item-time-active {
    color: var(--color-subtext0);
  }

  .delete-btn {
    opacity: 0;
    padding: 0.25rem;
    border-radius: 0.25rem;
    color: var(--color-overlay0);
    background: none;
    border: none;
    cursor: pointer;
    transition: all 0.15s;
    flex-shrink: 0;
  }

  .chat-item:hover .delete-btn {
    opacity: 1;
  }

  .delete-btn:hover {
    color: var(--color-red);
  }

  .icon-sm {
    width: 1.125rem;
    height: 1.125rem;
    flex-shrink: 0;
  }

  .icon-xs {
    width: 0.875rem;
    height: 0.875rem;
  }

  .empty-label {
    text-align: center;
    font-size: 0.8125rem;
    color: var(--color-overlay0);
    padding: 2rem 0.5rem;
    margin: 0;
  }

  .sidebar-bottom {
    padding: 0.75rem;
    border-top: 1px solid color-mix(in srgb, var(--color-surface0) 40%, transparent);
  }

  .home-link {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem;
    border-radius: 0.5rem;
    font-size: 0.8125rem;
    color: var(--color-subtext0);
    text-decoration: none;
    transition: all 0.15s;
    white-space: nowrap;
    overflow: hidden;
  }

  .home-link:hover {
    color: var(--color-text);
    background: color-mix(in srgb, var(--color-surface0) 40%, transparent);
  }
</style>
