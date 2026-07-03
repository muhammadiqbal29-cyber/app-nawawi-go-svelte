<script>
  let { data = [], color = '#0ea5e9', max = 100, labelFormat = (val) => val } = $props();

  let animationTriggered = $state(false);

  $effect(() => {
    // Small delay to trigger CSS transition on mount
    setTimeout(() => {
      animationTriggered = true;
    }, 50);
  });
</script>

<div class="bar-chart-container">
  <div class="y-axis">
    {#each [4, 3, 2, 1, 0] as i (i)}
      <div class="grid-line" style="bottom: {i * 25}%">
        <span class="y-label">{Math.round((max * i) / 4)}</span>
      </div>
    {/each}
  </div>

  <div class="bars-wrapper">
    {#each data as item (item.label)}
      <div class="bar-group">
        <div class="bar-track">
          <div
            class="bar-fill"
            style="height: {animationTriggered
              ? (item.value / max) * 100
              : 0}%; background-color: {color};"
          >
            <div class="bar-tooltip">
              <span style="font-weight: 700; font-size: 14px;">{item.label}</span>
              <span style="font-size: 13px; opacity: 0.9;">{labelFormat(item.value)}</span>
            </div>
          </div>
        </div>
        <span class="x-label" title={item.label}>
          {item.label.split(' ')[0]}
        </span>
      </div>
    {/each}
  </div>
</div>

<style>
  .bar-chart-container {
    position: relative;
    width: 100%;
    height: 350px;
    display: flex;
    padding-left: 30px;
    padding-bottom: 30px;
    padding-top: 20px;
    box-sizing: border-box;
  }

  .y-axis {
    position: absolute;
    top: 20px;
    bottom: 30px;
    left: 0;
    right: 0;
    pointer-events: none;
  }

  .grid-line {
    position: absolute;
    left: 0;
    right: 0;
    border-bottom: 1px dashed rgba(0, 0, 0, 0.08);
    display: flex;
    align-items: center;
  }

  .y-label {
    position: absolute;
    left: 0;
    transform: translateY(50%);
    font-size: 12px;
    color: #64748b;
    font-family: 'Inter', sans-serif;
  }

  .bars-wrapper {
    position: relative;
    flex-grow: 1;
    display: flex;
    justify-content: space-around;
    align-items: flex-end;
    margin-left: 15px;
    height: 100%;
  }

  .bar-group {
    display: flex;
    flex-direction: column;
    align-items: center;
    height: 100%;
    width: 15%;
    max-width: 40px;
  }

  .bar-track {
    flex-grow: 1;
    width: 100%;
    display: flex;
    align-items: flex-end;
    position: relative;
    border-radius: 6px;
  }

  .bar-fill {
    width: 100%;
    border-radius: 6px;
    transition:
      height 1.2s cubic-bezier(0.16, 1, 0.3, 1),
      filter 0.2s;
    position: relative;
    cursor: pointer;
  }

  .bar-fill:hover {
    filter: brightness(0.85);
  }

  .x-label {
    margin-top: 10px;
    font-size: 12px;
    font-weight: 500;
    color: #475569;
    font-family: 'Inter', sans-serif;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 100%;
    text-align: center;
  }

  .bar-tooltip {
    position: absolute;
    bottom: calc(100% + 10px);
    left: 50%;
    transform: translateX(-50%);
    background: #0f172a;
    color: white;
    padding: 8px 12px;
    border-radius: 6px;
    font-family: 'Outfit', sans-serif;
    white-space: nowrap;
    opacity: 0;
    pointer-events: none;
    transition:
      opacity 0.2s,
      transform 0.2s;
    z-index: 10;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.3);
  }

  .bar-tooltip::after {
    content: '';
    position: absolute;
    top: 100%;
    left: 50%;
    margin-left: -5px;
    border-width: 5px;
    border-style: solid;
    border-color: #0f172a transparent transparent transparent;
  }

  .bar-fill:hover .bar-tooltip {
    opacity: 1;
    transform: translateX(-50%) translateY(-5px);
  }
</style>
