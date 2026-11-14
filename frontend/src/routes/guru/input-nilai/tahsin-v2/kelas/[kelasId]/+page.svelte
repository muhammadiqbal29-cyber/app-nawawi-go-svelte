<script>
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import Icon from '$lib/components/ui/Icon.svelte';

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

  let kelasId = $derived($page.params.kelasId);
  let kelasLabel = $derived($page.url.searchParams.get('kelas') || `Kelas #${kelasId}`);
  let rombel = $derived($page.url.searchParams.get('rombel') || '__kosong__');
  let rombelText = $derived(rombel === '__kosong__' ? 'Tanpa Rombel' : `Rombel ${rombel}`);
  let sekolahId = $derived($page.url.searchParams.get('sekolah_id') || '1');
  let jenisPenilaian = $derived($page.url.searchParams.get('jenis_penilaian') || 'semester');
  let tahunAjaranId = $derived($page.url.searchParams.get('tahun_ajaran_id'));
  let semesterId = $derived($page.url.searchParams.get('semester_id'));

  let siswaList = $state([]);
  let loading = $state(false);
  let error = $state('');
  let searchQuery = $state('');
  let statusFilter = $state('all'); // 'all' | 'graded'

  // Reactive Stats
  let totalSiswa = $derived(siswaList.length);
  let jumlahDinilai = $derived(siswaList.filter((s) => s.sudahDinilai).length);

  let filteredSiswa = $derived(
    siswaList.filter((s) => {
      const q = searchQuery.trim().toLowerCase();
      const matchSearch =
        !q ||
        (s.nama || '').toLowerCase().includes(q) ||
        (s.nisn && s.nisn.includes(q));
      if (!matchSearch) return false;

      if (statusFilter === 'graded') return s.sudahDinilai;
      return true;
    })
  );

  onMount(loadSiswa);

  async function loadSiswa() {
    loading = true;
    error = '';
    try {
      const params = new URLSearchParams({
        mata_pelajaran: 'Tahsin',
        jenis_penilaian: jenisPenilaian,
        rombel
      });
      if (tahunAjaranId) params.append('tahun_ajaran_id', tahunAjaranId);
      if (semesterId) params.append('semester_id', semesterId);

      // Fetch masal tahsin V2
      let fetchedList = [];
      const resMasal = await fetch(
        `${apiUrl}/api/guru/kelas/${kelasId}/tahsin-v2-masal?${params.toString()}`,
        { credentials: 'include' }
      );

      if (resMasal.ok) {
        const rawMasal = await resMasal.json();
        if (Array.isArray(rawMasal)) {
          fetchedList = rawMasal.map((item) => {
            const s = item.siswa || {};
            const hasNilai = Boolean(item.nilai && (item.nilai.id || (item.details && item.details.length > 0)));
            return {
              id: s.id || s.siswa_id,
              nama: s.nama_lengkap || s.nama || '-',
              nisn: s.nisn || '-',
              jilid: item.nilai?.jilid || 1,
              sudahDinilai: hasNilai,
              nilai: item.nilai || null
            };
          });
        }
      }

      // Fallback endpoint jika belum terisi
      if (fetchedList.length === 0) {
        const res = await fetch(
          `${apiUrl}/api/guru/kelas/${kelasId}/nilai?${params.toString()}`,
          { credentials: 'include' }
        );
        if (!res.ok) {
          const body = await res.json().catch(() => ({}));
          throw new Error(body.message || 'Gagal memuat data siswa.');
        }
        const body = await res.json();
        const nilaiMap = body.nilai_map || {};
        fetchedList = (body.siswa || []).map((s) => {
          const n = nilaiMap[`${s.siswa_id}_Tahsin`];
          return {
            id: s.siswa_id,
            nama: s.nama_lengkap || '-',
            nisn: s.nisn || '-',
            jilid: 1,
            sudahDinilai: Boolean(n),
            nilai: n || null
          };
        });
      }

      siswaList = fetchedList;
    } catch (err) {
      if (siswaList.length === 0) {
        error = err.message || 'Gagal memuat data siswa.';
      }
    } finally {
      loading = false;
    }
  }

  function getInitials(name) {
    if (!name) return '?';
    const words = name.trim().split(/\s+/);
    if (words.length === 1) return words[0].slice(0, 2).toUpperCase();
    return (words[0][0] + words[1][0]).toUpperCase();
  }
</script>

<div class="page-container">
  <!-- TOP BREADCRUMB & HEADER -->
  <div class="page-header-actions">
    <div>
      <a class="back-link" href="/guru/input-nilai/tahsin-v2">
        <Icon name="arrowLeft" size={16} /> Kembali ke Pilih Kelas
      </a>
      <h1 class="page-title">{kelasLabel}</h1>
      <p class="page-subtitle">{rombelText} · {totalSiswa} siswa di rombel ini · Ujian Tahsin (Versi 2)</p>
    </div>
  </div>

  {#if error}
    <div class="alert-error">{error}</div>
  {/if}

  <!-- TOOLBAR: FILTER & SEARCH -->
  <div class="toolbar-section">
    <div class="filter-pills">
      <button
        type="button"
        class="filter-pill"
        class:active={statusFilter === 'all'}
        onclick={() => (statusFilter = 'all')}
      >
        Semua <span class="pill-count">{totalSiswa}</span>
      </button>

      <button
        type="button"
        class="filter-pill pill-graded"
        class:active={statusFilter === 'graded'}
        onclick={() => (statusFilter = 'graded')}
      >
        <span class="pill-icon">✓</span> Sudah Dinilai <span class="pill-count">{jumlahDinilai}</span>
      </button>
    </div>

    <div class="search-box">
      <Icon name="search" size={16} class="search-icon" />
      <input
        type="text"
        placeholder="Cari nama siswa atau NISN..."
        class="form-input search-input"
        bind:value={searchQuery}
      />
      {#if searchQuery}
        <button
          type="button"
          class="search-clear-btn"
          onclick={() => (searchQuery = '')}
          aria-label="Hapus pencarian"
        >
          ×
        </button>
      {/if}
    </div>
  </div>

  <!-- STUDENT LIST -->
  {#if loading}
    <div class="loading-state">
      <div class="loading-spinner"></div>
      <span>Memuat data siswa kelas...</span>
    </div>
  {:else}
    <div class="student-list">
      {#each filteredSiswa as siswa (siswa.id)}
        <div class="student-row" class:row-graded={siswa.sudahDinilai}>
          <div class="student-info-group">
            <div class="avatar-circle" class:avatar-graded={siswa.sudahDinilai}>
              {getInitials(siswa.nama)}
            </div>
            <div class="student-text">
              <h3 class="student-nama">{siswa.nama}</h3>
              <p class="student-meta">
                {kelasLabel} · {rombelText} · NISN: {siswa.nisn}
                {#if siswa.sudahDinilai && siswa.nilai}
                  · <span class="jilid-meta">Iqro {siswa.nilai.jilid || 1}</span>
                {/if}
              </p>
            </div>
          </div>

          <div class="student-right-actions">
            {#if siswa.sudahDinilai}
              <span class="status-badge badge-graded">
                <Icon name="check" size={13} /> Sudah Dinilai (Iqro {siswa.nilai?.jilid || siswa.jilid || 1})
                {#if siswa.nilai?.nilai_akhir}
                  · Skor: {siswa.nilai.nilai_akhir}
                {/if}
              </span>
              <a
                class="btn-action btn-edit"
                href={`/guru/input-nilai-tahsin-v2/${siswa.id}?kelas=${encodeURIComponent(`${kelasLabel} - ${rombelText}`)}&nama=${encodeURIComponent(siswa.nama)}&sekolah_id=${sekolahId}&kelas_id=${kelasId}&rombel=${encodeURIComponent(rombel)}&kelas_label=${encodeURIComponent(kelasLabel)}&jenis_penilaian=${encodeURIComponent(jenisPenilaian)}${tahunAjaranId ? `&tahun_ajaran_id=${tahunAjaranId}` : ''}${semesterId ? `&semester_id=${semesterId}` : ''}`}
              >
                <Icon name="edit" size={14} /> Edit Nilai
              </a>
            {:else}
              <a
                class="btn-action btn-input"
                href={`/guru/input-nilai-tahsin-v2/${siswa.id}?kelas=${encodeURIComponent(`${kelasLabel} - ${rombelText}`)}&nama=${encodeURIComponent(siswa.nama)}&sekolah_id=${sekolahId}&kelas_id=${kelasId}&rombel=${encodeURIComponent(rombel)}&kelas_label=${encodeURIComponent(kelasLabel)}&jenis_penilaian=${encodeURIComponent(jenisPenilaian)}${tahunAjaranId ? `&tahun_ajaran_id=${tahunAjaranId}` : ''}${semesterId ? `&semester_id=${semesterId}` : ''}`}
              >
                <Icon name="plus" size={14} /> Input Nilai
              </a>
            {/if}
          </div>
        </div>
      {:else}
        <div class="empty-state">
          <Icon name="search" size={40} />
          <h4>Tidak Ada Siswa Ditemukan</h4>
          <p>
            {#if searchQuery}
              Tidak ada siswa yang cocok dengan "{searchQuery}".
            {:else if statusFilter !== 'all'}
              Tidak ada siswa yang sudah dinilai tahsin di kelas ini.
            {:else}
              Belum ada siswa yang terdaftar di kelas ini.
            {/if}
          </p>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .page-container {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    padding-bottom: 3rem;
  }

  .back-link {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    font-size: 0.85rem;
    font-weight: 600;
    color: #0b7c3d;
    text-decoration: none;
    margin-bottom: 0.5rem;
  }
  .back-link:hover {
    text-decoration: underline;
  }

  .page-title {
    margin: 0;
    font-size: 1.6rem;
    font-weight: 800;
    color: #0f172a;
  }

  .page-subtitle {
    margin: 0.25rem 0 0;
    font-size: 0.9rem;
    color: #64748b;
  }

  /* TOOLBAR */
  .toolbar-section {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .filter-pills {
    display: flex;
    gap: 0.5rem;
    background: #f1f5f9;
    padding: 0.25rem;
    border-radius: 10px;
    border: 1px solid #e2e8f0;
  }

  .filter-pill {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.4rem 0.85rem;
    border-radius: 8px;
    border: none;
    background: transparent;
    font-size: 0.85rem;
    font-weight: 600;
    color: #64748b;
    cursor: pointer;
    transition: all 0.2s;
  }

  .filter-pill.active {
    background: #ffffff;
    color: #0f172a;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
  }

  .pill-count {
    font-size: 0.75rem;
    background: #e2e8f0;
    padding: 0.1rem 0.45rem;
    border-radius: 9999px;
  }

  .filter-pill.active .pill-count {
    background: #f1f5f9;
    color: #0f172a;
  }

  .search-box {
    position: relative;
    display: flex;
    align-items: center;
    min-width: 280px;
  }

  :global(.search-icon) {
    position: absolute;
    left: 0.85rem;
    color: #94a3b8;
    pointer-events: none;
  }

  .search-input {
    width: 100%;
    padding: 0.55rem 2rem 0.55rem 2.4rem;
    border-radius: 10px;
    border: 1px solid #cbd5e1;
    font-size: 0.88rem;
    outline: none;
    transition: border-color 0.2s;
    background: #fff;
  }

  .search-input:focus {
    border-color: #0b7c3d;
  }

  .search-clear-btn {
    position: absolute;
    right: 0.6rem;
    background: transparent;
    border: none;
    color: #94a3b8;
    font-size: 1.2rem;
    cursor: pointer;
    padding: 0 0.3rem;
  }

  /* STUDENT LIST */
  .student-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .student-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 1.25rem;
    background: #ffffff;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    transition: all 0.2s ease;
  }

  .student-row:hover {
    border-color: #cbd5e1;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  }

  .student-row.row-graded {
    border-left: 4px solid #0b7c3d;
  }

  .student-info-group {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .avatar-circle {
    width: 44px;
    height: 44px;
    border-radius: 50%;
    background: #f1f5f9;
    color: #475569;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 800;
    font-size: 0.95rem;
    flex-shrink: 0;
  }

  .avatar-circle.avatar-graded {
    background: #e6f7ec;
    color: #0b7c3d;
  }

  .student-nama {
    margin: 0;
    font-size: 1rem;
    font-weight: 700;
    color: #0f172a;
  }

  .student-meta {
    margin: 0.2rem 0 0;
    font-size: 0.82rem;
    color: #64748b;
  }

  .jilid-meta {
    font-weight: 700;
    color: #0b7c3d;
  }

  .student-right-actions {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .status-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    font-size: 0.78rem;
    font-weight: 700;
    padding: 0.3rem 0.75rem;
    border-radius: 9999px;
    white-space: nowrap;
  }

  .badge-graded {
    background: #ecfdf5;
    color: #059669;
    border: 1px solid #a7f3d0;
  }

  .btn-action {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.5rem 1rem;
    border-radius: 8px;
    font-size: 0.85rem;
    font-weight: 700;
    text-decoration: none;
    transition: all 0.2s;
    cursor: pointer;
    white-space: nowrap;
  }

  .btn-input {
    background: #0b7c3d;
    color: #ffffff;
    border: 1px solid #0b7c3d;
  }

  .btn-input:hover {
    background: #096833;
  }

  .btn-edit {
    background: #f8fafc;
    color: #334155;
    border: 1px solid #cbd5e1;
  }

  .btn-edit:hover {
    background: #f1f5f9;
    color: #0f172a;
  }

  .empty-state {
    text-align: center;
    padding: 3rem 1rem;
    color: #94a3b8;
  }
  .empty-state h4 {
    margin: 0.75rem 0 0.25rem;
    color: #334155;
  }
  .empty-state p {
    margin: 0;
    font-size: 0.88rem;
  }

  .loading-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem;
    gap: 0.75rem;
    color: #64748b;
  }
  .loading-spinner {
    width: 32px;
    height: 32px;
    border: 3px solid #e2e8f0;
    border-top-color: #0b7c3d;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }
  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  @media (max-width: 720px) {
    .page-header-actions {
      flex-direction: column;
      align-items: stretch;
    }

    .toolbar-section {
      flex-direction: column;
      align-items: stretch;
      gap: 0.75rem;
    }

    .filter-pills {
      width: 100%;
      display: flex;
    }

    .filter-pill {
      flex: 1;
      justify-content: center;
    }

    .search-box {
      width: 100%;
      min-width: 0;
    }

    .student-row {
      flex-direction: column;
      align-items: stretch;
      gap: 0.85rem;
      padding: 1rem;
    }

    .student-right-actions {
      width: 100%;
      justify-content: space-between;
      border-top: 1px solid #f1f5f9;
      padding-top: 0.75rem;
      flex-wrap: wrap;
      gap: 0.5rem;
    }

    .btn-action {
      flex: 1;
      justify-content: center;
    }
  }
</style>
