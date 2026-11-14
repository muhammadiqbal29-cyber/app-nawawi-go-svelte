<script>
  import Icon from '$lib/components/ui/Icon.svelte';
  import { guruState } from '$lib/guruState.svelte.js';

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

  let kelasList = $state([]);
  let selectedKelas = $state(null);
  let siswaList = $state([]);
  let tahsinMap = $state({}); // siswa_id -> { jilid, nilai_akhir, predikat }
  let searchQuery = $state('');
  let filterOnlyScored = $state(false);

  let filteredSiswa = $derived(
    siswaList
      .filter((s) => (s.nama_lengkap || '').toLowerCase().includes(searchQuery.toLowerCase()))
      .filter((s) => (filterOnlyScored ? !!tahsinMap[s.id] : true))
  );

  let loadingKelas = $state(false);
  let loadingSiswa = $state(false);
  let error = $state('');

  $effect(() => {
    const sekolahId = guruState.activeSekolahId;
    const taId = guruState.activeTahunId;
    const semId = guruState.activeSemesterId;
    if (sekolahId && taId && semId) {
      loadKelas().then(() => {
        if (typeof window !== 'undefined') {
          const saved = sessionStorage.getItem('rapor_tahsin_kelas');
          if (saved) {
            try {
              const k = JSON.parse(saved);
              if (k) pilihKelas(k);
            } catch (e) {
              console.error('Gagal restore kelas:', e);
            }
          }
        }
      });
    }
  });

  async function loadKelas() {
    if (!guruState.activeSekolahId) return;
    loadingKelas = true;
    try {
      let url = `${apiUrl}/api/guru/sekolah/${guruState.activeSekolahId}/kelas`;
      if (guruState.activeTahunId) url += `?tahun_ajaran_id=${guruState.activeTahunId}`;
      if (guruState.activeSemesterId) url += `&semester_id=${guruState.activeSemesterId}`;
      const res = await fetch(url, { credentials: 'include' });
      if (!res.ok) throw new Error('Gagal memuat data kelas.');
      kelasList = await res.json();
    } catch (err) {
      error = err.message;
    } finally {
      loadingKelas = false;
    }
  }

  async function pilihKelas(kelas) {
    selectedKelas = kelas;
    if (typeof window !== 'undefined') {
      sessionStorage.setItem('rapor_tahsin_kelas', JSON.stringify(kelas));
    }
    loadingSiswa = true;
    try {
      const params = new URLSearchParams({ rombel: kelas.rombel || '__kosong__' });
      if (guruState.activeTahunId) params.set('tahun_ajaran_id', guruState.activeTahunId);
      if (guruState.activeSemesterId) params.set('semester_id', guruState.activeSemesterId);

      // Ambil daftar siswa kelas
      const res = await fetch(`${apiUrl}/api/guru/kelas/${kelas.id}/siswa?${params.toString()}`, {
        credentials: 'include'
      });
      if (!res.ok) throw new Error('Gagal memuat daftar siswa.');
      const data = await res.json();
      siswaList = data.data || data || [];

      // Coba ambil rekap tahsin masal untuk info status nilai
      const paramsTahsin = new URLSearchParams({
        mata_pelajaran: 'Tahsin',
        jenis_penilaian: 'semester'
      });
      if (guruState.activeSekolahId) paramsTahsin.append('sekolah_id', guruState.activeSekolahId);
      if (kelas.rombel && kelas.rombel !== '__kosong__') paramsTahsin.append('rombel', kelas.rombel);
      if (guruState.activeTahunId) paramsTahsin.append('tahun_ajaran_id', guruState.activeTahunId);
      if (guruState.activeSemesterId) paramsTahsin.append('semester_id', guruState.activeSemesterId);

      let resTahsin = await fetch(`${apiUrl}/api/guru/kelas/${kelas.id}/tahsin-v2-masal?${paramsTahsin.toString()}`, {
        credentials: 'include'
      });
      if (!resTahsin.ok) {
        resTahsin = await fetch(`${apiUrl}/api/guru/kelas/${kelas.id}/tahsin-masal?${paramsTahsin.toString()}`, {
          credentials: 'include'
        });
      }

      if (resTahsin.ok) {
        const rawMasal = await resTahsin.json();
        const map = {};
        (rawMasal || []).forEach((item) => {
          if (item.nilai && item.nilai.jilid && item.details && item.details.length > 0) {
            const jilid = item.nilai.jilid;
            const currentInd = (item.indikator_list || []).filter((ind) => ind.jilid === jilid);
            const totalInd = currentInd.length;
            const detailMap = {};
            (item.details || []).forEach((d) => {
              detailMap[d.kode] = d.nilai;
            });
            const totalPoin = currentInd.reduce((sum, ind) => {
              const v = detailMap[ind.kode];
              return sum + (v !== undefined && v !== null ? Number(v) : 0);
            }, 0);

            const avg = totalInd > 0 ? Number((totalPoin / totalInd).toFixed(2)) : (item.nilai.nilai_akhir || 0);

            map[item.siswa?.id] = {
              jilid,
              nilai_akhir: avg,
              predikat: item.nilai.predikat
            };
          }
        });
        tahsinMap = map;
      }
    } catch (err) {
      error = err.message;
    } finally {
      loadingSiswa = false;
    }
  }

  function resetKelas() {
    selectedKelas = null;
    siswaList = [];
    tahsinMap = {};
    if (typeof window !== 'undefined') {
      sessionStorage.removeItem('rapor_tahsin_kelas');
    }
  }

  function kelasLabel(kelas) {
    return `${kelas.tingkat || ''} - ${kelas.nama_kelas || ''}`.trim();
  }
  function rombelLabel(rombel) {
    return !rombel || rombel === '__kosong__' ? 'Tanpa Rombel' : `Rombel ${rombel}`;
  }
</script>

<svelte:head>
  <title>Rapor Ujian Tahsin - SIAKAD</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div>
      <p class="eyebrow">Rapor Al-Qur'an</p>
      <h1>Rapor Ujian Tahsin</h1>
      <p class="subtitle">
        Pilih kelas lalu lihat atau cetak lembar rapor hasil ujian tahsin bacaan Iqro siswa.
      </p>
    </div>
  </div>

  {#if error}
    <div class="alert-error">{error}</div>
  {/if}

  {#if !guruState.activeSekolahId}
    <div class="empty-state">
      <Icon name="grid" size={48} />
      <h3>Pilih Jenjang Sekolah</h3>
      <p>Silakan klik tombol Pilih Jenjang terlebih dahulu di bilah atas.</p>
    </div>
  {:else if !selectedKelas}
    <!-- Pilih Kelas -->
    {#if loadingKelas}
      <div class="loading-box">Memuat data kelas...</div>
    {:else if kelasList.length === 0}
      <div class="empty-state">
        <Icon name="info" size={48} />
        <h3>Belum ada kelas untuk tahun ajaran & semester ini</h3>
      </div>
    {:else}
      <div class="kelas-grid">
        {#each kelasList as kelas (kelas.id + kelas.rombel)}
          <button class="kelas-card" onclick={() => pilihKelas(kelas)}>
            <div class="kelas-icon"><Icon name="book" size={20} /></div>
            <div class="kelas-card-text">
              <span class="kelas-name">{kelasLabel(kelas)}</span>
              <span class="kelas-rombel">{rombelLabel(kelas.rombel)}</span>
              <span class="kelas-count">{kelas.jumlah_siswa || 0} siswa</span>
            </div>
            <Icon name="chevronRight" size={18} />
          </button>
        {/each}
      </div>
    {/if}
  {:else}
    <!-- Detail Kelas & Pilih Siswa -->
    <div class="selected-header">
      <div class="header-left">
        <button class="btn-back" onclick={resetKelas}>
          <Icon name="chevronLeft" size={16} /> Ganti Kelas
        </button>
        <div class="selected-kelas">
          <Icon name="book" size={16} />
          <strong>{kelasLabel(selectedKelas)}</strong>
          <span class="badge">{rombelLabel(selectedKelas.rombel)}</span>
        </div>
      </div>

      <div class="header-actions">
        <a
          href={`/guru/rapor/tahsin/kelas/${selectedKelas.id}/catatan-masal?rombel=${selectedKelas.rombel || '__kosong__'}&sekolah_id=${guruState.activeSekolahId}&tahun_ajaran_id=${guruState.activeTahunId}&semester_id=${guruState.activeSemesterId}&kelas=${encodeURIComponent(kelasLabel(selectedKelas))}`}
          class="btn-catatan-masal"
        >
          <Icon name="edit" size={16} /> Isi Catatan Masal
        </a>
        <a
          href={`/guru/rapor/tahsin/kelas/${selectedKelas.id}/masal?rombel=${selectedKelas.rombel || '__kosong__'}&sekolah_id=${guruState.activeSekolahId}&tahun_ajaran_id=${guruState.activeTahunId}&semester_id=${guruState.activeSemesterId}`}
          class="btn-print-masal"
        >
          <Icon name="printer" size={16} /> Cetak Rapor Masal
        </a>
      </div>
    </div>

    <div class="filter-bar">
      <div class="search-bar">
        <Icon name="search" size={16} class="search-icon" />
        <input
          type="text"
          bind:value={searchQuery}
          placeholder="Cari nama siswa..."
          class="search-input"
        />
      </div>

      <label class="toggle-scored">
        <input type="checkbox" bind:checked={filterOnlyScored} />
        <span>Hanya tampilkan siswa yang telah dinilai</span>
      </label>
    </div>

    {#if loadingSiswa}
      <div class="loading-box">Memuat siswa...</div>
    {:else if siswaList.length === 0}
      <div class="empty-state">
        <Icon name="users" size={48} />
        <h3>Belum ada siswa di rombel ini</h3>
      </div>
    {:else}
      <div class="siswa-grid">
        {#each filteredSiswa as siswa (siswa.id)}
          {@const skorTahsin = tahsinMap[siswa.id]}
          <a
            class="siswa-card"
            href={`/guru/rapor/tahsin/${siswa.id}?sekolah_id=${guruState.activeSekolahId}&tahun_ajaran_id=${guruState.activeTahunId}&semester_id=${guruState.activeSemesterId}`}
          >
            <div class="siswa-info">
              <span class="siswa-nama">{siswa.nama_lengkap}</span>
              <span class="siswa-nisn">NISN: {siswa.nisn || '-'}</span>
              {#if skorTahsin}
                <div class="tahsin-status">
                  <span class="badge-jilid">Iqro {skorTahsin.jilid}</span>
                  <span class="badge-score">Nilai: {skorTahsin.nilai_akhir}</span>
                </div>
              {:else}
                <div class="tahsin-status">
                  <span class="badge-unscored">Belum Ujian</span>
                </div>
              {/if}
            </div>
            <div class="card-arrow">
              <Icon name="chevronRight" size={18} />
            </div>
          </a>
        {/each}
      </div>
    {/if}
  {/if}
</div>

<style>
  .page-container {
    max-width: 1100px;
    margin: 0 auto;
    padding: 1rem 0;
  }

  .page-header {
    margin-bottom: 2rem;
  }

  .eyebrow {
    margin: 0 0 0.5rem;
    color: #0b7c3d;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.12em;
    font-size: 0.75rem;
  }

  .page-header h1 {
    margin: 0;
    font-size: 1.85rem;
    color: #0f172a;
    font-weight: 800;
  }

  .subtitle {
    margin: 0.5rem 0 0;
    color: #64748b;
    font-size: 0.95rem;
  }

  .alert-error {
    background: #fef2f2;
    border: 1px solid #fecaca;
    color: #dc2626;
    padding: 0.75rem 1rem;
    border-radius: 8px;
    margin-bottom: 1.5rem;
  }

  .kelas-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 1rem;
  }

  .kelas-card {
    display: flex;
    align-items: center;
    gap: 1rem;
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    padding: 1.25rem;
    cursor: pointer;
    text-align: left;
    transition: all 0.2s ease;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  }

  .kelas-card:hover {
    border-color: #0b7c3d;
    transform: translateY(-2px);
    box-shadow: 0 6px 12px rgba(11, 124, 61, 0.1);
  }

  .kelas-icon {
    width: 44px;
    height: 44px;
    background: #ecfdf5;
    color: #0b7c3d;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .kelas-card-text {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
  }

  .kelas-name {
    font-weight: 700;
    color: #0f172a;
    font-size: 1rem;
  }

  .kelas-rombel {
    font-size: 0.825rem;
    color: #64748b;
  }

  .kelas-count {
    font-size: 0.75rem;
    color: #0b7c3d;
    font-weight: 600;
  }

  .selected-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 1rem;
    margin-bottom: 1.5rem;
    background: white;
    padding: 1rem 1.25rem;
    border-radius: 12px;
    border: 1px solid #e2e8f0;
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .btn-back {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    background: #f1f5f9;
    color: #475569;
    border: 1px solid #cbd5e1;
    border-radius: 8px;
    padding: 0.45rem 0.85rem;
    font-size: 0.85rem;
    font-weight: 600;
    cursor: pointer;
  }

  .btn-back:hover {
    background: #e2e8f0;
    color: #0f172a;
  }

  .selected-kelas {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.95rem;
    color: #0f172a;
  }

  .badge {
    background: #ecfdf5;
    color: #0b7c3d;
    padding: 0.15rem 0.55rem;
    border-radius: 999px;
    font-size: 0.75rem;
    font-weight: 600;
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .btn-catatan-masal {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    background: #f8fafc;
    color: #4f46e5;
    border: 1px solid #c7d2fe;
    border-radius: 8px;
    padding: 0.5rem 1rem;
    font-size: 0.85rem;
    font-weight: 600;
    text-decoration: none;
    transition: all 0.15s;
  }

  .btn-catatan-masal:hover {
    background: #eef2ff;
    border-color: #818cf8;
  }

  .btn-print-masal {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    background: #0b7c3d;
    color: white;
    border: none;
    border-radius: 8px;
    padding: 0.5rem 1.15rem;
    font-size: 0.85rem;
    font-weight: 700;
    text-decoration: none;
    box-shadow: 0 2px 4px rgba(11, 124, 61, 0.2);
    transition: all 0.15s;
  }

  .btn-print-masal:hover {
    background: #096330;
  }

  .filter-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .search-bar {
    position: relative;
    display: flex;
    align-items: center;
    width: 280px;
  }

  :global(.search-icon) {
    position: absolute;
    left: 0.75rem;
    color: #94a3b8;
  }

  .search-input {
    width: 100%;
    padding: 0.5rem 0.75rem 0.5rem 2.25rem;
    border: 1px solid #cbd5e1;
    border-radius: 8px;
    font-size: 0.875rem;
    outline: none;
  }

  .search-input:focus {
    border-color: #0b7c3d;
    box-shadow: 0 0 0 2px rgba(11, 124, 61, 0.15);
  }

  .toggle-scored {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    font-size: 0.85rem;
    font-weight: 600;
    color: #475569;
    cursor: pointer;
  }

  .siswa-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 1rem;
  }

  .siswa-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    padding: 1.15rem 1.25rem;
    text-decoration: none;
    color: inherit;
    transition: all 0.15s ease;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  }

  .siswa-card:hover {
    border-color: #0b7c3d;
    transform: translateY(-2px);
    box-shadow: 0 6px 12px rgba(11, 124, 61, 0.1);
  }

  .siswa-info {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .siswa-nama {
    font-weight: 700;
    color: #0f172a;
    font-size: 0.95rem;
  }

  .siswa-nisn {
    font-size: 0.75rem;
    color: #64748b;
  }

  .tahsin-status {
    display: flex;
    gap: 0.4rem;
    align-items: center;
    margin-top: 0.35rem;
  }

  .badge-jilid {
    font-size: 0.7rem;
    font-weight: 700;
    background: #dbeafe;
    color: #1e40af;
    padding: 0.1rem 0.5rem;
    border-radius: 999px;
  }

  .badge-score {
    font-size: 0.7rem;
    font-weight: 700;
    background: #ecfdf5;
    color: #065f46;
    padding: 0.1rem 0.5rem;
    border-radius: 999px;
  }

  .badge-unscored {
    font-size: 0.7rem;
    font-weight: 600;
    background: #f1f5f9;
    color: #64748b;
    padding: 0.1rem 0.5rem;
    border-radius: 999px;
  }

  .card-arrow {
    color: #94a3b8;
    transition: transform 0.15s;
  }

  .siswa-card:hover .card-arrow {
    transform: translateX(3px);
    color: #0b7c3d;
  }

  .loading-box,
  .empty-state {
    text-align: center;
    padding: 3rem 1rem;
    color: #64748b;
    background: white;
    border-radius: 12px;
    border: 1px dashed #cbd5e1;
  }
</style>
