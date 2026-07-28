<script>
  import { goto } from '$app/navigation';
  import Icon from '$lib/components/ui/Icon.svelte';

  let { isOpen = true, currentPath = '/guru', onLogout, fiturGuru = [], userRole = '', onClose = () => {} } = $props();

  // Fitur dari sekolah yang sedang aktif (fiturGuru di-passing dari layout)
  let hasPaketAlquran = $derived(fiturGuru.includes('paket_alquran') || fiturGuru.includes('input_tahfiz') || fiturGuru.includes('input_tahfiz_v2'));
  let hasRapor = $derived(fiturGuru.includes('cetak_rapor'));
  let hasDiniyah = $derived(fiturGuru.includes('input_diniyah'));
  let hasKonfigurasiMateri = $derived(fiturGuru.includes('konfigurasi_materi'));
  let hasKeseharian = $derived(fiturGuru.includes('rapor_keseharian'));

  function navigate(path) {
    goto(path);
    if (window.innerWidth <= 768) {
      onClose();
    }
  }
</script>

<aside class="sidebar" class:closed={!isOpen}>
  <div class="sidebar-header">
    <div class="logo">
      <Icon name="clipboard" size={20} />
    </div>
    <h2 class="brand" class:hidden={!isOpen}>GURU</h2>
  </div>

  <nav class="sidebar-nav">
    <ul>
      <li class:active={currentPath === '/guru'}>
        <button class="nav-link" onclick={() => navigate('/guru')}>
          <Icon name="grid" size={20} class="icon" />
          <span class="label" class:hidden={!isOpen}>Beranda</span>
        </button>
      </li>

      {#if hasPaketAlquran || hasDiniyah || hasKeseharian}
        <li class:active={currentPath.startsWith('/guru/input-nilai')}>
          <button class="nav-link" onclick={() => navigate('/guru/input-nilai')}>
            <Icon name="edit" size={20} class="icon" />
            <span class="label" class:hidden={!isOpen}>Input Nilai</span>
          </button>
        </li>
      {/if}

      {#if hasKonfigurasiMateri}
        <li class:active={currentPath.startsWith('/guru/konfigurasi')}>
          <button class="nav-link" onclick={() => navigate('/guru/konfigurasi')}>
            <Icon name="settings" size={20} class="icon" />
            <span class="label" class:hidden={!isOpen}>Konfigurasi</span>
          </button>
        </li>
      {/if}

      {#if hasRapor}
        <li class:active={currentPath.startsWith('/guru/rapor')}>
          <button class="nav-link" onclick={() => navigate('/guru/rapor')}>
            <Icon name="fileText" size={20} class="icon" />
            <span class="label" class:hidden={!isOpen}>Rapor</span>
          </button>
        </li>
      {/if}

      <li class:active={currentPath === '/guru/panduan'}>
        <button class="nav-link" onclick={() => navigate('/guru/panduan')}>
          <Icon name="helpCircle" size={20} class="icon" />
          <span class="label" class:hidden={!isOpen}>Panduan</span>
        </button>
      </li>

      <!-- Spacer to push Keluar to bottom -->
      <div style="flex-grow: 1;"></div>

      <li class:active={currentPath.startsWith('/guru/profil')}>
        <button class="nav-link" onclick={() => navigate('/guru/profil')}>
          <Icon name="user" size={20} class="icon" />
          <span class="label" class:hidden={!isOpen}>Profil Saya</span>
        </button>
      </li>

      <li>
        <button class="logout-btn" onclick={onLogout}>
          <Icon name="logOut" size={20} class="icon" />
          <span class="label" class:hidden={!isOpen}>Keluar</span>
        </button>
      </li>
    </ul>
  </nav>
</aside>

<style>
  .sidebar {
    width: var(--sidebar-width, 260px);
    background-color: var(--color-surface, #ffffff);
    border-right: 1px solid var(--color-border, #e2e8f0);
    height: 100vh;
    display: flex;
    flex-direction: column;
    transition: width 0.3s ease;
    position: fixed;
    left: 0;
    top: 0;
    z-index: 100;
    overflow-y: auto;
  }

  .sidebar.closed {
    width: 80px;
  }

  .sidebar-header {
    height: var(--header-height, 70px);
    display: flex;
    align-items: center;
    padding: 0 1.5rem;
    border-bottom: 1px solid var(--color-border, #e2e8f0);
  }

  .logo {
    width: 32px;
    height: 32px;
    background-color: var(--color-primary, #22c55e);
    color: white;
    border-radius: 8px;
    display: flex;
    justify-content: center;
    align-items: center;
    flex-shrink: 0;
  }

  .logo :global(.icon) {
    width: 20px;
    height: 20px;
  }

  .brand {
    margin-left: 1rem;
    font-size: 1.25rem;
    font-weight: 700;
    color: var(--color-primary, #22c55e);
    white-space: nowrap;
    opacity: 1;
    transition: opacity 0.2s;
  }

  .hidden {
    opacity: 0;
    pointer-events: none;
    display: none;
  }

  .sidebar-nav {
    padding: 1.5rem 1rem;
    flex: 1;
    display: flex;
    flex-direction: column;
  }

  .sidebar-nav ul {
    list-style: none;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    flex: 1;
    margin: 0;
    padding: 0;
  }

  .nav-link {
    display: flex;
    align-items: center;
    width: 100%;
    text-align: left;
    background: none;
    border: none;
    padding: 0.75rem 1rem;
    border-radius: 0.75rem;
    color: var(--color-text-muted, #64748b);
    font-weight: 500;
    transition: all 0.2s;
    cursor: pointer;
    font-family: inherit;
    font-size: 1rem;
  }

  .sidebar.closed .nav-link {
    justify-content: center;
    padding: 0.75rem 0;
  }

  .nav-link:hover {
    background-color: var(--color-surface-hover, #f1f5f9);
    color: var(--color-primary, #22c55e);
  }

  .sidebar-nav li.active .nav-link {
    background-color: rgba(34, 197, 94, 0.1);
    color: var(--color-primary, #22c55e);
  }

  .icon {
    font-size: 1.25rem;
    flex-shrink: 0;
  }

  .label {
    margin-left: 1rem;
    white-space: nowrap;
  }

  .logout-btn {
    display: flex;
    align-items: center;
    padding: 0.75rem 1rem;
    border-radius: 0.75rem;
    color: var(--color-error, #ef4444);
    font-weight: 500;
    transition: all 0.2s;
    width: 100%;
    border: none;
    background: none;
    cursor: pointer;
    font-family: inherit;
    font-size: 1rem;
  }

  .sidebar.closed .logout-btn {
    justify-content: center;
    padding: 0.75rem 0;
  }

  .logout-btn:hover {
    background-color: #fef2f2;
  }

  @media (max-width: 768px) {
    .sidebar {
      width: 260px;
      transform: translateX(0);
      z-index: 101; /* Above header and overlay */
      box-shadow: 2px 0 10px rgba(0,0,0,0.1);
    }
    .sidebar.closed {
      transform: translateX(-100%);
      width: 260px;
    }
    .label, .brand {
      opacity: 1 !important;
      display: inline-block !important;
      pointer-events: auto !important;
    }
  }
</style>
