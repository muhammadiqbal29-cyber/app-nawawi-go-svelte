<script>
  import { goto } from '$app/navigation';
  import Icon from '$lib/components/ui/Icon.svelte';

  let { currentPath = '/guru', fiturGuru = [] } = $props();

  let hasPaketAlquran = $derived(fiturGuru.includes('paket_alquran') || fiturGuru.includes('input_tahfiz') || fiturGuru.includes('input_tahfiz_v2'));
  let hasRapor = $derived(fiturGuru.includes('cetak_rapor'));
  let hasDiniyah = $derived(fiturGuru.includes('input_diniyah'));
  let hasKeseharian = $derived(fiturGuru.includes('rapor_keseharian'));
  let hasInputNilai = $derived(hasPaketAlquran || hasDiniyah || hasKeseharian);

  function navigate(path) {
    goto(path);
  }
</script>

<div class="bottom-nav-container">
  <div class="bottom-nav">
    <button class="nav-item" class:active={currentPath === '/guru'} onclick={() => navigate('/guru')}>
      <Icon name="grid" size={22} class="nav-icon" />
      <span>Beranda</span>
    </button>
    
    {#if hasInputNilai}
      <button class="nav-item" class:active={currentPath.startsWith('/guru/input-nilai')} onclick={() => navigate('/guru/input-nilai')}>
        <Icon name="edit" size={22} class="nav-icon" />
        <span>Input Nilai</span>
      </button>
    {/if}

    {#if hasDiniyah}
      <button class="nav-item" class:active={currentPath.startsWith('/guru/analitik')} onclick={() => navigate('/guru/analitik')}>
        <Icon name="barChart2" size={22} class="nav-icon" />
        <span>Analitik</span>
      </button>
    {/if}
    
    {#if hasRapor}
      <button class="nav-item" class:active={currentPath.startsWith('/guru/rapor')} onclick={() => navigate('/guru/rapor')}>
        <Icon name="fileText" size={22} class="nav-icon" />
        <span>Rapor</span>
      </button>
    {/if}

    <button class="nav-item" class:active={currentPath.startsWith('/guru/profil')} onclick={() => navigate('/guru/profil')}>
      <Icon name="user" size={22} class="nav-icon" />
      <span>Profil</span>
    </button>
  </div>
</div>

<style>
  .bottom-nav-container {
    display: none; /* Hidden on desktop */
  }

  @media (max-width: 768px) {
    .bottom-nav-container {
      display: block;
      position: fixed;
      left: 0;
      right: 0;
      bottom: 0;
      padding: 0 14px 16px;
      z-index: 1000;
      pointer-events: none; /* So clicks pass through padding area */
    }

    .bottom-nav {
      pointer-events: auto; /* Re-enable clicks for the bar itself */
      background: rgba(255, 255, 255, 0.95);
      backdrop-filter: blur(14px);
      -webkit-backdrop-filter: blur(14px);
      border: 1px solid var(--color-border, #e2e8f0);
      border-radius: 22px;
      display: flex;
      padding: 8px;
      box-shadow: 0 18px 40px -18px rgba(0, 0, 0, 0.18);
    }

    .nav-item {
      flex: 1;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      gap: 4px;
      padding: 8px 4px;
      border-radius: 16px;
      color: var(--color-text-muted, #64748b);
      background: transparent;
      border: none;
      cursor: pointer;
      position: relative;
      transition: all 0.2s ease;
    }

    .nav-item span {
      font-weight: 700;
      font-size: 10.5px;
      font-family: inherit;
    }

    .nav-item.active {
      background: var(--color-primary, #22c55e);
      color: #fff;
    }

    .nav-item.active span {
      color: #fff;
    }
  }
</style>
