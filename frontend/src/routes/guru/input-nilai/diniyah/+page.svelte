<script>
  import Icon from '$lib/components/ui/Icon.svelte';
  import { goto } from '$app/navigation';
  import { guruState } from '$lib/guruState.svelte.js';

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

  let kelasList = $state([]);
  let loadingKelas = $state(false);
  let error = $state('');

  let hasAccess = $derived(
    guruState.activeSekolahId &&
      guruState.aksesMap[guruState.activeSekolahId] &&
      guruState.aksesMap[guruState.activeSekolahId].includes('input_diniyah')
  );

  $effect(() => {
    if (guruState.activeSekolahId) {
      loadKelas();
    }
  });

  async function loadKelas() {
    loadingKelas = true;
    error = '';
    try {
      const res = await fetch(`${apiUrl}/api/guru/sekolah/${guruState.activeSekolahId}/kelas`, {
        credentials: 'include'
      });
      if (!res.ok) throw new Error('Gagal memuat data kelas.');
      const data = await res.json();
      kelasList = Array.isArray(data) ? data : [];
    } catch (err) {
      error = err.message || 'Gagal memuat data kelas.';
    } finally {
      loadingKelas = false;
    }
  }

  function kelasLabel(kelas) {
    return `${kelas.tingkat || ''} - ${kelas.nama_kelas || ''}`.trim();
  }

  function rombelLabel(rombel) {
    return !rombel || rombel === '__kosong__' ? 'Tanpa Rombel' : `Rombel ${rombel}`;
  }

  function cardHref(kelas) {
    const params = new URLSearchParams({
      kelas: kelasLabel(kelas),
      rombel: kelas.rombel || '__kosong__',
      sekolah_id: String(guruState.activeSekolahId)
    });
    return `/guru/input-nilai/diniyah/kelas/${kelas.id}?${params.toString()}`;
  }
</script>

<svelte:head>
  <title>Nilai Diniyah - SIAKAD</title>
</svelte:head>

{#if !guruState.activeSekolahId}
  <div class="empty-state">
    <Icon name="grid" size={48} />
    <h3>Pilih Jenjang Sekolah</h3>
    <p>Silakan pilih jenjang sekolah pada menu di pojok kanan atas.</p>
  </div>
{:else if !hasAccess}
  <div class="empty-state">
    <Icon name="alertCircle" size={48} />
    <h3>Akses Ditolak</h3>
    <p>
      Anda tidak memiliki hak akses <b>Nilai Diniyah</b> untuk jenjang {guruState.activeSekolahName}.
    </p>
  </div>
{:else}
  {#key guruState.activeSekolahId}
    <div class="page-container">
      <div class="page-header-actions">
        <div>
          <h1 class="page-title">Nilai Diniyah</h1>
          <p class="page-subtitle">Pilih kelas untuk menginput nilai mata pelajaran diniyah.</p>
        </div>
      </div>

      {#if error}
        <div class="alert-error">{error}</div>
      {/if}

      <div class="kelas-section">
        <div class="section-heading">
          <div>
            <h2>Pilih Kelas</h2>
            <p>Klik card rombel untuk membuka halaman input nilai.</p>
          </div>
          <div class="selected-info">
            {#if loadingKelas}
              Memuat kelas...
            {:else}
              {kelasList.length} rombel tersedia
            {/if}
          </div>
        </div>

        {#if loadingKelas}
          <div class="kelas-loading">Memuat data kelas...</div>
        {:else if kelasList.length > 0}
          <div class="kelas-grid" aria-label="Daftar kelas">
            {#each kelasList as kelas (kelas.id + kelas.rombel)}
              <a href={cardHref(kelas)} class="kelas-card">
                <div class="kelas-icon">
                  <Icon name="book" size={20} />
                </div>
                <div class="kelas-card-text">
                  <span class="kelas-name">{kelasLabel(kelas)}</span>
                  <span class="kelas-rombel">{rombelLabel(kelas.rombel)}</span>
                  <span class="kelas-count">{kelas.jumlah_siswa || 0} siswa</span>
                </div>
                <Icon name="chevronRight" size={18} class="kelas-arrow" />
              </a>
            {/each}
          </div>
        {:else}
          <div class="kelas-loading">Belum ada kelas tersedia.</div>
        {/if}
      </div>
    </div>
  {/key}
{/if}

<style>
  .page-container {
    display: flex;
    flex-direction: column;
    width: 100%;
    padding-bottom: 2rem;
  }

  .page-header-actions {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .page-title {
    margin: 0;
    font-size: 1.5rem;
    color: var(--color-text, #1e293b);
    font-weight: 700;
  }

  .page-subtitle {
    margin: 0.25rem 0 0;
    font-size: 0.9rem;
    color: var(--color-text-muted, #64748b);
  }

  .kelas-section {
    background: #fff;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    padding: 1rem;
    margin-bottom: 1rem;
  }

  .section-heading {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  .section-heading h2 {
    margin: 0;
    font-size: 1rem;
    font-weight: 800;
    color: #1e293b;
  }

  .section-heading p {
    margin: 0.2rem 0 0;
    color: #64748b;
    font-size: 0.85rem;
  }

  .selected-info {
    color: #475569;
    font-size: 0.9rem;
    font-weight: 700;
    padding-top: 0.1rem;
    white-space: nowrap;
  }

  .kelas-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.75rem;
  }

  .kelas-card {
    min-height: 88px;
    display: grid;
    grid-template-columns: 40px minmax(0, 1fr) 18px;
    align-items: center;
    gap: 0.75rem;
    width: 100%;
    padding: 0.85rem;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    background: #f8fafc;
    color: inherit;
    font-family: inherit;
    text-align: left;
    cursor: pointer;
    text-decoration: none;
    transition:
      border-color 0.2s ease,
      box-shadow 0.2s ease,
      background-color 0.2s ease;
  }

  .kelas-card:hover {
    border-color: #c7d2fe;
    background: #fff;
    box-shadow: 0 6px 16px rgba(15, 23, 42, 0.06);
  }

  .kelas-icon {
    width: 40px;
    height: 40px;
    border-radius: 8px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: #fef3c7;
    color: #d97706;
  }

  .kelas-card-text {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .kelas-name {
    color: #1e293b;
    font-size: 0.92rem;
    font-weight: 800;
    overflow-wrap: anywhere;
  }

  .kelas-count {
    color: #64748b;
    font-size: 0.82rem;
    font-weight: 700;
  }

  .kelas-rombel {
    color: #d97706;
    font-size: 0.9rem;
    font-weight: 700;
  }

  .kelas-card :global(.kelas-arrow) {
    color: #94a3b8;
  }

  .kelas-loading {
    color: #64748b;
    font-size: 0.9rem;
    font-weight: 700;
    border: 1px dashed #cbd5e1;
    border-radius: 8px;
    padding: 0.85rem 1rem;
    background: #f8fafc;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 4rem 2rem;
    gap: 1rem;
    color: #64748b;
    background: white;
    border-radius: 1rem;
    border: 1px solid #e2e8f0;
  }
  .empty-state h3 {
    margin: 0;
    color: #0f172a;
  }

  .alert-error {
    margin-bottom: 1rem;
    padding: 0.75rem 1rem;
    border-radius: 8px;
    font-size: 0.875rem;
    font-weight: 700;
    background: #fef2f2;
    color: #dc2626;
    border: 1px solid #fecaca;
  }

  @media (max-width: 720px) {
    .page-header-actions,
    .section-heading {
      flex-direction: column;
      align-items: stretch;
    }
    .kelas-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
