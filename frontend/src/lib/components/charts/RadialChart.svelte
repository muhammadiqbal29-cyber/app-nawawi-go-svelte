<script>
  let { averages = {} } = $props();

  const data = $derived([
    { label: 'Hafalan', value: averages.hafalan || 0, color: '#0ea5e9' },
    { label: 'Lancar', value: averages.lancar || 0, color: '#eab308' },
    { label: 'Hukum Mad', value: averages.mad || 0, color: '#84cc16' },
    { label: 'Tajwid', value: averages.tajwid || 0, color: '#ef4444' },
    { label: 'Makhraj', value: averages.makhraj || 0, color: '#f97316' }
  ]);

  const size = 340;
  const center = size / 2;
  const strokeWidth = 14;
  const gap = 20;
  const maxRadius = size / 2 - 20;

  let totalAvg = $derived(
    data.length > 0 ? Math.round(data.reduce((sum, d) => sum + d.value, 0) / data.length) : 0
  );
</script>

<div class="radial-chart-container">
  <svg width="100%" height="100%" viewBox="0 0 {size} {size}" class="radial-svg">
    {#each data as item, i (item.label)}
      {@const r = maxRadius - i * gap}
      {@const c = 2 * Math.PI * r}
      {@const offset = c - (item.value / 100) * c}

      <circle
        cx={center}
        cy={center}
        {r}
        fill="none"
        stroke="#f1f5f9"
        stroke-width={strokeWidth}
        stroke-linecap="round"
      />

      <circle
        cx={center}
        cy={center}
        {r}
        fill="none"
        stroke={item.color}
        stroke-width={strokeWidth}
        stroke-dasharray="{c} {c}"
        stroke-dashoffset={offset}
        stroke-linecap="round"
        class="animated-circle"
      />
    {/each}

    <text x={center} y={center - 10} text-anchor="middle" class="avg-label">Rata-rata</text>
    <text x={center} y={center + 30} text-anchor="middle" class="avg-value">{totalAvg}%</text>
  </svg>
</div>

<div class="mini-indicators-container">
  {#each data as item (item.label)}
    <div class="mini-indicator group">
      <div class="mini-svg-wrapper">
        <svg viewBox="0 0 36 36" class="mini-svg">
          <path
            class="circle-bg"
            stroke="#f1f5f9"
            d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
          />
          <path
            class="circle-value"
            stroke={item.color}
            stroke-dasharray="{item.value}, 100"
            d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
          />
          <text x="18" y="21.5" class="mini-percentage">{item.value}%</text>
        </svg>
      </div>
      <span class="mini-label">{item.label}</span>
    </div>
  {/each}
</div>

<style>
  .radial-chart-container {
    width: 100%;
    height: 380px;
    display: flex;
    justify-content: center;
    align-items: center;
  }

  .radial-svg {
    max-width: 380px;
    max-height: 380px;
    transform: rotate(-90deg);
  }

  .animated-circle {
    transition: stroke-dashoffset 1.5s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .avg-label {
    fill: #64748b;
    font-size: 16px;
    font-family: 'Inter', sans-serif;
    transform: rotate(90deg);
    transform-origin: center;
  }

  .avg-value {
    fill: #0f172a;
    font-size: 42px;
    font-weight: 700;
    font-family: 'Outfit', sans-serif;
    transform: rotate(90deg);
    transform-origin: center;
  }

  .mini-indicators-container {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 1.5rem;
    margin-top: 1rem;
  }

  @media (min-width: 640px) {
    .mini-indicators-container {
      gap: 3rem;
    }
  }

  .mini-indicator {
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  .mini-svg-wrapper {
    position: relative;
    margin-bottom: 0.5rem;
    transition: transform 0.3s;
    width: 64px;
    height: 64px;
  }

  @media (min-width: 640px) {
    .mini-svg-wrapper {
      width: 70px;
      height: 70px;
    }
  }

  .mini-indicator:hover .mini-svg-wrapper {
    transform: scale(1.1);
  }

  .mini-svg {
    display: block;
    width: 100%;
    height: 100%;
  }

  .circle-bg {
    fill: none;
    stroke-width: 3.5;
  }

  .circle-value {
    fill: none;
    stroke-width: 3.5;
    stroke-linecap: round;
    transition: stroke-dasharray 1s ease-out;
  }

  .mini-percentage {
    fill: #0f172a;
    font-family: 'Outfit', sans-serif;
    font-weight: 700;
    font-size: 0.6em;
    text-anchor: middle;
  }

  .mini-label {
    font-size: 11px;
    font-weight: 600;
    color: #94a3b8;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
</style>
