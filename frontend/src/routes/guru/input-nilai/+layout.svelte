<script>
  import { page } from '$app/stores';
  import { guruState } from '$lib/guruState.svelte.js';
  import Icon from '$lib/components/ui/Icon.svelte';

  let { children } = $props();

  let currentPath = $derived($page.url.pathname);

  // Check access for tabs based on `fiturGuru`
  let fiturGuru = $derived(
    guruState.activeSekolahId ? guruState.aksesMap[guruState.activeSekolahId] || [] : []
  );

  let hasTahfizV1 = $derived(fiturGuru.includes('input_tahfiz'));
  let hasTahfizV2 = $derived(fiturGuru.includes('input_tahfiz_v2'));
  let hasDiniyah = $derived(fiturGuru.includes('input_diniyah'));
  let hasKeseharian = $derived(fiturGuru.includes('rapor_keseharian')); // Note: uses rapor_keseharian feature flag

  let tabs = $derived(
    [
      hasTahfizV1 && {
        id: 'tahfiz',
        label: 'Tahfiz (V1)',
        href: '/guru/input-nilai/tahfiz',
        icon: 'edit'
      },
      hasTahfizV2 && {
        id: 'tahfiz-v2',
        label: 'Tahfiz (V2)',
        href: '/guru/input-nilai/tahfiz-v2',
        icon: 'clipboardCheck'
      },
      hasDiniyah && {
        id: 'diniyah',
        label: 'Diniyah',
        href: '/guru/input-nilai/diniyah',
        icon: 'book'
      },
      hasKeseharian && {
        id: 'keseharian',
        label: 'Keseharian',
        href: '/guru/input-nilai/keseharian',
        icon: 'smile'
      }
    ].filter(Boolean)
  );
</script>

<div class="tabs-container">
  <div class="tabs-header">
    {#each tabs as tab (tab.id)}
      <a
        href={tab.href}
        class="tab-link"
        class:active={currentPath.startsWith(tab.href)}
      >
        <Icon name={tab.icon} size={16} />
        {tab.label}
      </a>
    {/each}
  </div>

  <div class="tab-content">
    {@render children()}
  </div>
</div>

<style>
  .tabs-container {
    display: flex;
    flex-direction: column;
    width: 100%;
    min-height: 500px;
  }

  .tabs-header {
    display: flex;
    gap: 0.5rem;
    border-bottom: 2px solid var(--color-border, #e2e8f0);
    margin-bottom: 1.5rem;
    overflow-x: auto;
  }

  .tab-link {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1.25rem;
    color: var(--color-text-muted, #64748b);
    font-weight: 700;
    font-size: 0.95rem;
    text-decoration: none;
    border-bottom: 3px solid transparent;
    margin-bottom: -2px; /* overlap the border */
    transition: all 0.2s;
    white-space: nowrap;
  }

  .tab-link:hover {
    color: var(--color-primary, #4f46e5);
    background-color: var(--color-background, #f8fafc);
    border-radius: 8px 8px 0 0;
  }

  .tab-link.active {
    color: var(--color-primary, #4f46e5);
    border-bottom-color: var(--color-primary, #4f46e5);
  }

  .tab-content {
    flex: 1;
    display: flex;
    flex-direction: column;
  }
</style>
