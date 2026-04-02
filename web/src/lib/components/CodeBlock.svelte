<script lang="ts">
  import hljs from 'highlight.js';

  let {
    code,
    language = '',
  }: {
    code: string;
    language?: string;
  } = $props();

  let copied = $state(false);

  const highlighted = $derived.by(() => {
    if (language && hljs.getLanguage(language)) {
      return hljs.highlight(code, { language }).value;
    }
    return hljs.highlightAuto(code).value;
  });

  function copyCode() {
    navigator.clipboard.writeText(code);
    copied = true;
    setTimeout(() => {
      copied = false;
    }, 2000);
  }
</script>

<div class="relative group rounded-lg overflow-hidden my-3">
  <div class="flex items-center justify-between px-4 py-2 bg-crust text-xs text-subtext0">
    <span>{language || 'code'}</span>
    <button
      onclick={copyCode}
      class="flex items-center gap-1 px-2 py-1 rounded text-subtext0 hover:text-text hover:bg-surface0 transition-colors cursor-pointer"
    >
      {#if copied}
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
        </svg>
        <span>Copied</span>
      {:else}
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <rect x="9" y="9" width="13" height="13" rx="2" ry="2" stroke-width="2" />
          <path stroke-width="2" d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" />
        </svg>
        <span>Copy</span>
      {/if}
    </button>
  </div>
  <pre class="bg-mantle px-4 py-3 overflow-x-auto text-sm leading-relaxed"><code class="font-mono">{@html highlighted}</code></pre>
</div>
