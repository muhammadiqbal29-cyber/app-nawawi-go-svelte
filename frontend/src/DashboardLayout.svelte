<script>
  import Sidebar from '$lib/components/layout/Sidebar.svelte';
  import Header from '$lib/components/layout/Header.svelte';

  let {
    userName = 'Admin',
    userRole = 'admin',
    currentPath = '/',
    onNavigate,
    onLogout,
    children,
    title = 'Dashboard'
  } = $props();

  let sidebarOpen = $state(true);

  function toggleSidebar() {
    sidebarOpen = !sidebarOpen;
  }
</script>

<div class="layout-wrapper" class:sidebar-closed={!sidebarOpen}>
  <Sidebar isOpen={sidebarOpen} {currentPath} {onNavigate} {onLogout} />

  <div class="main-content">
    <Header {title} {userName} {userRole} onToggleSidebar={toggleSidebar} {onLogout} />

    <main class="content-area">
      {@render children()}
    </main>
  </div>
</div>

<style>
  .layout-wrapper {
    display: flex;
    min-height: 100vh;
    width: 100%;
    background-color: var(--color-bg-body, #f8fafc);
    transition: padding-left 0.3s ease;
    padding-left: var(--sidebar-width, 260px);
  }

  .layout-wrapper.sidebar-closed {
    padding-left: 80px;
  }

  .main-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0; /* Mencegah overflow pada flex item */
  }

  .content-area {
    padding: 2rem;
    flex: 1;
  }

  @media (max-width: 768px) {
    .layout-wrapper {
      padding-left: 0;
    }
    .layout-wrapper.sidebar-closed {
      padding-left: 0;
    }
    .content-area {
      padding: 1rem;
    }
  }
</style>
