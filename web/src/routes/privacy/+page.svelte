<script lang="ts">
  import CatLogo from '$lib/components/CatLogo.svelte';
</script>

<svelte:head>
  <title>Privacy Policy - coahGPT</title>
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
    <h1 class="title">Privacy Policy</h1>
    <p class="updated">Last updated: April 1, 2026</p>

    <section class="section">
      <h2>1. Information We Collect</h2>
      <p>When you use coahGPT, we collect the following:</p>
      <ul>
        <li><strong>Account info:</strong> your email address, display name, and a hashed version of your password</li>
        <li><strong>Chat data:</strong> your conversations with CoahGPT One</li>
        <li><strong>Usage data:</strong> timestamps and session tokens</li>
      </ul>
      <p>We do <strong>NOT</strong> collect:</p>
      <ul>
        <li>Your location</li>
        <li>Cookies (we're too lazy to implement tracking)</li>
        <li>Your browsing history</li>
        <li>Your deepest secrets (though you're welcome to share them with the cat)</li>
      </ul>
    </section>

    <section class="section">
      <h2>2. How We Store Your Data</h2>
      <ul>
        <li>Everything lives in a SQLite database on a single server in Cole's apartment</li>
        <li>Passwords are hashed with bcrypt (we're not animals)</li>
        <li>Session tokens are cryptographically random 32-byte hex strings</li>
        <li>Chat messages are stored in plain text because encrypting them would require effort</li>
        <li>Backups: lol</li>
      </ul>
    </section>

    <section class="section">
      <h2>3. How We Use Your Data</h2>
      <ul>
        <li>To provide the service (obviously)</li>
        <li>To send you verification emails from <span class="mono">noreply@coahgpt.com</span> via Proton Mail</li>
      </ul>
      <p>We do <strong>NOT</strong>:</p>
      <ul>
        <li>Use your data to train AI models (we literally cannot afford to)</li>
        <li>Sell your data (nobody has asked)</li>
        <li>Share your data with third parties (we don't know any third parties)</li>
      </ul>
    </section>

    <section class="section">
      <h2>4. Third-Party Services</h2>
      <ul>
        <li><strong>Cloudflare:</strong> DNS and tunnel routing (they see your traffic metadata)</li>
        <li><strong>Proton Mail:</strong> sends verification emails (they see your email address)</li>
        <li><strong>Ollama:</strong> runs locally, your prompts never leave the server</li>
      </ul>
      <p>That's it. No Google Analytics, no Facebook Pixel, no TikTok tracking. You're welcome.</p>
    </section>

    <section class="section">
      <h2>5. Data Retention</h2>
      <ul>
        <li><strong>Account data:</strong> kept until you delete your account or we forget the server exists</li>
        <li><strong>Chat messages:</strong> kept indefinitely in SQLite (it's a 0KB overhead, why would we delete them)</li>
        <li><strong>Session tokens:</strong> expire after 30 days</li>
        <li><strong>Verification codes:</strong> expire after 10 minutes</li>
      </ul>
    </section>

    <section class="section">
      <h2>6. Your Rights</h2>
      <ul>
        <li>You can request deletion of your account by emailing <span class="mono">noreply@coahgpt.com</span></li>
        <li>You can export your data by... actually we haven't built that yet</li>
        <li>You can stop using the service at any time (the cat will miss you)</li>
        <li><strong>GDPR:</strong> if you're in the EU, you have rights. We respect them. We just haven't built the UI for it yet.</li>
      </ul>
    </section>

    <section class="section">
      <h2>7. Security</h2>
      <ul>
        <li>bcrypt password hashing</li>
        <li>HTTPS via Cloudflare</li>
        <li>Session-based authentication</li>
        <li>Rate limiting on API endpoints</li>
        <li>The server is behind a locked door (Cole's apartment)</li>
      </ul>
      <p>We take security seriously. The rest of this project, not so much.</p>
    </section>

    <section class="section">
      <h2>8. Children's Privacy</h2>
      <p>coahGPT is not directed at children under 13. If you're under 13, please go play Minecraft instead. Actually, coahGPT would probably also tell you to go play Minecraft.</p>
    </section>

    <section class="section">
      <h2>9. Changes to This Policy</h2>
      <p>We may update this policy. We'll try to remember to change the date at the top. No promises.</p>
    </section>

    <section class="section">
      <h2>10. Contact</h2>
      <p>Email: <span class="mono">noreply@coahgpt.com</span></p>
      <p>The "noreply" in the address should tell you everything you need to know about our response times.</p>
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

  .title {
    font-size: 2rem;
    font-weight: 800;
    color: var(--color-text);
    margin: 0 0 0.5rem;
  }

  .updated {
    font-size: 0.875rem;
    color: var(--color-overlay0);
    margin: 0 0 2.5rem;
  }

  .section {
    margin-bottom: 2rem;
  }

  .section h2 {
    font-size: 1.25rem;
    font-weight: 700;
    color: var(--color-mauve);
    margin: 0 0 0.75rem;
  }

  .section p {
    font-size: 0.9375rem;
    line-height: 1.7;
    color: var(--color-subtext0);
    margin: 0 0 0.75rem;
  }

  .section ul {
    list-style: disc;
    padding-left: 1.5rem;
    margin: 0 0 0.75rem;
  }

  .section li {
    font-size: 0.9375rem;
    line-height: 1.7;
    color: var(--color-subtext0);
    margin-bottom: 0.25rem;
  }

  .section li strong {
    color: var(--color-text);
  }

  .mono {
    font-family: 'SF Mono', 'Fira Code', 'Cascadia Code', monospace;
    font-size: 0.875em;
    color: var(--color-text);
    background: rgba(69, 71, 90, 0.3);
    padding: 0.1em 0.4em;
    border-radius: 0.25rem;
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
      font-size: 1.5rem;
    }
    .content {
      padding: 2rem 1.25rem 3rem;
    }
  }
</style>
