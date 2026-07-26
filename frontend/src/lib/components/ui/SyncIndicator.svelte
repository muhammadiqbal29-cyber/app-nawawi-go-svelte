<script>
  import { syncManager } from '$lib/syncManager.svelte.js';
  import Icon from '$lib/components/ui/Icon.svelte';
</script>

{#if syncManager.queue.length > 0 || syncManager.isOffline || syncManager.isUnauthorized}
  {#if syncManager.isUnauthorized}
    <div class="sync-indicator unauthorized" title="Sesi kedaluwarsa. Nilai Anda aman di antrean lokal. Harap login kembali di tab baru lalu kembali ke sini.">
      <Icon name="alertTriangle" size={14} color="#dc2626" />
      <span>Sesi Berakhir ({syncManager.queue.length} tertunda)</span>
    </div>
  {:else}
    <div class="sync-indicator" class:offline={syncManager.isOffline} class:syncing={syncManager.isSyncing}>
      {#if syncManager.isOffline}
        <Icon name="cloudOff" size={14} color="#ef4444" />
        <span>Offline: {syncManager.queue.length} antrean</span>
      {:else if syncManager.isSyncing}
        <div class="spinner"></div>
        <span>Menyinkronkan {syncManager.queue.length}...</span>
      {:else}
        <Icon name="cloud" size={14} color="#f59e0b" />
        <span>{syncManager.queue.length} antrean</span>
      {/if}
    </div>
  {/if}
{/if}

<style>
  .sync-indicator {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.25rem 0.75rem;
    background: #fef3c7;
    border: 1px solid #fde68a;
    color: #b45309;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 500;
    margin-left: 0.5rem;
    box-shadow: 0 1px 2px rgba(0,0,0,0.05);
    transition: all 0.2s ease;
  }
  
  .sync-indicator.unauthorized {
    background: #fee2e2;
    border-color: #ef4444;
    color: #991b1b;
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
    cursor: help;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: .7; }
  }

  .sync-indicator.offline {
    background: #fee2e2;
    border-color: #fecaca;
    color: #b91c1c;
  }
  
  .sync-indicator.syncing {
    background: #e0f2fe;
    border-color: #bae6fd;
    color: #0369a1;
  }
  
  .spinner {
    width: 12px;
    height: 12px;
    border: 2px solid #bae6fd;
    border-top-color: #0284c7;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }
  
  @keyframes spin {
    to { transform: rotate(360deg); }
  }
</style>
