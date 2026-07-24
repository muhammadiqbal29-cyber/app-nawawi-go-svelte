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

  let hasPaketAlquran = $derived(
    fiturGuru.includes('paket_alquran') ||
      fiturGuru.includes('input_tahfiz') ||
      fiturGuru.includes('input_tahfiz_v2') ||
      fiturGuru.includes('input_tasmi') ||
      fiturGuru.includes('input_tasmi_v2')
  );
  let isV2 = $derived.by(() => {
    const sekolah = guruState.sekolahList.find(s => s.id == guruState.activeSekolahId);
    if (!sekolah || !sekolah.fitur) return false;
    try {
      const sf = JSON.parse(sekolah.fitur);
      return sf.includes('input_tahfiz_v2') || sf.includes('input_tasmi_v2');
    } catch {
      return false;
    }
  });
  let hasDiniyah = $derived(fiturGuru.includes('input_diniyah'));
  let hasKeseharian = $derived(fiturGuru.includes('rapor_keseharian')); // Note: uses rapor_keseharian feature flag

  let tabs = $derived(
    [
      hasPaketAlquran && {
        id: isV2 ? 'tahfiz-v2' : 'tahfiz',
        label: "Al-Qur'an",
        href: isV2 ? '/guru/input-nilai/tahfiz-v2' : '/guru/input-nilai/tahfiz',
        icon: 'bookOpen'
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
      <a href={tab.href} class="tab-link" class:active={currentPath.startsWith(tab.href)}>
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
