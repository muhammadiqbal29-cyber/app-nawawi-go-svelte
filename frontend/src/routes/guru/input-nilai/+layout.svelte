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
        <div class="icon-wrapper">
          <Icon name={tab.icon} size={18} />
        </div>
        <span>{tab.label}</span>
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
    scrollbar-width: none; /* Firefox */
    -ms-overflow-style: none; /* IE/Edge */
  }

  .tabs-header::-webkit-scrollbar {
    display: none; /* Safari/Chrome */
  }

  .icon-wrapper {
    display: flex;
    align-items: center;
    justify-content: center;
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

  @media (max-width: 768px) {
    .tabs-header {
      border-bottom: none;
      margin-bottom: 0.5rem;
      padding-bottom: 0.5rem;
      gap: 10px;
    }
    
    .tab-link {
      flex: 1;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      gap: 7px;
      padding: 12px 6px 10px;
      border-radius: 18px;
      background: var(--color-surface, #fff);
      border: 1px solid var(--color-border, #e2e8f0);
      margin-bottom: 0;
    }

    .tab-link:hover {
      background: var(--color-surface, #fff);
      border-radius: 18px;
    }

    .tab-link.active {
      background: var(--color-primary, #22c55e); /* Use primary green for consistency */
      border-color: var(--color-primary, #22c55e);
      color: #fff;
      transform: translateY(-2px);
      box-shadow: 0 10px 30px -12px rgba(34, 197, 94, 0.4); /* Green shadow */
    }
    
    /* Simulate the tile */
    .icon-wrapper {
      width: 34px;
      height: 34px;
      border-radius: 11px;
      background: #e6f7ec; /* Very soft green */
      color: var(--color-primary, #22c55e);
      transition: all 0.2s;
    }
    
    .tab-link.active .icon-wrapper {
      background: rgba(255,255,255,0.25);
      color: #fff;
    }

    /* Override span inside tab-link for mobile */
    .tab-link span {
      font-size: 11.5px;
      font-weight: 700;
    }
  }
</style>
