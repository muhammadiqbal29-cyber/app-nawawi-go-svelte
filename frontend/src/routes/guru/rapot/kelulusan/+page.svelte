<script>
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import Icon from '$lib/components/ui/Icon.svelte';

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';
  import { guruState } from '$lib/guruState.svelte.js';

  const mataPelajaran = 'Tahfiz';
  const jenisPenilaian = 'kelulusan';

  let hasAccess = $derived(
    guruState.activeSekolahId &&
      guruState.aksesMap[guruState.activeSekolahId] &&
      guruState.aksesMap[guruState.activeSekolahId].includes('cetak_rapor')
  );
  let kelasList = $state([]);
  let selectedKelasId = $state('');
  let siswaList = $state([]);
  let loadingKelas = $state(false);
  let loadingSiswa = $state(false);
  let errorKelas = $state('');
  let errorSiswa = $state('');
  let searchQuery = $state('');
  let querySelectedKelasId = '';

  onMount(async () => {
    querySelectedKelasId = $page.url.searchParams.get('selectedKelasId') || '';
    if (guruState.activeSekolahId) {
      await loadKelas();
      let restoredId = querySelectedKelasId;
      if (!restoredId && typeof window !== 'undefined') {
        restoredId = sessionStorage.getItem('rapot_kelulusan_kelas_id') || '';
      }
      if (restoredId) {
        selectedKelasId = restoredId;
        await loadSiswa();
      }
    }
  });

  $effect(() => {
    if (guruState.activeSekolahId && hasAccess) {
      selectedKelasId = '';
      if (typeof window !== 'undefined') {
        sessionStorage.removeItem('rapot_kelulusan_kelas_id');
      }
      siswaList = [];
      loadKelas();
    }
  });

  async function loadKelas() {
    if (!guruState.activeSekolahId || !hasAccess) return;
    loadingKelas = true;
    errorKelas = '';
    kelasList = [];
    siswaList = [];

    try {
      const params = new URLSearchParams({ jenis_penilaian: jenisPenilaian });
      const res = await fetch(
        `${apiUrl}/api/guru/sekolah/${guruState.activeSekolahId}/kelas?${params.toString()}`,
        {
          credentials: 'include'
        }
      );
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.message || 'Gagal memuat daftar kelas.');
      }
      const data = await res.json();
      kelasList = Array.isArray(data) ? data : [];
    } catch (err) {
      errorKelas = err.message || 'Gagal memuat daftar kelas.';
    } finally {
      loadingKelas = false;
    }
  }

  function kelasKey(kelas) {
    return `${kelas.id}::${kelas.rombel || '__kosong__'}`;
  }

  function parseKelasKey(key) {
    const [id, rombel] = key.split('::');
    return { id, rombel };
  }

  function findSelectedKelas() {
    if (!selectedKelasId) return null;
    const { id, rombel } = parseKelasKey(selectedKelasId);
    return (
      kelasList.find((k) => String(k.id) === id && (k.rombel || '__kosong__') === rombel) || null
    );
  }

  function studentRapotLink(siswa) {
    const params = new URLSearchParams({
      selectedKelasId,
      nama: siswa.nama,
      nisn: siswa.nisn,
      kelas: siswa.kelasLabel
    });
    return `/guru/rapot/kelulusan/${siswa.id}?${params.toString()}`;
  }

  async function loadSiswa() {
    if (!selectedKelasId) {
      siswaList = [];
      if (typeof window !== 'undefined') {
        sessionStorage.removeItem('rapot_kelulusan_kelas_id');
      }
      return;
    }

    if (typeof window !== 'undefined') {
      sessionStorage.setItem('rapot_kelulusan_kelas_id', selectedKelasId);
    }

    loadingSiswa = true;
    errorSiswa = '';
    siswaList = [];

    try {
      const selectedKelas = findSelectedKelas();
      const { id: kelasIdNum, rombel: selectedRombel } = parseKelasKey(selectedKelasId);
      const rombel = selectedKelas?.rombel || selectedRombel || '__kosong__';
      const params = new URLSearchParams({
        mata_pelajaran: mataPelajaran,
        jenis_penilaian: jenisPenilaian,
        rombel
      });
      const res = await fetch(`${apiUrl}/api/guru/kelas/${kelasIdNum}/nilai?${params.toString()}`, {
        credentials: 'include'
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.message || 'Gagal memuat daftar siswa.');
      }
      const body = await res.json();
      siswaList = (body.siswa || []).map((s) => ({
        id: s.siswa_id,
        nama: s.nama_lengkap || s.nama || '-',
        nisn: s.nisn || '-',
        kelasLabel: selectedKelas
          ? `${selectedKelas.tingkat || ''} - ${selectedKelas.nama_kelas || ''}`.trim()
          : ''
      }));
    } catch (err) {
      errorSiswa = err.message || 'Gagal memuat daftar siswa.';
    } finally {
      loadingSiswa = false;
    }
  }

  function kelasLabel(kelas) {
    return kelas ? `${kelas.tingkat || ''} - ${kelas.nama_kelas || ''}`.trim() : '';
  }

  function rombelLabel(kelas) {
    if (!kelas) return 'Tanpa Rombel';
    return kelas.rombel && kelas.rombel !== '__kosong__'
      ? `Rombel ${kelas.rombel}`
      : 'Tanpa Rombel';
  }

  let filteredSiswa = $derived(
    searchQuery
      ? siswaList.filter((s) => s.nama.toLowerCase().includes(searchQuery.toLowerCase()))
      : siswaList
  );
</script>

<div class="rapot-detail">
  <div class="detail-header">
    <div class="card-icon">
      <Icon name="award" size={32} />
    </div>
    <div>
      <h1>Rapot Ujian Kelulusan</h1>
      <p>Daftar siswa berdasarkan kelas kelulusan yang Anda ampu.</p>
    </div>
  </div>

  {#if !guruState.activeSekolahId}
    <div class="empty-state-message">
      <Icon name="grid" size={44} />
      <h3>Pilih Jenjang Sekolah</h3>
      <p>Silakan pilih jenjang sekolah pada menu di pojok kanan atas.</p>
    </div>
  {:else if !hasAccess}
    <div class="empty-state-message">
      <Icon name="alertCircle" size={44} />
      <h3>Akses Ditolak</h3>
      <p>
        Anda tidak memiliki hak akses <b>Cetak Rapor</b> untuk jenjang {guruState.activeSekolahName}.
      </p>
    </div>
  {:else}
    <div class="detail-content">
      <div class="form-row">
        <label for="kelas-select">Pilih Kelas</label>
        <select
          id="kelas-select"
          class="form-select"
          bind:value={selectedKelasId}
          onchange={loadSiswa}
          disabled={loadingKelas || !kelasList || kelasList.length === 0}
        >
          {#if loadingKelas}
            <option>Memuat kelas...</option>
          {:else if kelasList.length === 0}
            <option>Tidak ada kelas tersedia</option>
          {:else}
            <option value="" disabled selected={selectedKelasId === ''}>Pilih Kelas</option>
            {#each kelasList as kelas (kelasKey(kelas))}
              <option value={kelasKey(kelas)}>{kelasLabel(kelas)} — {rombelLabel(kelas)}</option>
            {/each}
          {/if}
        </select>
      </div>

      {#if errorKelas}
        <div class="alert-error">{errorKelas}</div>
      {/if}

      <div class="student-panel">
        <div class="student-panel-header">
          <div>
            <h2>Daftar Siswa</h2>
            {#if selectedKelasId}
              <p>
                {kelasLabel(findSelectedKelas())} · {rombelLabel(findSelectedKelas())}
              </p>
            {/if}
          </div>
          <input
            type="text"
            placeholder="Cari siswa..."
            class="form-input search-input"
            bind:value={searchQuery}
          />
        </div>

        {#if errorSiswa}
          <div class="alert-error">{errorSiswa}</div>
        {/if}

        {#if loadingSiswa}
          <div class="empty-state-message">Memuat siswa...</div>
        {:else if filteredSiswa.length > 0}
          <div class="student-list">
            {#each filteredSiswa as siswa (siswa.id)}
              <a class="student-row student-row-link" href={studentRapotLink(siswa)}>
                <div class="student-info-group">
                  <div class="avatar-circle">{siswa.nama.charAt(0)}</div>
                  <div class="student-text">
                    <h3>{siswa.nama}</h3>
                    <p>NISN {siswa.nisn}</p>
                  </div>
                </div>
                <span class="student-badge">{siswa.kelasLabel}</span>
              </a>
            {/each}
          </div>
        {:else}
          <div class="empty-state-message">
            <Icon name="users" size={44} />
            <h3>Tidak ada siswa ditemukan</h3>
            <p>Silakan pilih kelas lain atau cek kembali data siswa.</p>
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
  .rapot-detail {
    max-width: 1000px;
    width: 100%;
    background: #fff;
    border: 1px solid #e2e8f0;
    border-radius: 1rem;
    padding: 2rem;
    box-shadow: 0 16px 40px rgba(15, 23, 42, 0.06);
  }

  .detail-header {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 1.75rem;
  }

  .card-icon {
    width: 56px;
    height: 56px;
    border-radius: 18px;
    display: flex;
    justify-content: center;
    align-items: center;
    background: rgba(16, 185, 129, 0.12);
    color: #047857;
  }

  .detail-header h1 {
    margin: 0;
    font-size: 1.85rem;
    color: #0f172a;
  }

  .detail-header p {
    margin: 0.65rem 0 0;
    color: #475569;
    max-width: 720px;
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr;
    gap: 0.75rem;
    margin-bottom: 1.5rem;
  }

  label {
    font-size: 0.85rem;
    font-weight: 700;
    color: #334155;
  }

  .form-select,
  .form-input {
    width: 100%;
    border-radius: 0.85rem;
    border: 1px solid #cbd5e1;
    background: #f8fafc;
    padding: 0.9rem 1rem;
    color: #0f172a;
    font-size: 0.95rem;
    outline: none;
    transition:
      border-color 0.2s ease,
      box-shadow 0.2s ease;
    font-family: inherit;
  }

  .form-select:focus,
  .form-input:focus {
    border-color: #6366f1;
    box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.08);
  }

  .student-panel {
    margin-top: 1rem;
  }

  .student-panel-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 1rem;
    margin-bottom: 1.25rem;
    flex-wrap: wrap;
  }

  .student-panel-header h2 {
    margin: 0;
    font-size: 1.25rem;
    color: #0f172a;
  }

  .student-panel-header p {
    margin: 0.4rem 0 0;
    color: #64748b;
    font-size: 0.95rem;
  }

  .search-input {
    max-width: 320px;
  }

  .student-list {
    display: grid;
    gap: 0.85rem;
  }

  .student-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    background: #f8fafc;
    border: 1px solid #e2e8f0;
    border-radius: 1rem;
    padding: 1rem 1.2rem;
  }

  .student-info-group {
    display: flex;
    align-items: center;
    gap: 1rem;
    min-width: 0;
  }

  .avatar-circle {
    width: 44px;
    height: 44px;
    border-radius: 50%;
    background: linear-gradient(135deg, #dbeafe, #bfdbfe);
    color: #1d4ed8;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.1rem;
    font-weight: 800;
    border: 1px solid #93c5fd;
    flex-shrink: 0;
  }

  .student-text h3 {
    margin: 0;
    font-size: 1rem;
    font-weight: 700;
    color: #1e293b;
  }

  .student-text p {
    margin: 0.2rem 0 0;
    color: #475569;
    font-size: 0.9rem;
  }

  .student-badge {
    background: #eef2ff;
    color: #4338ca;
    padding: 0.55rem 0.9rem;
    border-radius: 999px;
    font-size: 0.8rem;
    white-space: nowrap;
  }

  .alert-error {
    margin: 0.75rem 0;
    padding: 1rem 1.2rem;
    border-radius: 1rem;
    color: #b91c1c;
    background: #fee2e2;
    border: 1px solid #fecaca;
  }

  .empty-state-message {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.75rem;
    padding: 2rem;
    border: 1px dashed #cbd5e1;
    border-radius: 1rem;
    color: #475569;
    text-align: center;
  }

  .empty-state-message h3 {
    margin: 0;
    font-size: 1.05rem;
    color: #0f172a;
  }

  .empty-state-message p {
    margin: 0;
    font-size: 0.95rem;
  }

  @media (max-width: 760px) {
    .detail-header {
      flex-direction: column;
      align-items: flex-start;
    }

    .student-panel-header {
      flex-direction: column;
      align-items: stretch;
    }

    .student-row {
      flex-direction: column;
      align-items: stretch;
    }

    .student-badge {
      align-self: flex-start;
    }
  }
</style>
