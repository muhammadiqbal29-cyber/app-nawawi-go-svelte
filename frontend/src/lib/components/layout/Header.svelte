<script>
  import Icon from '$lib/components/ui/Icon.svelte';

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
  <div class="header-left">
    <button class="toggle-btn" onclick={onToggleSidebar} aria-label="Toggle Sidebar">
      <Icon name="bars" size={24} />
    </button>
    <h1 class="page-title">{title}</h1>
  </div>

  <div class="header-right">
    {#if userRole === 'guru'}
      <button
        class="active-jenjang-badge {!activeJenjang ? 'pulse' : ''}"
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
        class="active-jenjang-badge {!activeTahun ? 'pulse' : ''}"
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
      <Icon name="chevronDown" size={14} class="dropdown-arrow" />
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
</header>

<style>
  .header {
    height: var(--header-height, 70px);
    background-color: var(--color-surface, #ffffff);
    border-bottom: 1px solid var(--color-border, #e2e8f0);
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 2rem;
    position: sticky;
    top: 0;
    z-index: 50;
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
    font-weight: 700;
    color: var(--color-text-main, #1e293b);
    margin: 0;
  }

  .active-year-badge {
    background-color: var(--color-primary-light, #dcfce7);
    color: var(--color-primary-dark, #15803d);
    padding: 0.25rem 0.75rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 600;
  }

  .active-jenjang-badge {
    display: flex;
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
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background-color: rgba(34, 197, 94, 0.1);
    color: var(--color-primary, #22c55e);
    display: flex;
    justify-content: center;
    align-items: center;
    font-weight: 700;
    font-size: 1.2rem;
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

  @media (max-width: 768px) {
    .user-info {
      display: none;
    }
    .header {
      padding: 0 1rem;
    }
    .header-right {
      gap: 0.5rem;
    }
    .active-jenjang-badge {
      padding: 0.25rem 0.5rem;
      font-size: 0.7rem;
    }
  }
</style>
