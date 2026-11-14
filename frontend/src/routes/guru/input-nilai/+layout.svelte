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
      fiturGuru.includes('input_tasmi_v2') ||
      fiturGuru.includes('input_tahsin') ||
      fiturGuru.includes('input_tahsin_v2')
  );
  let isV2 = $derived.by(() => {
    const sekolah = guruState.sekolahList.find(s => s.id == guruState.activeSekolahId);
    if (!sekolah || !sekolah.fitur) return false;
    try {
      const sf = JSON.parse(sekolah.fitur);
      return sf.includes('input_tahfiz_v2') || sf.includes('input_tasmi_v2') || sf.includes('input_tahsin_v2');
    } catch {
      return false;
    }
  });
  let hasDiniyah = $derived(fiturGuru.includes('input_diniyah'));
  let hasKeseharian = $derived(fiturGuru.includes('rapor_keseharian')); // Note: uses rapor_keseharian feature flag

  let tabs = $derived(
    [
      hasPaketAlquran && {
        id: 'alquran',
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

  function isTabActive(tabId) {
    if (tabId === 'alquran' || tabId === 'tahfiz' || tabId === 'tahfiz-v2') {
      return (
        currentPath.startsWith('/guru/input-nilai/tahfiz') ||
        currentPath.startsWith('/guru/input-nilai/tahsin') ||
        currentPath.startsWith('/guru/input-nilai/tasmi') ||
        currentPath.startsWith('/guru/input-nilai/kelulusan') ||
        currentPath === '/guru/input-nilai'
      );
    }
    if (tabId === 'diniyah') {
      return currentPath.startsWith('/guru/input-nilai/diniyah');
    }
    if (tabId === 'keseharian') {
      return currentPath.startsWith('/guru/input-nilai/keseharian');
    }
    return false;
  }
</script>

<div class="tabs-container">
  <nav class="nav-tabs-wrapper" aria-label="Navigasi Bidang Nilai">
    <div class="tabs-header">
      {#each tabs as tab (tab.id)}
        {@const active = isTabActive(tab.id)}
        <a
          href={tab.href}
          class="tab-link"
          class:active
          aria-current={active ? 'page' : undefined}
        >
          <div class="icon-wrapper">
            <Icon name={tab.icon} size={18} />
          </div>
          <span class="tab-label">{tab.label}</span>
          {#if active}
            <span class="active-badge">
              <span class="active-dot"></span>
              Aktif
            </span>
          {/if}
        </a>
      {/each}
    </div>
  </nav>

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

  .nav-tabs-wrapper {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    margin-bottom: 1.5rem;
  }

  .tabs-header {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.35rem 0.45rem;
    background: #f1f5f9; /* Slate 100 capsule bar */
    border: 1.5px solid #e2e8f0;
    border-radius: 14px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.03);
    overflow-x: auto;
    scrollbar-width: none;
    -ms-overflow-style: none;
  }

  .tabs-header::-webkit-scrollbar {
    display: none;
  }

  .icon-wrapper {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    border-radius: 8px;
    background: #ffffff;
    color: #64748b;
    transition: all 0.2s ease;
  }

  .tab-link {
    display: inline-flex;
    align-items: center;
    gap: 0.6rem;
    padding: 0.55rem 1.15rem;
    color: #475569;
    font-weight: 600;
    font-size: 0.95rem;
    text-decoration: none;
    border-radius: 10px;
    border: 1.5px solid transparent;
    background: transparent;
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
    white-space: nowrap;
    cursor: pointer;
  }

  .tab-link:hover {
    color: #0f172a;
    background: #ffffff;
    border-color: #cbd5e1;
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.04);
  }

  .tab-link:hover .icon-wrapper {
    background: #f8fafc;
    color: #0f172a;
  }

  .tab-link.active {
    background: #059669; /* Emerald 600 */
    border-color: #047857; /* Emerald 700 */
    color: #ffffff;
    font-weight: 700;
    box-shadow: 0 4px 12px -2px rgba(5, 150, 105, 0.35);
  }

  .tab-link.active .icon-wrapper {
    background: rgba(255, 255, 255, 0.22);
    color: #ffffff;
  }

  .tab-label {
    letter-spacing: -0.01em;
  }

  .active-badge {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    font-size: 0.68rem;
    font-weight: 700;
    padding: 2px 7px;
    border-radius: 999px;
    background: rgba(255, 255, 255, 0.25);
    color: #ffffff;
    letter-spacing: 0.02em;
    margin-left: 0.15rem;
    line-height: 1.2;
  }

  .active-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background-color: #ffffff;
    box-shadow: 0 0 6px rgba(255, 255, 255, 0.8);
  }

  .tab-content {
    flex: 1;
    display: flex;
    flex-direction: column;
  }

  @media (max-width: 768px) {
    .nav-tabs-wrapper {
      margin-bottom: 0.75rem;
    }

    .tabs-header {
      display: flex;
      width: 100%;
      background: transparent;
      border: none;
      box-shadow: none;
      padding: 0;
      gap: 8px;
    }

    .tab-link {
      flex: 1;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      gap: 6px;
      padding: 10px 6px;
      border-radius: 14px;
      background: #ffffff;
      border: 1.5px solid #e2e8f0;
      color: #64748b;
      margin-bottom: 0;
    }

    .tab-link:hover {
      background: #ffffff;
    }

    .tab-link.active {
      background: #059669;
      border-color: #047857;
      color: #ffffff;
      transform: translateY(-2px);
      box-shadow: 0 8px 20px -6px rgba(5, 150, 105, 0.4);
    }

    .icon-wrapper {
      width: 32px;
      height: 32px;
      border-radius: 10px;
      background: #f1f5f9;
      color: #64748b;
    }

    .tab-link.active .icon-wrapper {
      background: rgba(255, 255, 255, 0.22);
      color: #ffffff;
    }

    .tab-label {
      font-size: 11.5px;
      font-weight: 700;
    }

    .active-badge {
      display: none;
    }
  }
</style>
