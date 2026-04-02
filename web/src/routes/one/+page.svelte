<script lang="ts">
  import CatLogo from '$lib/components/CatLogo.svelte';

  // placeholder video ID — Cole will replace with real one
  const videoId = 'SbubJqFYdwA';
</script>

<svelte:head>
  <link
    rel="stylesheet"
    href="https://fonts.googleapis.com/css2?family=Playfair+Display:wght@700;800;900&display=swap"
  />
  <title>CoahGPT One</title>
</svelte:head>

<div class="page">
  <nav class="nav">
    <div class="nav-inner">
      <div class="nav-logo" style="display: flex; align-items: center; gap: 0.5rem;"><CatLogo size={24} tooltipPosition="bottom" /> <a href="/">coahGPT</a></div>
      <div class="nav-right">
        <a href="/login" class="nav-link">Sign in</a>
        <a href="/chat" class="nav-cta">Try CoahGPT One</a>
      </div>
    </div>
  </nav>

  <main class="hero">
    <div class="badge">
      <CatLogo size={20} clickable={false} />
      <span>Introducing CoahGPT One</span>
    </div>

    <h1 class="heading">Our most efficient,<br/>capable model yet.</h1>

    <p class="subheading">Watch a video from one of our founders:</p>

    <div class="video-container">
      <iframe
        src="https://www.youtube.com/embed/{videoId}"
        title="CoahGPT One — Introduction"
        frameborder="0"
        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
        allowfullscreen
      ></iframe>
    </div>

    <div class="cta-row">
      <a href="/chat" class="cta-primary">Try CoahGPT One &rarr;</a>
      <a href="/" class="cta-secondary">Learn about coah code</a>
    </div>
  </main>

  <section class="stats">
    <div class="stat">
      <span class="stat-value">7B</span>
      <span class="stat-label">Parameters</span>
    </div>
    <div class="stat-divider"></div>
    <div class="stat">
      <span class="stat-value">12GB</span>
      <span class="stat-label">VRAM</span>
    </div>
    <div class="stat-divider"></div>
    <div class="stat">
      <span class="stat-value">1</span>
      <span class="stat-label">RTX 3060</span>
    </div>
    <div class="stat-divider"></div>
    <div class="stat">
      <span class="stat-value">$0</span>
      <span class="stat-label">Per token</span>
    </div>
  </section>

  <footer class="footer">
    <p>&copy; 2026 coahGPT Inc. (not actually incorporated)</p>
  </footer>
</div>

<style>
  .page {
    min-height: 100vh;
    background-color: var(--color-base);
    color: var(--color-text);
    display: flex;
    flex-direction: column;
  }

  .nav {
    position: sticky;
    top: 0;
    z-index: 50;
    padding: 0 1.5rem;
    height: 4rem;
    background: rgba(30, 30, 46, 0.8);
    backdrop-filter: blur(20px) saturate(1.8);
    border-bottom: 1px solid rgba(69, 71, 90, 0.3);
  }

  .nav-inner {
    max-width: 72rem;
    margin: 0 auto;
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 100%;
  }

  .nav-logo {
    font-size: 1.25rem;
    font-weight: 800;
  }

  .nav-logo a {
    color: var(--color-mauve);
    text-decoration: none;
  }

  .nav-right {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .nav-link {
    font-size: 0.875rem;
    color: var(--color-subtext0);
    text-decoration: none;
    transition: color 0.2s;
  }
  .nav-link:hover { color: var(--color-text); }

  .nav-cta {
    padding: 0.5rem 1rem;
    border-radius: 0.5rem;
    background: var(--color-mauve);
    color: var(--color-crust);
    font-size: 0.875rem;
    font-weight: 600;
    text-decoration: none;
    transition: opacity 0.2s;
  }
  .nav-cta:hover { opacity: 0.9; }

  .hero {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 5rem 2rem 3rem;
    text-align: center;
    max-width: 800px;
    margin: 0 auto;
    width: 100%;
  }

  .badge {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.4rem 1rem 0.4rem 0.5rem;
    border-radius: 999px;
    background: rgba(203, 166, 247, 0.1);
    border: 1px solid rgba(203, 166, 247, 0.2);
    font-size: 0.875rem;
    color: var(--color-mauve);
    margin-bottom: 2rem;
  }

  .heading {
    font-family: 'Playfair Display', Georgia, serif;
    font-size: clamp(2.5rem, 6vw, 4rem);
    font-weight: 800;
    line-height: 1.15;
    letter-spacing: -0.02em;
    color: var(--color-text);
    margin: 0 0 1.5rem;
  }

  .subheading {
    font-size: 1.125rem;
    color: var(--color-subtext0);
    margin: 0 0 2rem;
  }

  .video-container {
    width: 100%;
    max-width: 720px;
    aspect-ratio: 16 / 9;
    border-radius: 1rem;
    overflow: hidden;
    border: 1px solid var(--color-surface1);
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4), 0 0 0 1px rgba(203, 166, 247, 0.1);
    margin-bottom: 2.5rem;
  }

  .video-container iframe {
    width: 100%;
    height: 100%;
    border: none;
  }

  .cta-row {
    display: flex;
    align-items: center;
    gap: 1.5rem;
    flex-wrap: wrap;
    justify-content: center;
  }

  .cta-primary {
    padding: 0.875rem 2rem;
    border-radius: 0.75rem;
    background: var(--color-mauve);
    color: var(--color-crust);
    font-size: 1rem;
    font-weight: 600;
    text-decoration: none;
    transition: opacity 0.2s, box-shadow 0.2s;
    box-shadow: 0 4px 16px rgba(203, 166, 247, 0.3);
  }
  .cta-primary:hover {
    opacity: 0.9;
    box-shadow: 0 4px 24px rgba(203, 166, 247, 0.5);
  }

  .cta-secondary {
    font-size: 1rem;
    color: var(--color-subtext0);
    text-decoration: none;
    transition: color 0.2s;
  }
  .cta-secondary:hover { color: var(--color-mauve); }

  .stats {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 2.5rem;
    padding: 3rem 2rem;
    border-top: 1px solid rgba(69, 71, 90, 0.3);
  }

  .stat {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.25rem;
  }

  .stat-value {
    font-family: 'Playfair Display', Georgia, serif;
    font-size: 2rem;
    font-weight: 800;
    color: var(--color-mauve);
  }

  .stat-label {
    font-size: 0.8125rem;
    color: var(--color-subtext0);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .stat-divider {
    width: 1px;
    height: 3rem;
    background: var(--color-surface1);
  }

  .footer {
    padding: 2rem;
    text-align: center;
    border-top: 1px solid rgba(69, 71, 90, 0.3);
  }

  .footer p {
    margin: 0;
    font-size: 0.8125rem;
    color: var(--color-overlay0);
  }

  @media (max-width: 640px) {
    .heading {
      font-size: 2rem;
    }
    .stats {
      gap: 1.5rem;
    }
    .stat-value {
      font-size: 1.5rem;
    }
  }
</style>
