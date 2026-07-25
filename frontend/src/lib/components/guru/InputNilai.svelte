<script>
  import Icon from '$lib/components/ui/Icon.svelte';
  import { guruState } from '$lib/guruState.svelte.js';

  let { sekolahId = 1, isV2 = false, isTasmi = false } = $props();

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

  let currentTab = $state(isTasmi ? 'tasmi' : 'tahfiz');
  let kelasList = $state([]);
  let loadingKelas = $state(false);
  let error = $state('');

  // Watch for changes in sekolahId or guruState year/semester
  // We must read these values synchronously so Svelte 5 can track them
  $effect(() => {
    const taId = guruState.activeTahunId;
    const semId = guruState.activeSemesterId;
    const skId = sekolahId;
    if (taId && semId && skId) {
      loadKelas(selectedJenisPenilaian(), taId, semId);
    }
  });



  async function loadKelas(jenisPenilaian = selectedJenisPenilaian(), taId = guruState.activeTahunId, semId = guruState.activeSemesterId) {
    loadingKelas = true;
    error = '';
    try {
      const params = new URLSearchParams({ jenis_penilaian: jenisPenilaian });
      if (taId) params.append('tahun_ajaran_id', taId);
      if (semId) params.append('semester_id', semId);
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
      await hydrateJumlahSiswa(jenisPenilaian, taId, semId);
    } catch (err) {
      error = err.message || 'Gagal memuat data kelas.';
    } finally {
      loadingKelas = false;
    }
  }

  async function hydrateJumlahSiswa(jenisPenilaian = selectedJenisPenilaian(), taId = guruState.activeTahunId, semId = guruState.activeSemesterId) {
    await Promise.all(
      kelasList.map(async (kelas) => {
        if (Number(kelas.jumlah_siswa || 0) > 0) return;

        try {
          const params = new URLSearchParams({
            mata_pelajaran: currentTab === 'tasmi' ? 'Ujian Tasmi' : 'Ujian Tahfiz',
            jenis_penilaian: jenisPenilaian,
            rombel: kelas.rombel || '__kosong__'
          });
          if (taId) params.append('tahun_ajaran_id', taId);
          if (semId) params.append('semester_id', semId);
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
    if (tab === 'tahfiz' || tab === 'kelulusan' || tab === 'tasmi') {
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
    if (guruState.activeTahunId) params.append('tahun_ajaran_id', guruState.activeTahunId);
    if (guruState.activeSemesterId) params.append('semester_id', guruState.activeSemesterId);
    let basePath = isV2 ? '/guru/input-nilai/tahfiz-v2' : '/guru/input-nilai/tahfiz';
    if (currentTab === 'tasmi') {
      basePath = isV2 ? '/guru/input-nilai/tasmi-v2' : '/guru/input-nilai/tasmi';
    }
    return `${basePath}/kelas/${kelas.id}?${params.toString()}`;
  }
</script>

<div class="page-container">
  <div class="sticky-header">
    <div class="page-header-actions">
      <div>
        <h1 class="page-title">{currentTab === 'tasmi' ? (isV2 ? 'Input Nilai Tasmi ' : 'Input Nilai Tasmi ') : (isV2 ? 'Input Nilai Tahfiz ' : 'Input Nilai ')}</h1>
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
  </div>



  {#if currentTab === 'tahfiz' || currentTab === 'kelulusan' || currentTab === 'tasmi'}
    <div class="kelas-section">
      <div class="section-heading">
        <div>
          <h2>Pilih Kelas</h2>
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
              <div class="kelas-icon tingkat-{kelas.tingkat}">
                {kelas.tingkat || 'X'}
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
    background-color: #f1f5f9;
    border-radius: 999px;
    padding: 0.35rem;
    margin-bottom: 1.5rem;
    border: 1px solid #e2e8f0;
  }

  .tab-btn {
    padding: 0.6rem 1.5rem;
    font-size: 0.9rem;
    font-weight: 600;
    border: none;
    background: transparent;
    color: var(--color-text-muted, #64748b);
    border-radius: 999px;
    cursor: pointer;
    transition: all 0.3s ease;
    font-family: inherit;
    white-space: nowrap;
  }

  .tab-btn.active {
    background-color: white;
    color: var(--color-primary, #4f46e5);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
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
    grid-template-columns: 48px minmax(0, 1fr) 18px;
    align-items: center;
    gap: 1rem;
    width: 100%;
    padding: 1.25rem;
    border: 1px solid #f1f5f9;
    border-radius: 16px;
    background: #ffffff;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.03);
    color: inherit;
    font-family: inherit;
    text-align: left;
    cursor: pointer;
    text-decoration: none;
    transition: all 0.3s ease;
  }

  .kelas-card:hover {
    border-color: #e0e7ff;
    box-shadow: 0 10px 25px rgba(79, 70, 229, 0.08);
    transform: translateY(-2px);
  }

  .kelas-icon {
    width: 48px;
    height: 48px;
    border-radius: 14px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: #f0f9ff;
    color: #0284c7;
    font-weight: 800;
    font-size: 15px;
  }
  
  .kelas-icon.tingkat-VII, .kelas-icon.tingkat-7 {
    background: #e6f7ec;
    color: #0b7c3d;
  }
  
  .kelas-icon.tingkat-VIII, .kelas-icon.tingkat-8 {
    background: #eaf2fe;
    color: #2f80ed;
  }
  
  .kelas-icon.tingkat-IX, .kelas-icon.tingkat-9 {
    background: #fef2e6;
    color: #f2994a;
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
    .sticky-header {
      position: sticky;
      top: 75px; /* Offset to stick below Header.svelte */
      z-index: 40;
      background: transparent;
      padding: 0;
      margin: 0 0 1rem 0;
    }
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
      scrollbar-width: none; /* Firefox */
      -ms-overflow-style: none; /* IE/Edge */
      
      background: transparent;
      border: none;
      padding: 0;
      gap: 8px;
      margin-bottom: 1rem;
    }
    .card-tabs::-webkit-scrollbar {
      display: none; /* Safari/Chrome */
    }
    
    .tab-btn {
      padding: 9px 14px;
      border-radius: 12px;
      font-size: 12.5px;
      background: var(--color-surface, #fff);
      border: 1px solid var(--color-border, #e2e8f0);
      box-shadow: none;
    }
    
    .tab-btn.active {
      background: var(--color-text-main, #1e293b);
      border-color: var(--color-text-main, #1e293b);
      color: #fff;
      box-shadow: none;
    }
  }

  .filter-bar {
    display: flex;
    gap: 1rem;
    align-items: center;
    padding: 0.75rem 1rem;
    background: #f8fafc;
    border: 1px solid #e2e8f0;
    border-radius: 10px;
    margin-bottom: 1.5rem;
  }
  .filter-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }
  .filter-item label {
    font-size: 0.85rem;
    font-weight: 600;
    color: #64748b;
    white-space: nowrap;
  }
  .filter-item select {
    padding: 0.35rem 0.6rem;
    border: 1px solid #cbd5e1;
    border-radius: 6px;
    font-size: 0.9rem;
    font-weight: 600;
    background: white;
    color: #1e293b;
    min-width: 150px;
  }
</style>
