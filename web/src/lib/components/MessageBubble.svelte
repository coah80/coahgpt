<script lang="ts">
  import { marked } from 'marked';
  import hljs from 'highlight.js';
  import type { Tokens } from 'marked';
  import CatLogo from './CatLogo.svelte';

  let {
    role,
    content,
    streaming = false,
  }: {
    role: 'user' | 'assistant';
    content: string;
    streaming?: boolean;
  } = $props();

  const renderer = new marked.Renderer();

  renderer.code = function ({ text, lang }: Tokens.Code): string {
    const language = lang || '';
    let highlighted: string;
    if (language && hljs.getLanguage(language)) {
      highlighted = hljs.highlight(text, { language }).value;
    } else {
      highlighted = hljs.highlightAuto(text).value;
    }
    return `
      <div class="code-block-wrapper relative group rounded-lg overflow-hidden my-4">
        <div class="flex items-center justify-between px-4 py-2 bg-[#11111b] text-xs text-[#a6adc8]">
          <span>${language || 'code'}</span>
          <button
            onclick="navigator.clipboard.writeText(this.closest('.code-block-wrapper').querySelector('code').textContent);this.querySelector('.copy-label').textContent='Copied!';setTimeout(()=>this.querySelector('.copy-label').textContent='Copy',2000)"
            class="flex items-center gap-1.5 px-2 py-1 rounded text-[#a6adc8] hover:text-[#cdd6f4] hover:bg-[#313244] transition-colors cursor-pointer"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <rect x="9" y="9" width="13" height="13" rx="2" ry="2" stroke-width="2"></rect>
              <path stroke-width="2" d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"></path>
            </svg>
            <span class="copy-label">Copy</span>
          </button>
        </div>
        <pre class="bg-[#181825] px-4 py-3 overflow-x-auto text-sm leading-relaxed"><code class="font-mono">${highlighted}</code></pre>
      </div>`;
  };

  renderer.codespan = function ({ text }: Tokens.Codespan): string {
    return `<code class="bg-[#313244] text-[#cba6f7] px-1.5 py-0.5 rounded text-sm font-mono">${text}</code>`;
  };

  renderer.link = function ({ href, text }: Tokens.Link): string {
    return `<a href="${href}" target="_blank" rel="noopener noreferrer" class="text-[#b4befe] hover:text-[#cba6f7] underline underline-offset-2 transition-colors">${text}</a>`;
  };

  marked.setOptions({ renderer, breaks: true });

  const renderedContent = $derived(
    role === 'assistant' ? marked.parse(content) as string : content
  );
</script>

{#if role === 'user'}
  <div class="msg-row msg-row-user">
    <div class="user-bubble">
      <p class="whitespace-pre-wrap leading-relaxed">{content}</p>
    </div>
  </div>
{:else}
  <div class="msg-row msg-row-assistant">
    <div class="assistant-avatar">
      <CatLogo size={28} clickable={false} />
    </div>
    <div class="assistant-content">
      <div class="prose prose-invert prose-base max-w-none
        [&_h1]:text-text [&_h1]:text-lg [&_h1]:font-bold [&_h1]:mb-3 [&_h1]:mt-5
        [&_h2]:text-text [&_h2]:text-base [&_h2]:font-semibold [&_h2]:mb-2 [&_h2]:mt-4
        [&_h3]:text-text [&_h3]:text-sm [&_h3]:font-semibold [&_h3]:mb-2 [&_h3]:mt-3
        [&_p]:text-text [&_p]:leading-[1.6] [&_p]:mb-3 [&_p]:last:mb-0
        [&_ul]:mb-3 [&_ul]:pl-5 [&_li]:text-text [&_li]:mb-1.5 [&_li]:leading-[1.6]
        [&_ol]:mb-3 [&_ol]:pl-5 [&_ol_li]:mb-1.5
        [&_strong]:text-text [&_strong]:font-semibold
        [&_em]:text-subtext1
        [&_blockquote]:border-l-2 [&_blockquote]:border-mauve/40 [&_blockquote]:pl-4 [&_blockquote]:italic [&_blockquote]:text-subtext0 [&_blockquote]:my-3
      ">
        {@html renderedContent}
      </div>
      {#if streaming}
        <span class="streaming-cursor"></span>
      {/if}
    </div>
  </div>
{/if}

<style>
  .msg-row {
    display: flex;
    width: 100%;
    margin-bottom: 1.75rem;
  }

  .msg-row-user {
    justify-content: flex-end;
  }

  .msg-row-assistant {
    justify-content: flex-start;
    gap: 0.75rem;
    align-items: flex-start;
  }

  .user-bubble {
    max-width: 75%;
    padding: 0.75rem 1rem;
    border-radius: 1.25rem;
    border-bottom-right-radius: 0.375rem;
    background: color-mix(in srgb, var(--color-mauve) 12%, transparent);
    color: var(--color-text);
    font-size: 0.9375rem;
    line-height: 1.6;
  }

  .assistant-avatar {
    flex-shrink: 0;
    width: 28px;
    height: 28px;
    margin-top: 2px;
  }

  .assistant-content {
    flex: 1;
    min-width: 0;
    font-size: 1rem;
    line-height: 1.6;
    color: var(--color-text);
  }

  .streaming-cursor {
    display: inline-block;
    width: 2px;
    height: 1.1em;
    background: var(--color-mauve);
    margin-left: 2px;
    vertical-align: text-bottom;
    animation: blink 1s steps(2) infinite;
    border-radius: 1px;
  }

  @keyframes blink {
    0%, 100% { opacity: 1; }
    50% { opacity: 0; }
  }
</style>
