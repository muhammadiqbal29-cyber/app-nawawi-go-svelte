<script>
  import { precacheState } from '$lib/offlinePrecacher.svelte.js';
  import Icon from '$lib/components/ui/Icon.svelte';
</script>

{#if precacheState.isPrecaching}
  <div class="precache-indicator">
    <div class="icon-spin">
      <Icon name="refreshCw" size={14} />
    </div>
    <span class="label">{precacheState.label}</span>
    <div class="progress-bar-container">
      <div 
        class="progress-bar-fill" 
        style="width: {precacheState.total ? Math.min(100, Math.round((precacheState.progress / precacheState.total) * 100)) : 0}%">
      </div>
    </div>
  </div>
{/if}

<style>
  .precache-indicator {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: #f0fdf4;
    border: 1px solid #bbf7d0;
    padding: 0.35rem 0.75rem;
    border-radius: 20px;
    font-size: 0.8rem;
    color: #166534;
    font-weight: 600;
  }

  .icon-spin {
    display: flex;
    animation: spin 1.5s linear infinite;
  }

  @keyframes spin {
    100% {
      transform: rotate(360deg);
    }
  }

  .label {
    white-space: nowrap;
    max-width: 150px;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .progress-bar-container {
    width: 60px;
    height: 6px;
    background: #dcfce7;
    border-radius: 3px;
    overflow: hidden;
    margin-left: 0.25rem;
  }

  .progress-bar-fill {
    height: 100%;
    background: #22c55e;
    transition: width 0.3s ease;
  }

  @media (max-width: 768px) {
    .precache-indicator {
      display: none; /* Sembunyikan di header sempit, atau bisa digabung ke hamburger menu nanti */
    }
  }
</style>
