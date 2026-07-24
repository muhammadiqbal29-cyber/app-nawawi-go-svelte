<script>
  import AlquranTab from './AlquranTab.svelte';
  import DiniyahTab from './DiniyahTab.svelte';
  import KeseharianTab from './KeseharianTab.svelte';
  import Icon from '$lib/components/ui/Icon.svelte';

  let activeTab = 'alquran';
</script>

<svelte:head>
  <title>Kelola Bank Catatan | Admin</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div>
      <h1 class="page-title">Kelola Bank Catatan</h1>
      <p class="page-subtitle">Atur bank catatan rapor otomatis untuk berbagai jenis penilaian</p>
    </div>
  </div>

  <div class="tabs">
    <button class="tab-btn {activeTab === 'alquran' ? 'active' : ''}" onclick={() => activeTab = 'alquran'}>
      <Icon name="fileText" size={18} />
      Alquran (Tahfiz)
    </button>
    <button class="tab-btn {activeTab === 'diniyah' ? 'active' : ''}" onclick={() => activeTab = 'diniyah'}>
      <Icon name="bookOpen" size={18} />
      Diniyah
    </button>
    <button class="tab-btn {activeTab === 'keseharian' ? 'active' : ''}" onclick={() => activeTab = 'keseharian'}>
      <Icon name="clipboard" size={18} />
      Keseharian
    </button>
  </div>

  <div class="tab-content">
    {#if activeTab === 'alquran'}
      <AlquranTab />
    {:else if activeTab === 'diniyah'}
      <DiniyahTab />
    {:else if activeTab === 'keseharian'}
      <KeseharianTab />
    {/if}
  </div>
</div>

<style>
  .page-container {
    padding: 2rem;
    max-width: 1200px;
    margin: 0 auto;
  }

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 2rem;
  }

  .page-title {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--color-text, #1e293b);
    margin: 0 0 0.5rem 0;
  }

  .page-subtitle {
    color: var(--color-text-muted, #64748b);
    margin: 0;
  }

  .tabs {
    display: flex;
    gap: 1rem;
    margin-bottom: 1.5rem;
    border-bottom: 1px solid var(--color-border, #e2e8f0);
    padding-bottom: 0.5rem;
  }

  .tab-btn {
    background: none;
    border: none;
    padding: 0.75rem 1.25rem;
    cursor: pointer;
    font-weight: 600;
    color: var(--color-text-muted, #64748b);
    display: flex;
    align-items: center;
    gap: 0.5rem;
    border-radius: 6px;
    transition: all 0.2s;
  }

  .tab-btn:hover {
    background: var(--color-surface-hover, #f1f5f9);
    color: var(--color-text, #1e293b);
  }

  .tab-btn.active {
    color: var(--color-primary, #22c55e);
    background: #dcfce7;
  }

  .tab-content {
    background: transparent;
  }
</style>
