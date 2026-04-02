<script lang="ts">
  import CyclingText from '$lib/components/CyclingText.svelte';
  import CatLogo from '$lib/components/CatLogo.svelte';

  let copied = $state(false);
  const installCmd = 'curl -fsSL https://coahgpt.com/install.sh | bash';

  function copyInstall() {
    navigator.clipboard.writeText(installCmd);
    copied = true;
    setTimeout(() => {
      copied = false;
    }, 2000);
  }
</script>

<svelte:head>
  <link
    rel="stylesheet"
    href="https://fonts.googleapis.com/css2?family=Playfair+Display:wght@700;800;900&display=swap"
  />
</svelte:head>

<div class="landing">
  <!-- ═══════════════ Nav ═══════════════ -->
  <nav class="nav">
    <div class="nav-inner">
      <div class="nav-logo" style="display: flex; align-items: center; gap: 0.5rem;"><CatLogo size={28} tooltipPosition="bottom" /> <a href="/">coahGPT</a></div>
      <div class="nav-links">
        <a
          href="https://github.com/coah80/coahgpt"
          target="_blank"
          rel="noopener noreferrer"
          class="nav-link"
        >
          GitHub
        </a>
        <a href="/login" class="nav-link">Sign in</a>
        <a href="/login" class="nav-cta">Try it</a>
      </div>
    </div>
  </nav>

  <!-- ═══════════════ Hero ═══════════════ -->
  <section class="hero">
    <div class="hero-content">
      <!-- Pill badge -->
      <div class="badge">
        <span class="badge-dot"></span>
        Powered by vibes and an RTX 3060
      </div>

      <!-- Main heading -->
      <h1 class="hero-heading">
        <span class="hero-heading-prefix">Built for</span>
        <CyclingText />
      </h1>

      <!-- Subtitle -->
      <p class="hero-subtitle">
        Work with coah directly in your codebase. Build, debug, and ship from your terminal.
      </p>

      <!-- Install command box -->
      <div class="install-box">
        <div class="install-label">
          Get coah code
          <svg class="install-chevron" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2.5"
              d="M19 9l-7 7-7-7"
            />
          </svg>
        </div>
        <div class="install-divider"></div>
        <code class="install-cmd">
          <span class="token-cmd">curl</span>
          <span class="token-flag"> -fsSL</span>
          <span class="token-url"> https://coahgpt.com/install.sh</span>
          <span class="token-pipe"> |</span>
          <span class="token-cmd"> bash</span>
        </code>
        <div class="install-divider"></div>
        <button onclick={copyInstall} class="install-copy" aria-label="Copy install command">
          {#if copied}
            <svg class="copy-icon copy-check" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M5 13l4 4L19 7"
              />
            </svg>
          {:else}
            <svg class="copy-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <rect x="9" y="9" width="13" height="13" rx="2" ry="2" stroke-width="2" />
              <path stroke-width="2" d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" />
            </svg>
          {/if}
        </button>
      </div>

      <!-- CTA button -->
      <a href="/login" class="hero-cta">
        Try it now in the browser
        <svg class="cta-arrow" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M17 8l4 4m0 0l-4 4m4-4H3"
          />
        </svg>
      </a>

      <!-- coah code callout -->
      <div class="callout">
        <span class="callout-badge">NEW!!</span>
        <span class="callout-name">coah code</span>
        <span class="callout-tagline">claude code but terrible.</span>
      </div>
    </div>
  </section>

  <!-- ═══════════════ Cat Section ═══════════════ -->
  <section class="cat-section">
    <div class="cat-float">
      <CatLogo size={80} />
    </div>
    <p class="cat-caption">meow (translation: i am sentient)</p>
  </section>

  <!-- ═══════════════ Feature Cards ═══════════════ -->
  <section class="features">
    <div class="features-grid">
      <!-- Code Assistant -->
      <div class="feature-card">
        <div class="feature-icon feature-icon-mauve">
          <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"
            />
          </svg>
        </div>
        <h3 class="feature-title">Code Assistant</h3>
        <p class="feature-desc">
          Writes code that actually works (sometimes). Supports every language you've heard of and
          several you haven't.
        </p>
      </div>

      <!-- Deep Research -->
      <div class="feature-card">
        <div class="feature-icon feature-icon-lavender">
          <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
            />
          </svg>
        </div>
        <h3 class="feature-title">Deep Research</h3>
        <p class="feature-desc">
          Googles things so you don't have to. Synthesizes entire Stack Overflow threads into one
          coherent answer.
        </p>
      </div>

      <!-- Self-Hosted -->
      <div class="feature-card">
        <div class="feature-icon feature-icon-green">
          <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"
            />
          </svg>
        </div>
        <h3 class="feature-title">Self-Hosted</h3>
        <p class="feature-desc">
          Running on bare metal, not some soulless cloud. Your data stays on hardware you can
          physically kick.
        </p>
      </div>
    </div>
  </section>

  <!-- ═══════════════ Footer ═══════════════ -->
  <footer class="site-footer">
    <div class="footer-inner">
      <p>Made with 💜 and an RTX 3060 by coah</p>
      <p>&copy; 2026 coahGPT Inc. (not actually incorporated)</p>
    </div>
  </footer>
</div>

<style>
  /* ───────── Layout ───────── */
  .landing {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    background: var(--color-base);
    position: relative;
    overflow-x: hidden;
  }

  .landing::before {
    content: '';
    position: absolute;
    top: -40%;
    left: 50%;
    transform: translateX(-50%);
    width: 140%;
    height: 80%;
    background: radial-gradient(
      ellipse at center,
      color-mix(in srgb, var(--color-mauve) 8%, transparent) 0%,
      color-mix(in srgb, var(--color-lavender) 4%, transparent) 30%,
      transparent 70%
    );
    pointer-events: none;
    z-index: 0;
  }

  /* ───────── Nav ───────── */
  .nav {
    position: sticky;
    top: 0;
    z-index: 50;
    border-bottom: 1px solid color-mix(in srgb, var(--color-surface0) 40%, transparent);
    background: color-mix(in srgb, var(--color-base) 80%, transparent);
    backdrop-filter: blur(20px) saturate(1.8);
    -webkit-backdrop-filter: blur(20px) saturate(1.8);
  }

  .nav-inner {
    max-width: 72rem;
    margin: 0 auto;
    padding: 0 1.5rem;
    height: 4rem;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .nav-logo {
    font-size: 1.25rem;
    font-weight: 800;
    color: var(--color-mauve);
    text-decoration: none;
    letter-spacing: -0.02em;
  }

  .nav-links {
    display: flex;
    align-items: center;
    gap: 1.25rem;
  }

  .nav-link {
    font-size: 0.875rem;
    color: var(--color-subtext0);
    text-decoration: none;
    transition: color 0.2s;
  }

  .nav-link:hover {
    color: var(--color-text);
  }

  .nav-cta {
    padding: 0.5rem 1.25rem;
    border-radius: 0.625rem;
    background: var(--color-mauve);
    color: var(--color-crust);
    font-size: 0.875rem;
    font-weight: 600;
    text-decoration: none;
    transition: opacity 0.2s;
  }

  .nav-cta:hover {
    opacity: 0.9;
  }

  /* ───────── Hero ───────── */
  .hero {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 6rem 1.5rem 4rem;
    position: relative;
    z-index: 1;
  }

  .hero-content {
    max-width: 52rem;
    margin: 0 auto;
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0;
  }

  /* Badge */
  .badge {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.375rem 1rem;
    border-radius: 9999px;
    background: color-mix(in srgb, var(--color-surface0) 50%, transparent);
    border: 1px solid color-mix(in srgb, var(--color-surface1) 40%, transparent);
    font-size: 0.8125rem;
    color: var(--color-subtext0);
    margin-bottom: 2.5rem;
  }

  .badge-dot {
    width: 0.5rem;
    height: 0.5rem;
    border-radius: 50%;
    background: var(--color-green);
    animation: pulse-dot 2s ease-in-out infinite;
  }

  @keyframes pulse-dot {
    0%,
    100% {
      opacity: 1;
      box-shadow: 0 0 0 0 color-mix(in srgb, var(--color-green) 40%, transparent);
    }
    50% {
      opacity: 0.7;
      box-shadow: 0 0 0 6px transparent;
    }
  }

  /* Heading */
  .hero-heading {
    font-family: 'Playfair Display', Georgia, 'Times New Roman', serif;
    font-size: clamp(3rem, 8vw, 5.5rem);
    font-weight: 800;
    letter-spacing: -0.03em;
    line-height: 1.1;
    color: var(--color-text);
    margin: 0 0 1.5rem;
  }

  .hero-heading-prefix {
    color: var(--color-text);
  }

  /* Subtitle */
  .hero-subtitle {
    font-size: 1.25rem;
    line-height: 1.7;
    color: var(--color-subtext0);
    max-width: 36rem;
    margin: 0 0 2.5rem;
  }

  /* ───────── Install Box ───────── */
  .install-box {
    display: inline-flex;
    align-items: stretch;
    background: var(--color-crust);
    border: 1px solid var(--color-surface0);
    border-radius: 1rem;
    overflow: hidden;
    box-shadow:
      0 0 0 1px color-mix(in srgb, var(--color-surface0) 30%, transparent),
      0 4px 24px color-mix(in srgb, #000 30%, transparent),
      0 1px 3px color-mix(in srgb, #000 20%, transparent);
    max-width: 100%;
    margin-bottom: 1.75rem;
  }

  .install-label {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.875rem 1.25rem;
    background: var(--color-text);
    color: var(--color-crust);
    font-weight: 600;
    font-size: 0.875rem;
    white-space: nowrap;
    cursor: default;
    user-select: none;
  }

  .install-chevron {
    width: 0.75rem;
    height: 0.75rem;
    opacity: 0.5;
  }

  .install-divider {
    width: 1px;
    background: var(--color-surface0);
  }

  .install-cmd {
    flex: 1;
    padding: 0.875rem 1.25rem;
    font-family: var(--font-mono);
    font-size: 0.9375rem;
    white-space: nowrap;
    user-select: all;
    display: flex;
    align-items: center;
  }

  .token-cmd {
    color: var(--color-mauve);
    font-weight: 600;
  }

  .token-flag {
    color: var(--color-peach);
  }

  .token-url {
    color: var(--color-text);
  }

  .token-pipe {
    color: var(--color-overlay0);
  }

  .install-copy {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0.875rem 1rem;
    background: transparent;
    border: none;
    color: var(--color-subtext0);
    cursor: pointer;
    transition: all 0.2s;
  }

  .install-copy:hover {
    color: var(--color-text);
    background: color-mix(in srgb, var(--color-surface0) 40%, transparent);
  }

  .copy-icon {
    width: 1.125rem;
    height: 1.125rem;
  }

  .copy-check {
    color: var(--color-green);
  }

  /* ───────── CTA Button ───────── */
  .hero-cta {
    display: inline-flex;
    align-items: center;
    gap: 0.625rem;
    padding: 0.875rem 2rem;
    border-radius: 0.875rem;
    background: var(--color-mauve);
    color: var(--color-crust);
    font-weight: 600;
    font-size: 1.0625rem;
    text-decoration: none;
    transition: all 0.25s;
    box-shadow: 0 2px 12px color-mix(in srgb, var(--color-mauve) 30%, transparent);
    margin-bottom: 2rem;
  }

  .hero-cta:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 20px color-mix(in srgb, var(--color-mauve) 40%, transparent);
    opacity: 0.95;
  }

  .cta-arrow {
    width: 1.125rem;
    height: 1.125rem;
    transition: transform 0.25s;
  }

  .hero-cta:hover .cta-arrow {
    transform: translateX(3px);
  }

  /* ───────── Callout ───────── */
  .callout {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.25rem;
  }

  .callout-badge {
    font-size: 0.6875rem;
    font-weight: 700;
    color: var(--color-green);
    text-transform: uppercase;
    letter-spacing: 0.1em;
  }

  .callout-name {
    font-size: 1.25rem;
    font-weight: 700;
    color: var(--color-text);
  }

  .callout-tagline {
    font-size: 0.875rem;
    color: var(--color-subtext0);
    font-style: italic;
  }

  /* ───────── Cat Section ───────── */
  .cat-section {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 2rem 1.5rem 5rem;
    position: relative;
    z-index: 1;
  }

  .cat-float {
    animation: float 4s ease-in-out infinite;
  }

  @keyframes float {
    0%,
    100% {
      transform: translateY(0);
    }
    50% {
      transform: translateY(-12px);
    }
  }

  .cat-caption {
    margin-top: 1.25rem;
    font-size: 0.875rem;
    color: var(--color-overlay0);
    font-style: italic;
  }

  /* ───────── Feature Cards ───────── */
  .features {
    padding: 0 1.5rem 6rem;
    position: relative;
    z-index: 1;
  }

  .features-grid {
    max-width: 64rem;
    margin: 0 auto;
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 1.5rem;
  }

  .feature-card {
    padding: 2rem;
    border-radius: 1.25rem;
    background: color-mix(in srgb, var(--color-mantle) 60%, transparent);
    border: 1px solid color-mix(in srgb, var(--color-surface0) 40%, transparent);
    transition: all 0.3s;
  }

  .feature-card:hover {
    border-color: color-mix(in srgb, var(--color-mauve) 25%, transparent);
    transform: translateY(-2px);
    box-shadow: 0 8px 32px color-mix(in srgb, #000 20%, transparent);
  }

  .feature-icon {
    width: 2.75rem;
    height: 2.75rem;
    border-radius: 0.75rem;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 1.25rem;
    transition: background 0.3s;
  }

  .feature-icon svg {
    width: 1.375rem;
    height: 1.375rem;
  }

  .feature-icon-mauve {
    background: color-mix(in srgb, var(--color-mauve) 12%, transparent);
    color: var(--color-mauve);
  }

  .feature-card:hover .feature-icon-mauve {
    background: color-mix(in srgb, var(--color-mauve) 20%, transparent);
  }

  .feature-icon-lavender {
    background: color-mix(in srgb, var(--color-lavender) 12%, transparent);
    color: var(--color-lavender);
  }

  .feature-card:hover .feature-icon-lavender {
    background: color-mix(in srgb, var(--color-lavender) 20%, transparent);
  }

  .feature-icon-green {
    background: color-mix(in srgb, var(--color-green) 12%, transparent);
    color: var(--color-green);
  }

  .feature-card:hover .feature-icon-green {
    background: color-mix(in srgb, var(--color-green) 20%, transparent);
  }

  .feature-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-text);
    margin: 0 0 0.625rem;
  }

  .feature-desc {
    font-size: 0.9375rem;
    line-height: 1.6;
    color: var(--color-subtext0);
    margin: 0;
  }

  /* ───────── Footer ───────── */
  .site-footer {
    border-top: 1px solid color-mix(in srgb, var(--color-surface0) 40%, transparent);
    padding: 2rem 1.5rem;
    position: relative;
    z-index: 1;
  }

  .footer-inner {
    max-width: 64rem;
    margin: 0 auto;
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 0.875rem;
    color: var(--color-overlay0);
  }

  .footer-inner p {
    margin: 0;
  }

  /* ───────── Responsive ───────── */
  @media (max-width: 768px) {
    .hero {
      padding: 4rem 1.5rem 3rem;
    }

    .hero-heading {
      font-size: clamp(2.25rem, 10vw, 3.5rem);
    }

    .hero-subtitle {
      font-size: 1.0625rem;
    }

    .install-box {
      flex-direction: column;
      align-items: stretch;
      border-radius: 0.875rem;
      width: 100%;
      max-width: 24rem;
    }

    .install-label {
      justify-content: center;
      padding: 0.75rem 1rem;
    }

    .install-divider {
      width: auto;
      height: 1px;
    }

    .install-cmd {
      padding: 0.75rem 1rem;
      font-size: 0.8125rem;
      overflow-x: auto;
    }

    .install-copy {
      padding: 0.75rem 1rem;
      justify-content: center;
    }

    .features-grid {
      grid-template-columns: 1fr;
      max-width: 28rem;
    }

    .footer-inner {
      flex-direction: column;
      gap: 0.5rem;
      text-align: center;
    }
  }
</style>
