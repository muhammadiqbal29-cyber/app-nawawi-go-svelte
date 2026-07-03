<script>
  import { onMount } from 'svelte';
  import Icon from '$lib/components/ui/Icon.svelte';

  let { sekolahId = 1 } = $props();

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

  let currentTab = $state('tahfiz');
  let kelasList = $state([]);
  let loadingKelas = $state(false);
  let error = $state('');

  onMount(loadKelas);

  async function loadKelas(jenisPenilaian = selectedJenisPenilaian()) {
    loadingKelas = true;
    error = '';
    try {
      const params = new URLSearchParams({ jenis_penilaian: jenisPenilaian });
      const res = await fetch(
        `${apiUrl}/api/guru/sekolah/${sekolahId}/kelas?${params.toString()}`,
        {
          credentials: 'include'
        }
      );
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.message || 'Gagal memuat data kelas.');
      }
      const data = await res.json();
      kelasList = Array.isArray(data) ? data : [];
      await hydrateJumlahSiswa(jenisPenilaian);
    } catch (err) {
      error = err.message || 'Gagal memuat data kelas.';
    } finally {
      loadingKelas = false;
    }
  }

  async function hydrateJumlahSiswa(jenisPenilaian = selectedJenisPenilaian()) {
    await Promise.all(
      kelasList.map(async (kelas) => {
        if (Number(kelas.jumlah_siswa || 0) > 0) return;

        try {
          const params = new URLSearchParams({
            mata_pelajaran: 'Ujian Tahfiz',
            jenis_penilaian: jenisPenilaian,
            rombel: kelas.rombel || '__kosong__'
          });
          const res = await fetch(
            `${apiUrl}/api/guru/kelas/${kelas.id}/nilai?${params.toString()}`,
            {
              credentials: 'include'
            }
          );
          if (!res.ok) return;

          const body = await res.json();
          const jumlahSiswa = (body.siswa || []).length;
          kelasList = kelasList.map((item) =>
            String(item.id) === String(kelas.id) && item.rombel === kelas.rombel
              ? { ...item, jumlah_siswa: jumlahSiswa }
              : item
          );
        } catch {
          // Biarkan angka dari endpoint utama jika cek ulang gagal.
        }
      })
    );
  }

  function changeTab(tab) {
    currentTab = tab;
    if (tab === 'tahfiz' || tab === 'kelulusan') {
      loadKelas(tab === 'kelulusan' ? 'kelulusan' : 'semester');
    }
  }

  function selectedJenisPenilaian() {
    return currentTab === 'kelulusan' ? 'kelulusan' : 'semester';
  }

  function kelasLabel(kelas) {
    if (!kelas) return '';
    return `${kelas.tingkat || ''} - ${kelas.nama_kelas || ''}`.trim();
  }

  function jumlahSiswaLabel(total) {
    return `${Number(total || 0)} siswa`;
  }

  function rombelLabel(rombel) {
    return !rombel || rombel === '__kosong__' ? 'Tanpa Rombel' : `Rombel ${rombel}`;
  }

  function cardHref(kelas) {
    const params = new URLSearchParams({
      kelas: kelasLabel(kelas),
      rombel: kelas.rombel || '__kosong__',
      sekolah_id: String(sekolahId),
      jenis_penilaian: selectedJenisPenilaian()
    });
    return `/guru/input-nilai/kelas/${kelas.id}?${params.toString()}`;
  }
</script>

<div class="page-container">
  <div class="page-header-actions">
    <div>
      <h1 class="page-title">Input Nilai</h1>
      <p class="page-subtitle">Pilih kelas terlebih dahulu, lalu pilih siswa yang akan dinilai.</p>
    </div>
  </div>

  {#if error}
    <div class="alert-error">{error}</div>
  {/if}

  <div class="tabs-header card-tabs">
    <button
      class="tab-btn"
      class:active={currentTab === 'tahfiz'}
      onclick={() => changeTab('tahfiz')}
    >
      Ujian Tahfiz
    </button>
    <button
      class="tab-btn"
      class:active={currentTab === 'kelulusan'}
      onclick={() => changeTab('kelulusan')}
    >
      Ujian Kelulusan
    </button>
    <button
      class="tab-btn"
      class:active={currentTab === 'tahsin'}
      onclick={() => changeTab('tahsin')}
    >
      Penilaian Tahsin
    </button>
    <button
      class="tab-btn"
      class:active={currentTab === 'tasmi'}
      onclick={() => changeTab('tasmi')}
    >
      Ujian Tasmi'
    </button>
  </div>

  {#if currentTab === 'tahfiz' || currentTab === 'kelulusan'}
    <div class="kelas-section">
      <div class="section-heading">
        <div>
          <h2>Pilih Kelas</h2>
          <p>Klik card rombel untuk membuka daftar siswa di halaman baru.</p>
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
                <Icon name="users" size={20} />
              </div>
              <div class="kelas-card-text">
                <span class="kelas-name">{kelasLabel(kelas)}</span>
                <span class="kelas-rombel">{rombelLabel(kelas.rombel)}</span>
                <span class="kelas-count">{jumlahSiswaLabel(kelas.jumlah_siswa)}</span>
              </div>
              <Icon name="chevronRight" size={18} class="kelas-arrow" />
            </a>
          {/each}
        </div>
      {:else}
        <div class="kelas-loading">
          {currentTab === 'kelulusan'
            ? 'Belum ada kelas yang dikonfigurasi untuk ujian kelulusan.'
            : 'Belum ada kelas tersedia.'}
        </div>
      {/if}
    </div>
  {:else}
    <div class="empty-state-feature">
      <Icon name="info" size={48} />
      <h3>Fitur Dalam Pengembangan</h3>
      <p>
        Mohon maaf, fitur {currentTab === 'tahsin' ? 'Penilaian Tahsin' : "Ujian Tasmi'"} sedang dalam
        tahap pengembangan dan akan segera hadir.
      </p>
    </div>
  {/if}
</div>

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

  .card-tabs {
    display: inline-flex;
    align-self: flex-start;
    background-color: var(--color-surface, #fff);
    border: 1px solid var(--color-border, #e2e8f0);
    border-radius: 12px;
    padding: 0.35rem;
    box-shadow: var(--shadow-sm, 0 1px 2px rgba(0, 0, 0, 0.05));
    margin-bottom: 1.5rem;
  }

  .tab-btn {
    padding: 0.6rem 1.5rem;
    font-size: 0.9rem;
    font-weight: 600;
    border: none;
    background: transparent;
    color: var(--color-text-muted, #64748b);
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s;
    font-family: inherit;
  }

  .tab-btn.active {
    background-color: var(--color-primary, #4f46e5);
    color: white;
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
    background: #e0f2fe;
    color: #0369a1;
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
    color: #0369a1;
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

  .empty-state-feature {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 280px;
    text-align: center;
    color: #64748b;
    background: #fff;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    padding: 2rem;
  }

  .empty-state-feature h3 {
    margin: 1rem 0 0.35rem;
    color: #1e293b;
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

    .selected-info {
      white-space: normal;
    }

    .kelas-grid {
      grid-template-columns: 1fr;
    }

    .card-tabs {
      display: flex;
      width: 100%;
      overflow-x: auto;
    }
  }
</style>
