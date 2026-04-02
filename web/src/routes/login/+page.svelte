<script lang="ts">
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import CatLogo from '$lib/components/CatLogo.svelte';
  import { login, signup, verifyEmail, resendVerification, isLoggedIn } from '$lib/auth.js';

  type AuthState = 'signin' | 'signup' | 'verify';

  let state = $state<AuthState>('signin');
  let email = $state('');
  let password = $state('');
  let displayName = $state('');
  let loading = $state(false);
  let error = $state('');
  let success = $state('');

  let digits = $state<string[]>(['', '', '', '', '', '']);
  let digitInputs = $state<HTMLInputElement[]>([]);

  let resendCooldown = $state(false);

  const verificationCode = $derived(digits.join(''));

  onMount(() => {
    if (isLoggedIn()) {
      goto('/chat');
    }
  });

  function switchState(next: AuthState) {
    state = next;
    error = '';
    success = '';
  }

  function resetDigits() {
    digits = ['', '', '', '', '', ''];
  }

  function handleDigitInput(index: number, e: Event) {
    const input = e.target as HTMLInputElement;
    const value = input.value;

    if (value.length > 1) {
      const chars = value.replace(/\D/g, '').slice(0, 6).split('');
      const updated = [...digits];
      for (let i = 0; i < 6; i++) {
        updated[i] = chars[i] ?? updated[i];
      }
      digits = updated;
      const focusIdx = Math.min(chars.length, 5);
      digitInputs[focusIdx]?.focus();
      return;
    }

    if (!/^\d?$/.test(value)) {
      const updated = [...digits];
      updated[index] = '';
      digits = updated;
      return;
    }

    const updated = [...digits];
    updated[index] = value;
    digits = updated;

    if (value && index < 5) {
      digitInputs[index + 1]?.focus();
    }
  }

  function handleDigitKeydown(index: number, e: KeyboardEvent) {
    if (e.key === 'Backspace' && !digits[index] && index > 0) {
      e.preventDefault();
      const updated = [...digits];
      updated[index - 1] = '';
      digits = updated;
      digitInputs[index - 1]?.focus();
    }
    if (e.key === 'Enter') {
      e.preventDefault();
      handleVerify();
    }
  }

  function handleDigitPaste(e: ClipboardEvent) {
    e.preventDefault();
    const pasted = (e.clipboardData?.getData('text') ?? '').replace(/\D/g, '').slice(0, 6);
    if (!pasted) return;
    const chars = pasted.split('');
    const updated = [...digits];
    for (let i = 0; i < 6; i++) {
      updated[i] = chars[i] ?? '';
    }
    digits = updated;
    const focusIdx = Math.min(chars.length, 5);
    digitInputs[focusIdx]?.focus();
  }

  async function handleSignin() {
    error = '';
    success = '';

    if (!email.trim()) {
      error = 'Email is required';
      return;
    }
    if (!password.trim()) {
      error = 'Password is required';
      return;
    }

    loading = true;
    const result = await login(email.trim(), password);
    loading = false;

    if (!result.ok) {
      if (result.status === 403) {
        switchState('verify');
        success = 'Your email needs verification. Check your inbox.';
        return;
      }
      error = result.error ?? 'Login failed';
      return;
    }

    goto('/chat');
  }

  async function handleSignup() {
    error = '';
    success = '';

    if (!displayName.trim()) {
      error = 'Display name is required';
      return;
    }
    if (!email.trim()) {
      error = 'Email is required';
      return;
    }
    if (!password.trim()) {
      error = 'Password is required';
      return;
    }
    if (password.length < 8) {
      error = 'Password must be at least 8 characters';
      return;
    }

    loading = true;
    const result = await signup(email.trim(), displayName.trim(), password);
    loading = false;

    if (!result.ok) {
      error = result.error ?? 'Signup failed';
      return;
    }

    resetDigits();
    switchState('verify');
    success = 'Check your email for a verification code.';
  }

  async function handleVerify() {
    error = '';
    success = '';

    if (verificationCode.length !== 6) {
      error = 'Enter all 6 digits';
      return;
    }

    loading = true;
    const result = await verifyEmail(email.trim(), verificationCode);

    if (!result.ok) {
      loading = false;
      error = result.error ?? 'Verification failed';
      return;
    }

    const loginResult = await login(email.trim(), password);
    loading = false;

    if (!loginResult.ok) {
      success = 'Email verified! Sign in with your credentials.';
      switchState('signin');
      return;
    }

    goto('/chat');
  }

  async function handleResend() {
    if (resendCooldown) return;
    error = '';

    resendCooldown = true;
    const result = await resendVerification(email.trim());

    if (!result.ok) {
      error = result.error ?? 'Failed to resend code';
      resendCooldown = false;
      return;
    }

    success = 'Code resent! Check your inbox.';
    setTimeout(() => {
      resendCooldown = false;
    }, 30000);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault();
      if (state === 'signin') handleSignin();
      else if (state === 'signup') handleSignup();
    }
  }
</script>

<svelte:head>
  <title>
    {state === 'signin' ? 'Sign in' : state === 'signup' ? 'Create account' : 'Verify email'} - coahGPT
  </title>
</svelte:head>

<div class="min-h-screen bg-base flex items-center justify-center px-4 py-12">
  <div class="w-full max-w-sm">
    <div class="bg-mantle border border-surface0/60 rounded-2xl p-8 shadow-xl">
      <!-- Logo -->
      <div class="flex justify-center mb-6">
        <CatLogo size={56} />
      </div>

      <!-- Heading -->
      <h1 class="text-xl font-bold text-text text-center mb-8">
        {#if state === 'signin'}
          Sign in to <span class="text-mauve">coahGPT</span>
        {:else if state === 'signup'}
          Create your <span class="text-mauve">coahGPT</span> account
        {:else}
          Verify your email
        {/if}
      </h1>

      <!-- Error -->
      {#if error}
        <div class="mb-4 px-3 py-2 rounded-lg bg-red/10 border border-red/20 text-red text-sm">
          {error}
        </div>
      {/if}

      <!-- Success -->
      {#if success}
        <div class="mb-4 px-3 py-2 rounded-lg bg-green/10 border border-green/20 text-green text-sm">
          {success}
        </div>
      {/if}

      <!-- Sign In form -->
      {#if state === 'signin'}
        <form onsubmit={(e) => { e.preventDefault(); handleSignin(); }} class="space-y-4">
          <div>
            <label for="email" class="block text-sm font-medium text-subtext0 mb-1.5">
              Email address
            </label>
            <input
              id="email"
              type="email"
              bind:value={email}
              onkeydown={handleKeydown}
              placeholder="you@example.com"
              autocomplete="email"
              class="w-full px-3.5 py-2.5 rounded-xl bg-crust border border-surface0/80 text-text text-sm
                placeholder-overlay0 outline-none
                focus:border-mauve/60 focus:ring-1 focus:ring-mauve/30
                hover:border-surface1 transition-all"
            />
          </div>

          <div>
            <label for="password" class="block text-sm font-medium text-subtext0 mb-1.5">
              Password
            </label>
            <input
              id="password"
              type="password"
              bind:value={password}
              onkeydown={handleKeydown}
              placeholder="&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;"
              autocomplete="current-password"
              class="w-full px-3.5 py-2.5 rounded-xl bg-crust border border-surface0/80 text-text text-sm
                placeholder-overlay0 outline-none
                focus:border-mauve/60 focus:ring-1 focus:ring-mauve/30
                hover:border-surface1 transition-all"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            class="w-full py-2.5 rounded-xl bg-mauve text-crust text-sm font-semibold
              hover:bg-mauve/90 active:bg-mauve/80
              disabled:opacity-60 disabled:cursor-not-allowed
              transition-all cursor-pointer flex items-center justify-center gap-2"
          >
            {#if loading}
              <div class="w-4 h-4 border-2 border-crust/30 border-t-crust rounded-full animate-spin"></div>
            {/if}
            Sign in
          </button>
        </form>

        <div class="flex items-center gap-3 my-6">
          <div class="flex-1 h-px bg-surface0/60"></div>
          <span class="text-xs text-overlay0">or</span>
          <div class="flex-1 h-px bg-surface0/60"></div>
        </div>

        <button
          onclick={() => switchState('signup')}
          class="w-full py-2.5 rounded-xl bg-surface0/30 text-text text-sm font-medium
            hover:bg-surface0/50 transition-colors cursor-pointer border border-surface0/60"
        >
          Create account
        </button>

      <!-- Sign Up form -->
      {:else if state === 'signup'}
        <form onsubmit={(e) => { e.preventDefault(); handleSignup(); }} class="space-y-4">
          <div>
            <label for="displayName" class="block text-sm font-medium text-subtext0 mb-1.5">
              Display name
            </label>
            <input
              id="displayName"
              type="text"
              bind:value={displayName}
              onkeydown={handleKeydown}
              placeholder="coah"
              autocomplete="name"
              class="w-full px-3.5 py-2.5 rounded-xl bg-crust border border-surface0/80 text-text text-sm
                placeholder-overlay0 outline-none
                focus:border-mauve/60 focus:ring-1 focus:ring-mauve/30
                hover:border-surface1 transition-all"
            />
          </div>

          <div>
            <label for="signupEmail" class="block text-sm font-medium text-subtext0 mb-1.5">
              Email address
            </label>
            <input
              id="signupEmail"
              type="email"
              bind:value={email}
              onkeydown={handleKeydown}
              placeholder="you@example.com"
              autocomplete="email"
              class="w-full px-3.5 py-2.5 rounded-xl bg-crust border border-surface0/80 text-text text-sm
                placeholder-overlay0 outline-none
                focus:border-mauve/60 focus:ring-1 focus:ring-mauve/30
                hover:border-surface1 transition-all"
            />
          </div>

          <div>
            <label for="signupPassword" class="block text-sm font-medium text-subtext0 mb-1.5">
              Password
            </label>
            <input
              id="signupPassword"
              type="password"
              bind:value={password}
              onkeydown={handleKeydown}
              placeholder="&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;"
              autocomplete="new-password"
              class="w-full px-3.5 py-2.5 rounded-xl bg-crust border border-surface0/80 text-text text-sm
                placeholder-overlay0 outline-none
                focus:border-mauve/60 focus:ring-1 focus:ring-mauve/30
                hover:border-surface1 transition-all"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            class="w-full py-2.5 rounded-xl bg-mauve text-crust text-sm font-semibold
              hover:bg-mauve/90 active:bg-mauve/80
              disabled:opacity-60 disabled:cursor-not-allowed
              transition-all cursor-pointer flex items-center justify-center gap-2"
          >
            {#if loading}
              <div class="w-4 h-4 border-2 border-crust/30 border-t-crust rounded-full animate-spin"></div>
            {/if}
            Create account
          </button>
        </form>

        <div class="flex items-center gap-3 my-6">
          <div class="flex-1 h-px bg-surface0/60"></div>
          <span class="text-xs text-overlay0">or</span>
          <div class="flex-1 h-px bg-surface0/60"></div>
        </div>

        <p class="text-sm text-subtext0 text-center">
          Already have an account?
          <button
            onclick={() => switchState('signin')}
            class="text-mauve hover:text-mauve/80 font-medium cursor-pointer transition-colors"
          >
            Sign in
          </button>
        </p>

      <!-- Verify form -->
      {:else}
        <p class="text-sm text-subtext0 text-center mb-6">
          We sent a 6-digit code to <span class="text-text font-medium">{email}</span>
        </p>

        <form onsubmit={(e) => { e.preventDefault(); handleVerify(); }} class="space-y-6">
          <!-- 6-digit code input -->
          <div class="flex justify-center gap-2.5">
            {#each digits as digit, i}
              <input
                bind:this={digitInputs[i]}
                type="text"
                inputmode="numeric"
                maxlength="6"
                value={digit}
                oninput={(e) => handleDigitInput(i, e)}
                onkeydown={(e) => handleDigitKeydown(i, e)}
                onpaste={handleDigitPaste}
                class="digit-input"
                aria-label="Digit {i + 1}"
              />
            {/each}
          </div>

          <button
            type="submit"
            disabled={loading || verificationCode.length !== 6}
            class="w-full py-2.5 rounded-xl bg-mauve text-crust text-sm font-semibold
              hover:bg-mauve/90 active:bg-mauve/80
              disabled:opacity-60 disabled:cursor-not-allowed
              transition-all cursor-pointer flex items-center justify-center gap-2"
          >
            {#if loading}
              <div class="w-4 h-4 border-2 border-crust/30 border-t-crust rounded-full animate-spin"></div>
            {/if}
            Verify
          </button>
        </form>

        <div class="mt-6 flex flex-col items-center gap-3">
          <button
            onclick={handleResend}
            disabled={resendCooldown}
            class="text-sm text-mauve hover:text-mauve/80 font-medium cursor-pointer transition-colors
              disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {resendCooldown ? 'Code resent' : 'Resend code'}
          </button>

          <button
            onclick={() => { resetDigits(); switchState('signin'); }}
            class="text-sm text-subtext0 hover:text-text cursor-pointer transition-colors"
          >
            Back to sign in
          </button>
        </div>
      {/if}
    </div>

    <!-- Footer -->
    <p class="text-xs text-overlay0 text-center mt-6">
      By continuing, you agree to our <a href="/terms" class="text-subtext0 hover:text-mauve transition-colors">Terms of Use</a> and <a href="/privacy" class="text-subtext0 hover:text-mauve transition-colors">Privacy Policy</a>
    </p>
  </div>
</div>

<style>
  .digit-input {
    width: 3rem;
    height: 3.5rem;
    text-align: center;
    font-size: 1.25rem;
    font-weight: 600;
    border-radius: 0.75rem;
    background: var(--color-surface0);
    border: 1.5px solid var(--color-surface1);
    color: var(--color-text);
    outline: none;
    caret-color: var(--color-mauve);
    transition: all 0.15s;
  }

  .digit-input:focus {
    border-color: var(--color-mauve);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-mauve) 20%, transparent);
  }

  .digit-input::placeholder {
    color: var(--color-overlay0);
  }
</style>
