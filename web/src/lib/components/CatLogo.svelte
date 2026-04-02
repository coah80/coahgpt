<script lang="ts">
  let { size = 32, clickable = true, tooltipPosition = 'top' }: { size?: number; clickable?: boolean; tooltipPosition?: 'top' | 'bottom' } = $props();

  let meowing = $state(false);
  let showTooltip = $state(false);
  let showMeowTip = $state(false);

  function pet() {
    if (!clickable) return;
    meowing = true;
    showTooltip = false;
    showMeowTip = true;

    // play random minecraft cat meow with random pitch/speed variation
    try {
      const n = Math.floor(Math.random() * 5) + 2;
      const ctx = new AudioContext();
      fetch(`/meow${n}.ogg`)
        .then(r => r.arrayBuffer())
        .then(buf => ctx.decodeAudioData(buf))
        .then(decoded => {
          const source = ctx.createBufferSource();
          const gain = ctx.createGain();
          source.buffer = decoded;
          // random playback rate: 0.7 (deep slow) to 1.4 (chipmunk fast)
          source.playbackRate.value = 0.7 + Math.random() * 0.7;
          gain.gain.value = 0.5;
          source.connect(gain);
          gain.connect(ctx.destination);
          source.start();
          source.onended = () => ctx.close();
        })
        .catch(() => {});
    } catch {}

    setTimeout(() => {
      meowing = false;
      showMeowTip = false;
    }, 1200);
  }
</script>

<span
  class="cat-logo"
  class:clickable
  style:width="{size}px"
  style:height="{size}px"
  onmouseenter={() => { if (clickable && !meowing) showTooltip = true; }}
  onmouseleave={() => { showTooltip = false; }}
  onclick={pet}
  role={clickable ? 'button' : 'img'}
  tabindex={clickable ? 0 : -1}
>
  <img
    src={meowing ? '/cat-meow.png' : '/cat-normal.png'}
    alt="coahGPT"
    width={size}
    height={size}
    class="cat-img"
    class:shake={meowing}
  />
  {#if showTooltip}
    <span class="tooltip" class:tooltip-bottom={tooltipPosition === 'bottom'}>pet?</span>
  {/if}
  {#if showMeowTip}
    <span class="tooltip meow-tip" class:tooltip-bottom={tooltipPosition === 'bottom'}>meow!</span>
  {/if}
</span>

<style>
  .cat-logo {
    position: relative;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 6px;
    transition: transform 0.15s ease;
  }

  .clickable {
    cursor: pointer;
  }

  .clickable:hover {
    transform: scale(1.1);
  }

  .cat-img {
    border-radius: 6px;
    object-fit: contain;
  }

  .shake {
    animation: shake 0.4s ease-in-out;
  }

  @keyframes shake {
    0%, 100% { transform: rotate(0deg); }
    20% { transform: rotate(-8deg); }
    40% { transform: rotate(8deg); }
    60% { transform: rotate(-5deg); }
    80% { transform: rotate(5deg); }
  }

  .tooltip {
    position: absolute;
    bottom: calc(100% + 6px);
    left: 50%;
    transform: translateX(-50%);
    padding: 4px 10px;
    border-radius: 6px;
    background: var(--color-surface0, #313244);
    color: var(--color-text, #cdd6f4);
    font-size: 0.75rem;
    font-weight: 600;
    white-space: nowrap;
    pointer-events: none;
    animation: fade-in 0.15s ease;
  }

  .meow-tip {
    color: var(--color-mauve, #cba6f7);
    font-style: italic;
  }

  .tooltip-bottom {
    bottom: auto;
    top: calc(100% + 6px);
  }

  @keyframes fade-in {
    from { opacity: 0; transform: translateX(-50%) translateY(4px); }
    to { opacity: 1; transform: translateX(-50%) translateY(0); }
  }
</style>
