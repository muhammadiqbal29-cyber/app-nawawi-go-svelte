<script>
  import { onMount } from 'svelte';
  import Icon from '$lib/components/ui/Icon.svelte';
  import '$lib/components/guru/AlquranMobileUI.css';

  onMount(() => {
    if (typeof window !== 'undefined') {
      sessionStorage.removeItem('rapor_ujian_tahfiz_kelas');
      sessionStorage.removeItem('rapor_ujian_tahfiz_v2_kelas');
      sessionStorage.removeItem('rapor_diniyah_kelas');
      sessionStorage.removeItem('rapor_keseharian_kelas');
    }
  });
  import { guruState } from '$lib/guruState.svelte.js';

  let fiturGuru = $derived(
    guruState.activeSekolahId ? guruState.aksesMap[guruState.activeSekolahId] || [] : []
  );

  let hasPaketAlquran = $derived(
    fiturGuru.includes('paket_alquran') ||
      fiturGuru.includes('input_tahfiz') ||
      fiturGuru.includes('input_tahfiz_v2')
  );
  let isV2 = $derived(fiturGuru.includes('input_tahfiz_v2'));

  let reportCards = $derived(
    [
      hasPaketAlquran && {
        title: 'Ujian Tahfiz',
        fullTitle: 'Rapor Ujian Tahfiz',
        description: 'Lihat rapor hasil ujian tahfiz siswa dan unduh cetakannya.',
        href: isV2 ? '/guru/rapor/ujian-tahfiz-v2' : '/guru/rapor/ujian-tahfiz',
        icon: 'award',
        abbr: 'UT',
        gradient: 'linear-gradient(155deg,#8B85F6,#6A63E6)',
        shadow: 'rgba(107,99,230,.4)'
      },
      hasPaketAlquran && {
        title: "Ujian Tasmi'",
        fullTitle: "Rapor Ujian Tasmi'",
        description: "Lihat rapor hasil ujian tasmi' siswa dan tinjau perkembangan bacaannya.",
        href: isV2 ? '/guru/rapor/tasmi-v2' : '/guru/rapor/tasmi',
        icon: 'bookOpen',
        abbr: 'TS',
        gradient: 'linear-gradient(155deg,#3FC97E,#0FA85E)',
        shadow: 'rgba(15,168,94,.4)'
      },
      hasPaketAlquran && {
        title: 'Ujian Kelulusan',
        fullTitle: 'Rapor Ujian Kelulusan',
        description: 'Lihat rapor ujian kelulusan siswa serta status kelulusan dan rekomendasi.',
        href: isV2 ? '/guru/rapor/kelulusan-v2' : '/guru/rapor/kelulusan',
        icon: 'checkSquare',
        abbr: 'UK',
        gradient: 'linear-gradient(155deg,#FFB25E,#F2942A)',
        shadow: 'rgba(242,148,42,.4)'
      },
      fiturGuru.includes('input_diniyah') && {
        title: 'Diniyah',
        fullTitle: 'Rapor Diniyah',
        description: 'Lihat laporan hasil pembelajaran mata pelajaran diniyyah siswa.',
        href: '/guru/rapor/diniyah',
        icon: 'book',
        abbr: 'RD',
        gradient: 'linear-gradient(155deg,#5FA8F5,#2F80ED)',
        shadow: 'rgba(47,128,237,.4)'
      },
      fiturGuru.includes('rapor_keseharian') && {
        title: 'Keseharian',
        fullTitle: 'Rapor Keseharian',
        description: 'Pantau catatan kedisiplinan dan nilai keseharian siswa.',
        href: '/guru/rapor/keseharian',
        icon: 'clipboardCheck',
        abbr: 'RK',
        gradient: 'linear-gradient(155deg,#F373B8,#DB4C9E)',
        shadow: 'rgba(219,76,158,.4)'
      },
      {
        title: 'Lainnya',
        fullTitle: 'Lainnya',
        description: 'Menu rapor tambahan.',
        href: '#',
        icon: 'plus',
        abbr: '+',
        gradient: 'var(--bg, #F6F8F7)',
        shadow: 'transparent',
        border: '1.5px dashed #C7CFC9',
        textColor: 'var(--muted, #7C8681)'
      }
    ].filter(Boolean)
  );
</script>

<div class="rapor-shell desktop-only">
  <div class="section-header">
    <div>
      <p class="eyebrow">Rapor</p>
      <h1>Jenis Rapor</h1>
    </div>
  </div>

  <div class="report-grid-desktop">
    {#each reportCards as card (card.href)}
      {#if card.title !== 'Lainnya'}
      <a class="report-card-desktop" href={card.href}>
        <div class="card-icon-desktop">
          <Icon name={card.icon} size={28} />
        </div>
        <div class="card-content-desktop">
          <h2>{card.fullTitle}</h2>
        </div>
      </a>
      {/if}
    {/each}
  </div>
</div>

<main class="sheet-page mobile-app-ui" style="padding-top: 0 !important;">
  <div class="mobile-app-header">
    <h1 class="page-title">Jenis Rapor</h1>
  </div>

  <div class="menu-grid">
    {#each reportCards as card (card.href)}
      <a class="menu-item" href={card.href}>
        <div class="menu-icon" style="background: {card.gradient}; --sh: {card.shadow}; {card.border ? `border: ${card.border};` : ''}">
          <span style="color: {card.border ? '#9AA39D' : 'white'}; font-weight: 800; font-size: 26px; font-family: 'Plus Jakarta Sans', sans-serif;">{card.abbr}</span>
        </div>
        <div class="menu-label" style="{card.textColor ? `color: ${card.textColor};` : ''}">{card.title}</div>
      </a>
    {/each}
  </div>
</main>

<style>
  /* Desktop & Mobile visibility toggles */
  .desktop-only { display: block; }

  @media (max-width: 768px) {
    .desktop-only { display: none !important; }
  }

  .rapor-shell {
    max-width: 1100px;
    margin: 0 auto;
    padding-bottom: 2rem;
  }

  .section-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 1.5rem;
    margin-bottom: 2rem;
  }

  .eyebrow {
    margin: 0 0 0.5rem;
    color: #12A150;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.12em;
    font-size: 0.75rem;
  }

  .section-header h1 {
    margin: 0;
    font-size: 2rem;
    color: #181B1A;
    font-weight: 800;
    letter-spacing: -0.01em;
  }

  .subtitle {
    margin: 0.5rem 0 0;
    color: #7C8681;
    line-height: 1.6;
    max-width: 680px;
  }

  /* Desktop Styles */
  .report-grid-desktop {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 1.5rem;
  }

  .report-card-desktop {
    display: block;
    background: #ffffff;
    border: 1px solid #e2e8f0;
    border-radius: 1rem;
    padding: 1.5rem;
    text-decoration: none;
    transition:
      transform 0.2s ease,
      box-shadow 0.2s ease,
      border-color 0.2s ease;
    color: inherit;
    min-height: 180px;
  }

  .report-card-desktop:hover {
    transform: translateY(-4px);
    box-shadow: 0 18px 40px rgba(18, 161, 80, 0.12);
    border-color: #12A150;
  }

  .card-icon-desktop {
    width: 48px;
    height: 48px;
    border-radius: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #E6F7EC;
    color: #0B7C3D;
    margin-bottom: 1.25rem;
  }

  .card-content-desktop h2 {
    margin: 0;
    font-size: 1.15rem;
    color: #111827;
  }

  .card-content-desktop p {
    margin: 0.75rem 0 0;
    color: #64748b;
    line-height: 1.7;
    font-size: 0.95rem;
  }

  /* Mobile Styles (App Launcher Grid) */
  .menu-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 20px 12px;
    padding-top: 10px;
  }

  .menu-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 9px;
    background: none;
    border: none;
    padding: 0;
    text-decoration: none;
    cursor: pointer;
  }

  .menu-icon {
    width: 100%;
    aspect-ratio: 1;
    border-radius: 22px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 10px 18px -10px var(--sh, rgba(0, 0, 0, .25));
    transition: transform .12s ease;
  }

  .menu-item:active .menu-icon {
    transform: scale(.92);
  }

  .menu-label {
    font-family: 'Plus Jakarta Sans', sans-serif;
    font-weight: 700;
    font-size: 12px;
    color: var(--text, #181B1A);
    text-align: center;
    line-height: 1.3;
  }

  /* Mobile Styles are now handled by AlquranMobileUI.css */
</style>

