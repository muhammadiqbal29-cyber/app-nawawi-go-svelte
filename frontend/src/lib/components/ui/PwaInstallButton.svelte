<script>
  import { onMount } from 'svelte';
  import Icon from '$lib/components/ui/Icon.svelte';

  let deferredPrompt = $state(null);
  let showInstallButton = $state(false);

  onMount(() => {
    // Listen for the beforeinstallprompt event
    window.addEventListener('beforeinstallprompt', (e) => {
      // Prevent the mini-infobar from appearing on mobile
      e.preventDefault();
      // Stash the event so it can be triggered later.
      deferredPrompt = e;
      // Update UI notify the user they can install the PWA
      showInstallButton = true;
    });

    // Optionally, if it's already installed, don't show the button
    window.addEventListener('appinstalled', () => {
      showInstallButton = false;
      deferredPrompt = null;
    });
  });

  async function installApp() {
    if (deferredPrompt) {
      // Show the install prompt
      deferredPrompt.prompt();
      // Wait for the user to respond to the prompt
      const { outcome } = await deferredPrompt.userChoice;
      if (outcome === 'accepted') {
        console.log('User accepted the install prompt');
      } else {
        console.log('User dismissed the install prompt');
      }
      // We've used the prompt, and can't use it again, throw it away
      deferredPrompt = null;
      showInstallButton = false;
    }
  }
</script>

{#if showInstallButton}
  <button class="pwa-install-btn" onclick={installApp} aria-label="Install Aplikasi">
    <Icon name="download" size={20} />
    <span>Install Aplikasi</span>
  </button>
{/if}

<style>
  .pwa-install-btn {
    position: fixed;
    bottom: 80px; /* di atas bottom nav jika ada */
    right: 20px;
    z-index: 9999;
    background: linear-gradient(135deg, #0ea5e9, #3b82f6);
    color: white;
    border: none;
    border-radius: 50px;
    padding: 0.75rem 1.25rem;
    font-weight: 700;
    font-size: 0.875rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    box-shadow: 0 4px 15px rgba(14, 165, 233, 0.4);
    cursor: pointer;
    transition: transform 0.2s, box-shadow 0.2s;
  }
  .pwa-install-btn:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 20px rgba(14, 165, 233, 0.6);
  }
  @media (max-width: 768px) {
    .pwa-install-btn {
      bottom: 90px;
    }
  }
</style>
