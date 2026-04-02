<script lang="ts">
  let { size = 120, glow = true }: { size?: number; glow?: boolean } = $props();
</script>

<div class="inline-flex items-center justify-center" style:width="{size}px" style:height="{size}px">
  <svg
    viewBox="0 0 200 200"
    width={size}
    height={size}
    xmlns="http://www.w3.org/2000/svg"
    class="drop-shadow-lg"
    style:filter={glow ? 'drop-shadow(0 0 24px rgba(203, 166, 247, 0.4))' : 'none'}
  >
    <!-- Left ear -->
    <polygon points="45,85 30,20 80,65" fill="#cba6f7" />
    <polygon points="55,78 42,35 75,68" fill="#b4befe" opacity="0.6" />

    <!-- Right ear -->
    <polygon points="155,85 170,20 120,65" fill="#cba6f7" />
    <polygon points="145,78 158,35 125,68" fill="#b4befe" opacity="0.6" />

    <!-- Head -->
    <ellipse cx="100" cy="110" rx="68" ry="62" fill="#cba6f7" />

    <!-- Inner face shading -->
    <ellipse cx="100" cy="115" rx="55" ry="48" fill="#b4befe" opacity="0.12" />

    <!-- Left eye -->
    <ellipse cx="75" cy="100" rx="16" ry="18" fill="#11111b" />
    <ellipse cx="75" cy="100" rx="13" ry="15" fill="#1e1e2e" />
    <ellipse cx="70" cy="94" rx="6" ry="7" fill="#cdd6f4" />
    <ellipse cx="79" cy="104" rx="3" ry="3.5" fill="#cdd6f4" opacity="0.6" />

    <!-- Right eye -->
    <ellipse cx="125" cy="100" rx="16" ry="18" fill="#11111b" />
    <ellipse cx="125" cy="100" rx="13" ry="15" fill="#1e1e2e" />
    <ellipse cx="120" cy="94" rx="6" ry="7" fill="#cdd6f4" />
    <ellipse cx="129" cy="104" rx="3" ry="3.5" fill="#cdd6f4" opacity="0.6" />

    <!-- Nose -->
    <ellipse cx="100" cy="120" rx="5" ry="3.5" fill="#f5c2e7" />

    <!-- Mouth -->
    <path d="M 91 127 Q 96 134 100 127" fill="none" stroke="#f5c2e7" stroke-width="2" stroke-linecap="round" />
    <path d="M 100 127 Q 104 134 109 127" fill="none" stroke="#f5c2e7" stroke-width="2" stroke-linecap="round" />

    <!-- Whiskers left -->
    <line x1="28" y1="106" x2="60" y2="114" stroke="#cdd6f4" stroke-width="1.5" opacity="0.6" />
    <line x1="26" y1="118" x2="60" y2="120" stroke="#cdd6f4" stroke-width="1.5" opacity="0.6" />
    <line x1="30" y1="130" x2="61" y2="126" stroke="#cdd6f4" stroke-width="1.5" opacity="0.6" />

    <!-- Whiskers right -->
    <line x1="172" y1="106" x2="140" y2="114" stroke="#cdd6f4" stroke-width="1.5" opacity="0.6" />
    <line x1="174" y1="118" x2="140" y2="120" stroke="#cdd6f4" stroke-width="1.5" opacity="0.6" />
    <line x1="170" y1="130" x2="139" y2="126" stroke="#cdd6f4" stroke-width="1.5" opacity="0.6" />

    <!-- Blush spots -->
    <ellipse cx="58" cy="122" rx="11" ry="6" fill="#f5c2e7" opacity="0.25" />
    <ellipse cx="142" cy="122" rx="11" ry="6" fill="#f5c2e7" opacity="0.25" />
  </svg>
</div>
