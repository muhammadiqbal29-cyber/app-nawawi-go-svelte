<script>
  import Icon from '$lib/components/ui/Icon.svelte';
  import SyncIndicator from '$lib/components/ui/SyncIndicator.svelte';

  let {
    title = 'Dashboard',
    activeTahun = '',
    userName = 'User',
    userRole = 'Pengguna',
    activeJenjang = '',
    onGantiJenjang = () => {},
    onGantiTahun = () => {},
    onLogout,
    onToggleSidebar
  } = $props();

  let showDropdown = $state(false);
</script>

<header class="header">
  <div class="app-header-top">
    <div class="header-left">
      <button class="toggle-btn" onclick={onToggleSidebar} aria-label="Toggle Sidebar">
        <Icon name="bars" size={24} />
      </button>
      <h1 class="page-title">{title}</h1>
    </div>

    <div class="header-right">
      <SyncIndicator />
      {#if userRole === 'guru'}
        <button
          class="active-jenjang-badge desktop-badge {!activeJenjang ? 'pulse' : ''}"
          onclick={onGantiJenjang}
          aria-label="Ganti Jenjang"
        >
          <Icon name="grid" size={14} />
          <span>{activeJenjang ? `Jenjang: ${activeJenjang}` : 'Pilih Jenjang'}</span>
          <Icon name="chevronDown" size={14} />
        </button>
      {/if}
      {#if activeTahun}
        <button
          class="active-jenjang-badge desktop-badge {!activeTahun ? 'pulse' : ''}"
          onclick={onGantiTahun}
          aria-label="Ganti Tahun Ajaran"
          style="background-color: #dcfce7; color: #15803d; border-color: transparent;"
        >
          <Icon name="calendar" size={14} />
          <span>{activeTahun}</span>
          <Icon name="chevronDown" size={14} />
        </button>
      {/if}
      
      <!-- svelte-ignore a11y-click-events-have-key-events -->
      <!-- svelte-ignore a11y-interactive-supports-focus -->
      <div class="user-profile" onclick={() => (showDropdown = !showDropdown)} role="button">
        <div class="avatar">{userName.charAt(0).toUpperCase()}</div>
        <div class="user-info">
          <span class="user-name">{userName}</span>
          <span class="user-role">{userRole}</span>
        </div>
        <Icon name="chevronDown" size={14} class="dropdown-arrow desktop-only" />
      </div>

      {#if showDropdown}
        <div class="dropdown-menu" role="menu">
          <button class="dropdown-item logout" onclick={onLogout} role="menuitem">
            <Icon name="logOut" size={16} />
            <span>Keluar</span>
          </button>
        </div>
        <div class="dropdown-backdrop" onclick={() => (showDropdown = false)}></div>
      {/if}
    </div>
  </div>

  <div class="chip-row mobile-only">
    {#if userRole === 'guru'}
      <button class="filter-chip a" onclick={onGantiJenjang}>
        <Icon name="grid" size={12} />
        {activeJenjang ? activeJenjang : 'Pilih Jenjang'}
      </button>
    {/if}
    {#if activeTahun}
      <button class="filter-chip b" onclick={onGantiTahun}>
        <Icon name="calendar" size={12} />
        {activeTahun}
      </button>
    {/if}
  </div>
</header>

<style>
  .header {
    background-color: var(--color-surface, #ffffff);
    border-bottom: 1px solid var(--color-border, #e2e8f0);
    display: flex;
    flex-direction: column;
    padding: 0 2rem;
    position: sticky;
    top: 0;
    z-index: 50;
  }

  .app-header-top {
    height: var(--header-height, 70px);
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 1.5rem;
  }

  .toggle-btn {
    font-size: 1.5rem;
    color: var(--color-text-muted, #64748b);
    display: flex;
    align-items: center;
    justify-content: center;
    transition: color 0.2s;
    background: none;
    border: none;
    cursor: pointer;
    padding: 0.5rem;
  }

  .toggle-btn:hover {
    color: var(--color-primary, #22c55e);
  }

  .page-title {
    font-size: 1.25rem;
    font-weight: 800;
    color: var(--color-text-main, #1e293b);
    margin: 0;
    letter-spacing: -0.01em;
  }

  .desktop-badge {
    display: flex;
  }

  .active-jenjang-badge {
    align-items: center;
    gap: 0.5rem;
    background-color: #eef2ff;
    color: #4f46e5;
    padding: 0.35rem 0.85rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 600;
    border: 1px solid transparent;
    cursor: pointer;
    transition: all 0.2s;
  }

  .active-jenjang-badge:hover {
    background-color: #e0e7ff;
    border-color: #c7d2fe;
  }

  .header-right {
    position: relative;
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .user-profile {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    cursor: pointer;
    padding: 0.25rem 0.5rem;
    border-radius: 8px;
    transition: background 0.2s;
  }

  .user-profile:hover {
    background: var(--color-surface-hover, #f1f5f9);
  }

  .avatar {
    width: 38px;
    height: 38px;
    border-radius: 12px;
    background: linear-gradient(155deg, var(--color-primary, #22c55e), #15803d);
    color: #fff;
    display: flex;
    justify-content: center;
    align-items: center;
    font-weight: 800;
    font-size: 1.1rem;
  }

  .user-info {
    display: flex;
    flex-direction: column;
  }

  .user-name {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--color-text-main, #1e293b);
  }

  .user-role {
    font-size: 0.75rem;
    color: var(--color-text-muted, #64748b);
  }

  .dropdown-arrow {
    color: var(--color-text-muted, #64748b);
    transition: transform 0.2s;
  }

  .dropdown-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 98;
  }

  .dropdown-menu {
    position: absolute;
    top: calc(100% + 8px);
    right: 0;
    background: var(--color-surface, #ffffff);
    border: 1px solid var(--color-border, #e2e8f0);
    border-radius: 8px;
    box-shadow: var(--shadow-md, 0 4px 6px -1px rgba(0, 0, 0, 0.1));
    min-width: 160px;
    z-index: 99;
    overflow: hidden;
  }

  .dropdown-item {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.7rem 1rem;
    width: 100%;
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-text, #1e293b);
    cursor: pointer;
    transition: background 0.15s;
    border: none;
    background: none;
    font-family: inherit;
  }

  .dropdown-item:hover {
    background: var(--color-surface-hover, #f1f5f9);
  }

  .dropdown-item.logout {
    color: var(--color-error, #ef4444);
  }

  .dropdown-item.logout:hover {
    background: #fef2f2;
  }

  .mobile-only {
    display: none;
  }

  @media (max-width: 768px) {
    .desktop-only {
      display: none !important;
    }
    .desktop-badge {
      display: none !important;
    }
    .mobile-only {
      display: flex;
    }
    .user-info {
      display: none;
    }
    .header {
      padding: 0.5rem 1rem;
      border-bottom: none;
      background: rgba(248, 250, 252, 0.9);
      backdrop-filter: blur(8px);
    }
    .app-header-top {
      height: auto;
      padding-bottom: 0.5rem;
    }
    .toggle-btn {
      display: none !important;
    }
    .page-title {
      font-size: 1.2rem;
      max-width: none;
    }
    .header-left {
      gap: 0.75rem;
    }
    
    .chip-row {
      display: flex;
      gap: 8px;
      overflow-x: auto;
      scrollbar-width: none;
      padding-bottom: 2px;
      margin-top: 0.25rem;
      margin-bottom: 0.5rem;
    }
    .chip-row::-webkit-scrollbar {
      display: none;
    }
    .filter-chip {
      flex-shrink: 0;
      display: flex;
      align-items: center;
      gap: 6px;
      padding: 8px 12px;
      border-radius: 999px;
      font-family: inherit;
      font-weight: 600;
      font-size: 0.8rem;
      white-space: nowrap;
      border: none;
      cursor: pointer;
    }
    .filter-chip.a {
      background: #eef2ff;
      color: #4f46e5;
    }
    .filter-chip.b {
      background: var(--color-surface, #fff);
      color: var(--color-text-main, #1e293b);
      border: 1px solid var(--color-border, #e2e8f0);
    }
  }
</style>
