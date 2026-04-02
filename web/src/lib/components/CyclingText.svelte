<script lang="ts">
  import { onMount } from 'svelte';

  const words: ReadonlyArray<{ text: string; color: string }> = [
    { text: 'gamers', color: 'var(--color-green)' },
    { text: 'modders', color: 'var(--color-mauve)' },
    { text: 'degens', color: 'var(--color-pink)' },
    { text: 'hackers', color: 'var(--color-lavender)' },
    { text: 'shitposters', color: 'var(--color-peach)' },
    { text: 'homebrew', color: 'var(--color-yellow)' },
    { text: 'OpenAI', color: 'var(--color-green)' },
    { text: 'Anthropic', color: 'var(--color-lavender)' },
  ];

  let currentIndex = $state(0);
  let visible = $state(true);

  onMount(() => {
    const interval = setInterval(() => {
      visible = false;
      setTimeout(() => {
        currentIndex = (currentIndex + 1) % words.length;
        visible = true;
      }, 300);
    }, 2000);

    return () => clearInterval(interval);
  });
</script>

<span class="inline-block relative">
  <span
    class="inline-block transition-all duration-300 ease-out"
    class:opacity-0={!visible}
    class:opacity-100={visible}
    class:translate-y-2={!visible}
    class:translate-y-0={visible}
    style:color={words[currentIndex].color}
  >
    {words[currentIndex].text}
  </span>
</span>
