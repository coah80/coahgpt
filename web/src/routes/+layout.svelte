<script lang="ts">
  import '../app.css';

  let { children } = $props();
</script>

<svelte:head>
  <title>coahGPT</title>
  <meta name="description" content="Built for gamers, modders, and degens. Work with coah directly in your terminal." />
</svelte:head>

{@render children()}
