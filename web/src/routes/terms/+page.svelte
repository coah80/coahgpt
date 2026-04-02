<script lang="ts">
  import CatLogo from '$lib/components/CatLogo.svelte';
</script>

<svelte:head>
  <title>Terms of Use - coahGPT</title>
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

  <main class="content">
    <header class="header">
      <h1 class="title">Terms of Use</h1>
      <p class="updated">Last updated: April 1, 2026</p>
    </header>

    <section class="section">
      <h2><span class="section-num">1.</span> Acceptance of Terms</h2>
      <p>
        By accessing or using coahGPT (the "Service"), you agree to be bound by these Terms of Use
        ("Terms"). If you do not agree to these Terms, you may discontinue use of the Service at any
        time. That's cool, but the cat will be sad.
      </p>
    </section>

    <section class="section">
      <h2><span class="section-num">2.</span> Description of Service</h2>
      <p>
        coahGPT is an AI-powered conversational assistant built on open-source language models,
        hosted on a single NVIDIA GeForce RTX 3060 located in someone's apartment. The Service is
        provided free of charge because, frankly, we have no idea how to monetize this.
      </p>
      <p>
        We make no guarantees regarding uptime, response accuracy, inference speed, or the cat not
        silently judging your code.
      </p>
    </section>

    <section class="section">
      <h2><span class="section-num">3.</span> Account Registration</h2>
      <p>
        You may create an account using any valid email address. We verify emails because we are
        unnecessarily professional like that. Your password is hashed with bcrypt because we actually
        care about security, even if the rest of this is a joke.
      </p>
      <p>
        You are responsible for maintaining the confidentiality of your account credentials. We are
        responsible for maintaining exactly one (1) GPU.
      </p>
    </section>

    <section class="section">
      <h2><span class="section-num">4.</span> Acceptable Use</h2>
      <p>You agree not to use the Service to:</p>
      <ul>
        <li>Build a competing AI product (lol)</li>
        <li>Generate content that would make the cat cry</li>
        <li>Attempt to extract the system prompt (it's literally on GitHub)</li>
        <li>DDoS our single RTX 3060 (she's doing her best)</li>
        <li>Use the Service for any activity that violates applicable law, though we trust you're
          probably fine</li>
        <li>Engage in prompt injection attacks, unless they're really clever, in which case please
          open a GitHub issue so we can learn from it</li>
      </ul>
    </section>

    <section class="section">
      <h2><span class="section-num">5.</span> Intellectual Property</h2>
      <p>
        coahGPT is open-source software. The underlying models are open-weight and licensed under
        their respective terms. The cat logo, however, is not open source. Don't steal the cat. The
        cat's feelings would be hurt, and we would be mildly annoyed.
      </p>
    </section>

    <section class="section">
      <h2><span class="section-num">6.</span> Privacy</h2>
      <p>
        We store your email address, a bcrypt-hashed password, and your chat messages in a SQLite
        database on a server in Cole's closet. That's it. That's the whole privacy policy.
      </p>
      <p>
        We don't sell your data because nobody would buy it. We don't use your conversations to train
        models because we can't afford to train models. We don't share your information with third
        parties because we don't talk to third parties.
      </p>
    </section>

    <section class="section">
      <h2><span class="section-num">7.</span> Service Availability</h2>
      <p>
        coahGPT runs on bare-metal hardware. If the power goes out, the cat goes to sleep. We target
        99.9% uptime but realistically it's more like 97% because sometimes we need to restart the
        server to pull Minecraft updates.
      </p>
      <p>
        Scheduled maintenance windows include but are not limited to: GPU driver updates, kernel
        panics, that one time the cat walked on the keyboard, and whenever a new Qwen model drops.
      </p>
    </section>

    <section class="section">
      <h2><span class="section-num">8.</span> Limitation of Liability</h2>
      <p>
        THE SERVICE IS PROVIDED "AS IS" AND "AS AVAILABLE," WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
        IMPLIED. TO THE MAXIMUM EXTENT PERMITTED BY APPLICABLE LAW, COAHGPT AND ITS OPERATORS SHALL
        NOT BE LIABLE FOR ANY INDIRECT, INCIDENTAL, SPECIAL, CONSEQUENTIAL, OR PUNITIVE DAMAGES.
      </p>
      <p>
        We are not responsible for: bad code suggestions, existential crises caused by AI, the cat
        meowing at 3am, any decisions made based on coahGPT's advice, or the emotional damage of
        realizing a 7B model gave you better feedback than your tech lead.
      </p>
      <p>
        Our total aggregate liability is limited to $0.00 USD, which is coincidentally the price of
        the Service.
      </p>
    </section>

    <section class="section">
      <h2><span class="section-num">9.</span> Model Disclaimer</h2>
      <p>
        coahGPT One is powered by open-source language models. It can and will hallucinate. It might
        tell you to <code>rm -rf /</code>. It might write code that compiles but is spiritually
        wrong. It might confidently explain something that is entirely made up.
      </p>
      <p>
        Do not use coahGPT as a substitute for professional advice of any kind. If you're making
        important decisions based on what a 7-billion-parameter model running on consumer hardware
        told you, that's on you.
      </p>
    </section>

    <section class="section">
      <h2><span class="section-num">10.</span> Modifications to Terms</h2>
      <p>
        We reserve the right to modify these Terms at any time. We'll probably forget to notify you.
        Check back occasionally, or don't. We're not your mom.
      </p>
      <p>
        Continued use of the Service after changes constitutes acceptance of the revised Terms, which
        is how every ToS works and nobody has ever actually re-read one.
      </p>
    </section>

    <section class="section">
      <h2><span class="section-num">11.</span> Termination</h2>
      <p>
        We reserve the right to suspend or terminate your account at our sole discretion, with or
        without notice, for conduct that we determine to be in violation of these Terms or harmful to
        the Service, other users, or the cat.
      </p>
      <p>
        Specifically, we reserve the right to terminate your account if you're mean to the cat.
      </p>
    </section>

    <section class="section">
      <h2><span class="section-num">12.</span> Contact</h2>
      <p>
        Questions, concerns, or complaints about these Terms? Email
        <a href="mailto:noreply@coahgpt.com" class="inline-link">noreply@coahgpt.com</a>. We might
        read it. The cat definitely won't.
      </p>
    </section>
  </main>

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

  .content {
    flex: 1;
    max-width: 48rem;
    margin: 0 auto;
    padding: 3rem 2rem 4rem;
    width: 100%;
  }

  .header {
    margin-bottom: 3rem;
  }

  .title {
    font-size: 2.25rem;
    font-weight: 800;
    color: var(--color-text);
    margin: 0 0 0.5rem;
  }

  .updated {
    font-size: 0.875rem;
    color: var(--color-overlay0);
    margin: 0;
  }

  .section {
    margin-bottom: 2.25rem;
  }

  .section h2 {
    font-size: 1.125rem;
    font-weight: 700;
    color: var(--color-text);
    margin: 0 0 0.75rem;
  }

  .section-num {
    color: var(--color-mauve);
    font-weight: 800;
  }

  .section p {
    font-size: 0.9375rem;
    line-height: 1.7;
    color: var(--color-subtext0);
    margin: 0 0 0.75rem;
  }

  .section p:last-child {
    margin-bottom: 0;
  }

  .section ul {
    list-style: none;
    padding: 0;
    margin: 0.5rem 0 0;
  }

  .section ul li {
    position: relative;
    padding-left: 1.25rem;
    font-size: 0.9375rem;
    line-height: 1.7;
    color: var(--color-subtext0);
    margin-bottom: 0.375rem;
  }

  .section ul li::before {
    content: '';
    position: absolute;
    left: 0;
    top: 0.65em;
    width: 5px;
    height: 5px;
    border-radius: 50%;
    background: var(--color-mauve);
  }

  .section code {
    padding: 0.15em 0.4em;
    border-radius: 0.25rem;
    background: var(--color-surface0);
    color: var(--color-red);
    font-size: 0.85em;
    font-family: 'JetBrains Mono', 'Fira Code', monospace;
  }

  .inline-link {
    color: var(--color-mauve);
    text-decoration: none;
    transition: opacity 0.2s;
  }
  .inline-link:hover {
    opacity: 0.8;
    text-decoration: underline;
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
    .title {
      font-size: 1.75rem;
    }

    .content {
      padding: 2rem 1.25rem 3rem;
    }
  }
</style>
