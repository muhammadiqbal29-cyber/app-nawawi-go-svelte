<script>
  import { goto } from '$app/navigation';
  import Icon from '$lib/components/ui/Icon.svelte';

  let { isOpen = true, currentPath = '/', onLogout } = $props();
</script>

<aside class="sidebar" class:closed={!isOpen}>
  <div class="sidebar-header">
    <div class="logo">
      <Icon name="award" size={20} />
    </div>
    <h2 class="brand" class:hidden={!isOpen}>SIAKAD</h2>
  </div>

  <nav class="sidebar-nav">
    <ul>
      <li class:active={currentPath === '/'}>
        <button class="nav-link" onclick={() => goto('/')}>
          <Icon name="grid" size={20} class="icon" />
          <span class="label" class:hidden={!isOpen}>Beranda</span>
        </button>
      </li>
      <li class:active={currentPath.startsWith('/institusi')}>
        <button class="nav-link" onclick={() => goto('/institusi')}>
          <Icon name="database" size={20} class="icon" />
          <span class="label" class:hidden={!isOpen}>Institusi</span>
        </button>
      </li>
      <li class:active={currentPath.startsWith('/akademik')}>
        <button class="nav-link" onclick={() => goto('/akademik')}>
          <Icon name="users" size={20} class="icon" />
          <span class="label" class:hidden={!isOpen}>Siswa</span>
        </button>
      </li>

      <li class:active={currentPath.startsWith('/admin/guru')}>
        <button class="nav-link" onclick={() => goto('/admin/guru')}>
          <Icon name="users" size={20} class="icon" />
          <span class="label" class:hidden={!isOpen}>Guru</span>
        </button>
      </li>

      <li class:active={currentPath === '/presensi' || currentPath.startsWith('/presensi/konfigurasi') || currentPath.startsWith('/presensi/riwayat')}>
        <button class="nav-link" onclick={() => goto('/presensi')}>
          <Icon name="clock" size={20} class="icon" />
          <span class="label" class:hidden={!isOpen}>Presensi & Config</span>
        </button>
      </li>
      <li class:active={currentPath.startsWith('/presensi/izin')}>
        <button class="nav-link" onclick={() => goto('/presensi/izin')}>
          <Icon name="fileText" size={20} class="icon" />
          <span class="label" class:hidden={!isOpen}>Persetujuan Izin</span>
        </button>
      </li>
      <li class:active={currentPath.startsWith('/presensi/kalender')}>
        <button class="nav-link" onclick={() => goto('/presensi/kalender')}>
          <Icon name="calendar" size={20} class="icon" />
          <span class="label" class:hidden={!isOpen}>Kalender Libur</span>
        </button>
      </li>
      <li class:active={currentPath.startsWith('/kenaikan-kelas')}>
        <button class="nav-link" onclick={() => goto('/kenaikan-kelas')}>
          <Icon name="trendingUp" size={20} class="icon" />
          <span class="label" class:hidden={!isOpen}>Kenaikan Kelas</span>
        </button>
      </li>
      <li class:active={currentPath.startsWith('/admin/backup')}>
        <button class="nav-link" onclick={() => goto('/admin/backup')}>
          <Icon name="database" size={20} class="icon" />
          <span class="label" class:hidden={!isOpen}>Sistem Backup</span>
        </button>
      </li>
      <li class:active={currentPath.startsWith('/admin/bank-catatan')}>
        <button class="nav-link" onclick={() => goto('/admin/bank-catatan')}>
          <Icon name="fileText" size={20} class="icon" />
          <span class="label" class:hidden={!isOpen}>Bank Catatan</span>
        </button>
      </li>
      <li class:active={currentPath.startsWith('/admin/bak-sampah')}>
        <button class="nav-link" onclick={() => goto('/admin/bak-sampah')}>
          <Icon name="trash" size={20} class="icon" />
          <span class="label" class:hidden={!isOpen}>Bak Sampah</span>
        </button>
      </li>
      <li class:active={currentPath.startsWith('/admin/pengaturan-sistem')}>
        <button class="nav-link" onclick={() => goto('/admin/pengaturan-sistem')}>
          <Icon name="settings" size={20} class="icon" />
          <span class="label" class:hidden={!isOpen}>Pengaturan Sistem</span>
        </button>
      </li>

      <!-- Spacer to push Profil to bottom -->
      <div style="flex-grow: 1;"></div>

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
      display: none;
    }
  }
</style>
