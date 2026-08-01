<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { auth } from '$lib/auth.svelte.js';
  import ToastContainer from '$lib/components/ui/ToastContainer.svelte';
  import ConfirmModal from '$lib/components/ui/ConfirmModal.svelte';
  import '../app.css';

  let { children } = $props();
  let checkingAuth = $state(true);

  onMount(async () => {
    // --- Global Fetch Interceptor untuk Anti-CSRF ---
    const originalFetch = window.fetch;
    window.fetch = async function () {
      let [resource, config] = arguments;
      if (config == null) {
        config = {};
      }

      const method = (config.method || 'GET').toUpperCase();
      const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';
      const isApiRequest = typeof resource === 'string' && resource.startsWith(apiUrl);

      if (isApiRequest && ['POST', 'PUT', 'DELETE', 'PATCH'].includes(method)) {
        if (config.headers == null) {
          config.headers = {};
        }
        if (config.headers instanceof Headers) {
          config.headers.append('X-Requested-With', 'XMLHttpRequest');
        } else {
          config.headers['X-Requested-With'] = 'XMLHttpRequest';
        }
      }
      return originalFetch.apply(this, [resource, config]);
    };
    // ------------------------------------------------

    const ctrl = new AbortController();
    const timer = setTimeout(() => ctrl.abort(), 5000);
    try {
      const res = await fetch(`${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/api/me`, {
        credentials: 'include',
        signal: ctrl.signal
      });
      if (res.ok) {
        const data = await res.json();
        auth.loggedInUser = data.role;
        auth.role = data.role;
        if (typeof window !== 'undefined') {
          localStorage.setItem('siakad_role', data.role);
        }
      } else if (res.status === 401 || res.status === 403) {
        // Clear local role if explicitly unauthorized
        if (typeof window !== 'undefined') {
          localStorage.removeItem('siakad_role');
        }
      }
    } catch (e) {
      console.error('Network Error checking auth:', e);
      // OFFLINE FALLBACK: if network fails, check if we have a saved role
      if (typeof window !== 'undefined') {
        const savedRole = localStorage.getItem('siakad_role');
        if (savedRole) {
          auth.loggedInUser = savedRole;
          auth.role = savedRole;
        }
      }
    } finally {
      clearTimeout(timer);
      checkingAuth = false;
    }
  });

  $effect(() => {
    if (!checkingAuth) {
      if (!auth.loggedInUser && $page.url.pathname !== '/login') {
        goto('/login');
      } else if (auth.loggedInUser && $page.url.pathname === '/login') {
        goto('/');
      }
    }
  });
</script>

<svelte:head>
  <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1">
</svelte:head>

{#if checkingAuth}
  <div class="loading-screen">
    <div class="loading-spinner"></div>
    <p class="loading-text">Memverifikasi sesi...</p>
  </div>
{:else}
  {@render children()}
{/if}

{#if import.meta.env.VITE_APP_ENV === 'staging'}
  <div class="staging-banner">
    ⚠️ MODE STAGING
  </div>
{/if}

<ToastContainer />
<ConfirmModal />

<style>
  .loading-screen {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    background: var(--color-bg-body, #f8fafc);
    gap: 1rem;
  }
  .loading-spinner {
    width: 40px;
    height: 40px;
    border: 3px solid var(--color-border, #e2e8f0);
    border-top-color: var(--color-primary, #22c55e);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }
  .loading-text {
    color: var(--color-text-muted, #64748b);
    font-size: 0.95rem;
  }
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .staging-banner {
    position: fixed;
    bottom: 1.5rem;
    right: 1.5rem;
    background: #ef4444;
    color: white;
    padding: 0.5rem 1rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: bold;
    letter-spacing: 0.05em;
    z-index: 9999;
    box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1);
    pointer-events: none;
    animation: pulse-staging 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }

  @keyframes pulse-staging {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.7; }
  }
</style>
