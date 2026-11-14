<script>
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import Icon from '$lib/components/ui/Icon.svelte';
  import { SvelteMap } from 'svelte/reactivity';

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

  let siswaId = $derived($page.params.siswaId);
  let namaSiswa = $derived($page.url.searchParams.get('nama') || `Siswa #${siswaId}`);
  let kelas = $derived($page.url.searchParams.get('kelas') || '-');
  let sekolahId = $derived($page.url.searchParams.get('sekolah_id') || '1');
  let kelasId = $derived($page.url.searchParams.get('kelas_id') || '');
  let rombel = $derived($page.url.searchParams.get('rombel') || '__kosong__');
  let kelasLabel = $derived($page.url.searchParams.get('kelas_label') || '');
  let jenisPenilaian = $derived($page.url.searchParams.get('jenis_penilaian') || 'semester');
  let tahunAjaranId = $derived($page.url.searchParams.get('tahun_ajaran_id'));
  let semesterId = $derived($page.url.searchParams.get('semester_id'));

  let backUrl = $derived(
    kelasId
      ? `/guru/input-nilai/tahsin/kelas/${kelasId}?kelas=${encodeURIComponent(kelasLabel)}&rombel=${encodeURIComponent(rombel)}&sekolah_id=${sekolahId}&jenis_penilaian=${encodeURIComponent(jenisPenilaian)}${tahunAjaranId ? `&tahun_ajaran_id=${tahunAjaranId}` : ''}${semesterId ? `&semester_id=${semesterId}` : ''}`
      : '/guru/input-nilai/tahsin'
  );

  // State
  let loading = $state(true);
  let error = $state('');
  let globalSaveStatus = $state('Tersimpan');
  let selectedJilid = $state(1); // Iqro yang diujikan (1-6)

  let jilidList = $state([]);
  let allIndikators = $state([]);
  let savedDetails = $state([]);
  let headerData = $state(null);

  const saveTimers = new SvelteMap();
  const rowStatus = new SvelteMap();

  let currentJilidInfo = $derived(
    jilidList.find((j) => j.jilid === selectedJilid) || {
      jilid: selectedJilid,
      nama_jilid: `Jilid ${selectedJilid}`,
      fokus: '-',
      kriteria_kelulusan: '-'
    }
  );

  let currentIndikators = $derived(
    allIndikators
      .filter((ind) => ind.jilid === selectedJilid)
      .sort((a, b) => (a.urutan || 0) - (b.urutan || 0))
  );

  let scoreMap = $state({});

  // Kunci jilid jika sudah ada indikator yang dinilai
  let hasAnyScore = $derived(
    Object.values(scoreMap).some((val) => val !== undefined && val !== null && val !== '')
  );
  let isLocked = $derived(hasAnyScore);

  // Total indikator yang ada di jilidnya
  let totalIndikatorJilid = $derived(currentIndikators.length);

  // Summary (V1 scale 1-5)
  let totalDiisi = $derived(
    currentIndikators.filter((ind) => scoreMap[ind.kode] !== undefined && scoreMap[ind.kode] !== null && scoreMap[ind.kode] !== '').length
  );
  let totalPoin = $derived(
    currentIndikators.reduce((sum, ind) => {
      const val = scoreMap[ind.kode];
      return sum + (val !== undefined && val !== null && val !== '' ? Number(val) : 0);
    }, 0)
  );
  // Pembagian nilai akhir adalah rata-rata dari total indikator yang ada di jilidnya
  let rataRata = $derived(
    totalDiisi > 0 && totalIndikatorJilid > 0
      ? Number((totalPoin / totalIndikatorJilid).toFixed(2))
      : 0
  );
  let predikat = $derived(
    totalDiisi === 0
      ? '-'
      : rataRata >= 4.5
        ? 'A (Sangat Baik)'
        : rataRata >= 3.5
          ? 'B (Baik)'
          : rataRata >= 2.5
            ? 'C (Cukup)'
            : rataRata >= 1.5
              ? 'D (Kurang)'
              : 'E (Perlu Bimbingan)'
  );

  onMount(loadData);

  async function loadData() {
    loading = true;
    error = '';
    try {
      const params = new URLSearchParams({
        jenis_penilaian: jenisPenilaian,
        mata_pelajaran: 'Tahsin'
      });
      if (tahunAjaranId) params.append('tahun_ajaran_id', tahunAjaranId);
      if (semesterId) params.append('semester_id', semesterId);

      const res = await fetch(
        `${apiUrl}/api/guru/siswa/${siswaId}/nilai-tahsin?${params.toString()}`,
        { credentials: 'include' }
      );

      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.message || 'Gagal memuat lembar penilaian tahsin');
      }

      const data = await res.json();
      jilidList = data.jilid_list || [];
      allIndikators = data.indikator_list || [];
      headerData = data.nilai || null;
      savedDetails = data.details || [];

      if (headerData) {
        selectedJilid = headerData.jilid || 1;
      }

      const map = {};
      for (const d of savedDetails) {
        if (d.nilai !== null && d.nilai !== undefined) {
          map[d.kode] = d.nilai;
        }
      }
      scoreMap = map;
    } catch (err) {
      error = err.message || 'Terjadi kesalahan saat memuat data';
    } finally {
      loading = false;
    }
  }

  async function selectJilid(j) {
    if (selectedJilid === j) return;
    if (isLocked) {
      alert(`Jilid terkunci karena sudah ada indikator yang dinilai pada Iqro ${selectedJilid}. Hapus seluruh nilai terlebih dahulu untuk mengganti jilid.`);
      return;
    }

    selectedJilid = j;
    globalSaveStatus = 'Menyimpan Iqro...';

    try {
      const params = {
        tahun_ajaran_id: Number(tahunAjaranId) || 0,
        semester_id: Number(semesterId) || 0,
        jenis_penilaian: jenisPenilaian,
        mata_pelajaran: 'Tahsin',
        jilid: j
      };

      const res = await fetch(`${apiUrl}/api/guru/siswa/${siswaId}/nilai-tahsin/jilid`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(params)
      });

      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.message || 'Gagal menyimpan pilihan Iqro');
      }

      const body = await res.json();
      headerData = body.header;
      globalSaveStatus = 'Tersimpan';
    } catch (err) {
      globalSaveStatus = 'Gagal menyimpan Iqro';
    }
  }

  function handleScoreInput(event, indikator) {
    let raw = event.target.value.trim();
    if (raw === '') {
      scoreMap[indikator.kode] = '';
      scheduleAutosave(indikator, null);
      return;
    }

    let val = Number(raw);
    if (Number.isNaN(val)) {
      scoreMap[indikator.kode] = '';
      return;
    }

    // Clamp score 1 - 5 for V1
    if (val > 5) val = 5;
    if (val < 1) val = 1;

    event.target.value = val;
    scoreMap[indikator.kode] = val;
    scheduleAutosave(indikator, val);
  }

  function scheduleAutosave(indikator, val) {
    rowStatus.set(indikator.kode, 'Menyimpan...');
    globalSaveStatus = 'Menyimpan...';

    if (saveTimers.has(indikator.kode)) {
      clearTimeout(saveTimers.get(indikator.kode));
    }

    const timer = setTimeout(async () => {
      saveTimers.delete(indikator.kode);
      await saveScoreToServer(indikator, val);
    }, 400);

    saveTimers.set(indikator.kode, timer);
  }

  async function saveScoreToServer(indikator, val) {
    try {
      const payload = {
        tahun_ajaran_id: Number(tahunAjaranId) || 0,
        semester_id: Number(semesterId) || 0,
        jenis_penilaian: jenisPenilaian,
        mata_pelajaran: 'Tahsin',
        jilid: selectedJilid,
        kode: indikator.kode,
        indikator: indikator.indikator,
        urutan: indikator.urutan || 1,
        nilai: val === null || val === '' ? null : Number(val)
      };

      const res = await fetch(`${apiUrl}/api/guru/siswa/${siswaId}/nilai-tahsin/detail`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });

      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.message || 'Gagal menyimpan nilai');
      }

      const body = await res.json();
      headerData = body.header;
      rowStatus.set(indikator.kode, 'Tersimpan');
      globalSaveStatus = 'Tersimpan';
    } catch (err) {
      rowStatus.set(indikator.kode, 'Gagal');
      globalSaveStatus = 'Gagal menyimpan';
    }
  }

  async function confirmClearJilidScores() {
    if (!confirm(`Hapus seluruh nilai indikator pada Iqro ${selectedJilid}?\n\nSetelah dihapus, kunci jilid akan terbuka sehingga Anda dapat memilih jilid yang berbeda.`)) {
      return;
    }

    globalSaveStatus = 'Menghapus nilai...';
    try {
      const payload = {
        tahun_ajaran_id: Number(tahunAjaranId) || 0,
        semester_id: Number(semesterId) || 0,
        jenis_penilaian: jenisPenilaian,
        mata_pelajaran: 'Tahsin'
      };

      const res = await fetch(`${apiUrl}/api/guru/siswa/${siswaId}/nilai-tahsin/reset`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });

      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.message || 'Gagal menghapus nilai');
      }

      scoreMap = {};
      rowStatus.clear();
      globalSaveStatus = 'Nilai terhapus, jilid terbuka';
    } catch (err) {
      globalSaveStatus = 'Gagal menghapus';
      alert('Gagal menghapus nilai: ' + err.message);
    }
  }

  function handleKeyDown(e, idx) {
    if (e.key === 'ArrowDown' || e.key === 'Enter') {
      e.preventDefault();
      const nextInput = document.getElementById(`score-input-${idx + 1}`);
      if (nextInput) {
        nextInput.focus();
        nextInput.select();
      }
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      const prevInput = document.getElementById(`score-input-${idx - 1}`);
      if (prevInput) {
        prevInput.focus();
        prevInput.select();
      }
    }
  }
</script>

<svelte:head>
  <title>Input Nilai Tahsin V1 - {namaSiswa}</title>
</svelte:head>

<div class="sheet-page">
  <!-- TOP STICKY HEADER -->
  <header class="sheet-header">
    <div class="header-left">
      <a class="btn-back" href={backUrl}>
        <Icon name="arrowLeft" size={18} />
        <span>Kembali ke Kelas</span>
      </a>
      <div class="student-info-box">
        <h1 class="student-name">{namaSiswa}</h1>
        <p class="student-meta">{kelas} · Ujian Tahsin (Versi 1 - Skala 1-5)</p>
      </div>
    </div>

    <div class="header-right">
      <div class="status-indicator-pill" class:status-saved={globalSaveStatus === 'Tersimpan'}>
        <span class="status-dot"></span>
        <span class="status-text">{globalSaveStatus}</span>
      </div>
    </div>
  </header>

  {#if error}
    <div class="alert-error">{error}</div>
  {/if}

  <!-- RINGKASAN SKOR & PREDIKAT -->
  <div class="card-section">
    <div class="stats-grid">
      <div class="stat-card">
        <span class="stat-label">Rata-rata Nilai</span>
        <strong class="stat-value text-emerald">{rataRata > 0 ? rataRata : '-'}</strong>
        <span class="stat-desc">Dibagi {totalIndikatorJilid} indikator</span>
      </div>
      <div class="stat-card">
        <span class="stat-label">Indikator Dinilai</span>
        <strong class="stat-value text-blue">{totalDiisi} / {currentIndikators.length}</strong>
      </div>
      <div class="stat-card">
        <span class="stat-label">Predikat</span>
        <strong class="stat-value text-purple">{predikat}</strong>
      </div>
    </div>
  </div>

  <!-- CONTROL: PILIH IQRO YANG DIUJIKAN -->
  <div class="card-section">
    <div class="section-title-row">
      <div class="title-with-icon">
        <Icon name="book" size={20} class="text-emerald" />
        <div>
          <h2 class="section-title">Pilih Iqro yang Diujikan</h2>
          <p class="section-desc">Hanya 1 jilid yang dinilai per siswa. Jilid otomatis terkunci saat nilai diisi.</p>
        </div>
      </div>
      <span class="current-iqro-badge">
        Iqro Aktif: <strong>Jilid {selectedJilid}</strong>
        {#if isLocked}
          <span class="locked-mini-tag">🔒 Terkunci</span>
        {/if}
      </span>
    </div>

    <!-- LOCK WARNING BANNER -->
    {#if isLocked}
      <div class="jilid-lock-banner">
        <div class="lock-banner-left">
          <span class="lock-banner-icon">🔒</span>
          <div>
            <strong>Iqro {selectedJilid} Sedang Terkunci</strong>
            <p>Jilid dikunci karena sudah ada nilai indikator yang diisi. Untuk mengganti ke jilid lain, hapus seluruh nilai terlebih dahulu.</p>
          </div>
        </div>
        <button type="button" class="btn-reset-jilid" onclick={confirmClearJilidScores}>
          <Icon name="trash" size={14} /> Hapus Seluruh Nilai & Buka Kunci
        </button>
      </div>
    {/if}

    <div class="iqro-selector-grid">
      {#each [1, 2, 3, 4, 5, 6] as j}
        {@const lockedOther = isLocked && selectedJilid !== j}
        <button
          type="button"
          class="iqro-btn"
          class:active={selectedJilid === j}
          class:locked-btn={lockedOther}
          disabled={lockedOther}
          onclick={() => selectJilid(j)}
        >
          <div class="iqro-btn-num">Jilid {j}</div>
          <span class="iqro-btn-label">Iqro {j}</span>
          {#if selectedJilid === j}
            <span class="iqro-check-badge">✓ Aktif</span>
          {:else if lockedOther}
            <span class="iqro-lock-badge">🔒 Terkunci</span>
          {/if}
        </button>
      {/each}
    </div>

    <!-- INFO IQRO / FOCUS & KRITERIA -->
    <div class="iqro-info-banner">
      <div class="info-row">
        <span class="info-tag">Fokus Jilid {selectedJilid}:</span>
        <span class="info-val">{currentJilidInfo.fokus || 'Belum diatur'}</span>
      </div>
      {#if currentJilidInfo.kriteria_kelulusan}
        <div class="info-row" style="margin-top: 0.35rem;">
          <span class="info-tag">Kriteria Kelulusan:</span>
          <span class="info-val">{currentJilidInfo.kriteria_kelulusan}</span>
        </div>
      {/if}
    </div>
  </div>

  <!-- LEMBAR PENILAIAN INDIKATOR -->
  <div class="card-section">
    <div class="table-header-info">
      <div>
        <h2 class="section-title">Lembar Penilaian Indikator (Iqro {selectedJilid})</h2>
        <p class="section-desc">
          Isi skor penilaian untuk tiap indikator pada rentang <strong>1 s.d 5</strong>. Nilai tersimpan otomatis.
        </p>
      </div>
      <div class="score-guide-pill desktop-only">
        Skala: <strong>1 - 5</strong> (Tekan ↑ ↓ untuk pindah baris)
      </div>
    </div>

    {#if loading}
      <div class="loading-box">
        <div class="spinner"></div>
        <span>Memuat indikator penilaian...</span>
      </div>
    {:else if currentIndikators.length > 0}
      <!-- DESKTOP TABLE VIEW -->
      <div class="desktop-only table-responsive">
        <table class="tahsin-table">
          <thead>
            <tr>
              <th style="width: 60px; text-align: center;">No</th>
              <th style="width: 110px;">Kode</th>
              <th>Indikator Penilaian</th>
              <th style="width: 170px; text-align: center;">Nilai (1 - 5)</th>
            </tr>
          </thead>
          <tbody>
            {#each currentIndikators as ind, idx (ind.id || ind.kode)}
              {@const currentScore = scoreMap[ind.kode] !== undefined ? scoreMap[ind.kode] : ''}
              {@const status = rowStatus.get(ind.kode)}
              <tr class:row-filled={currentScore !== '' && currentScore !== null}>
                <td style="text-align: center; font-weight: 700; color: #94a3b8;">
                  {idx + 1}
                </td>
                <td>
                  <span class="kode-tag">{ind.kode}</span>
                </td>
                <td>
                  <span class="indikator-label">{ind.indikator}</span>
                </td>
                <td style="text-align: center;">
                  <div class="input-cell-wrapper">
                    <input
                      id={`score-input-${idx}`}
                      type="number"
                      min="1"
                      max="5"
                      class="score-number-input"
                      class:input-filled={currentScore !== '' && currentScore !== null}
                      placeholder="1-5"
                      value={currentScore}
                      oninput={(e) => handleScoreInput(e, ind)}
                      onkeydown={(e) => handleKeyDown(e, idx)}
                    />
                    {#if status}
                      <span class="inline-status" class:status-saved={status === 'Tersimpan'}>
                        {status === 'Tersimpan' ? '✓' : status}
                      </span>
                    {/if}
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>

      <!-- MOBILE CARDS VIEW (LIKE TAHFIZ) -->
      <div class="mobile-only mobile-cards-view">
        {#each currentIndikators as ind, idx (ind.id || ind.kode)}
          {@const currentScore = scoreMap[ind.kode] !== undefined ? scoreMap[ind.kode] : ''}
          {@const status = rowStatus.get(ind.kode)}
          <div class="mobile-indicator-card" class:card-filled={currentScore !== '' && currentScore !== null}>
            <div class="card-top-row">
              <div class="card-badge-group">
                <span class="card-idx-badge">#{idx + 1}</span>
                <span class="card-kode-badge">{ind.kode}</span>
              </div>
              {#if status}
                <span class="card-status-badge" class:status-saved={status === 'Tersimpan'}>
                  {status === 'Tersimpan' ? '✓ Tersimpan' : status}
                </span>
              {/if}
            </div>

            <div class="card-indicator-text">
              {ind.indikator}
            </div>

            <div class="card-input-row">
              <label for={`mobile-score-input-${idx}`} class="card-input-label">Nilai (1 - 5):</label>
              <input
                id={`mobile-score-input-${idx}`}
                type="number"
                min="1"
                max="5"
                class="mobile-score-input"
                class:input-filled={currentScore !== '' && currentScore !== null}
                placeholder="1 - 5"
                value={currentScore}
                oninput={(e) => handleScoreInput(e, ind)}
              />
            </div>
          </div>
        {/each}
      </div>
    {:else}
      <div class="empty-indikator">
        <p>Belum ada indikator yang dikonfigurasi untuk Iqro {selectedJilid}.</p>
        <p class="text-sm">Silakan tambahkan indikator di menu <strong>Konfigurasi Tahsin</strong>.</p>
      </div>
    {/if}
  </div>
</div>

<style>
  :global(body) {
    background-color: #f8fafc;
    margin: 0;
  }

  .desktop-only {
    display: block;
  }
  .mobile-only {
    display: none !important;
  }

  @media (max-width: 768px) {
    .desktop-only {
      display: none !important;
    }
    .mobile-only {
      display: flex !important;
    }
  }

  .sheet-page {
    max-width: 1100px;
    margin: 0 auto;
    padding: 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  /* HEADER */
  .sheet-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.25rem 1.5rem;
    background: #ffffff;
    border-radius: 16px;
    border: 1px solid #e2e8f0;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
    gap: 1rem;
    flex-wrap: wrap;
  }

  .header-left {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .btn-back {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    color: #0b7c3d;
    font-size: 0.88rem;
    font-weight: 700;
    text-decoration: none;
    transition: opacity 0.2s;
  }
  .btn-back:hover {
    opacity: 0.8;
  }

  .student-name {
    margin: 0;
    font-size: 1.5rem;
    font-weight: 800;
    color: #0f172a;
  }

  .student-meta {
    margin: 0.2rem 0 0;
    font-size: 0.85rem;
    color: #64748b;
  }

  .status-indicator-pill {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.45rem 0.95rem;
    background: #f1f5f9;
    border: 1px solid #e2e8f0;
    border-radius: 9999px;
    font-size: 0.82rem;
    font-weight: 700;
    color: #64748b;
  }
  .status-indicator-pill.status-saved {
    background: #ecfdf5;
    border-color: #a7f3d0;
    color: #059669;
  }
  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #059669;
  }

  .alert-error {
    padding: 0.75rem 1rem;
    border-radius: 10px;
    background: #fef2f2;
    border: 1px solid #fecaca;
    color: #dc2626;
    font-weight: 600;
  }

  /* CARD SECTIONS */
  .card-section {
    background: #ffffff;
    border-radius: 16px;
    border: 1px solid #e2e8f0;
    padding: 1.5rem;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.03);
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 1rem;
  }

  .stat-card {
    display: flex;
    flex-direction: column;
    padding: 1rem;
    background: #f8fafc;
    border-radius: 12px;
    border: 1px solid #f1f5f9;
  }

  .stat-label {
    font-size: 0.8rem;
    font-weight: 600;
    color: #64748b;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .stat-desc {
    margin-top: 0.25rem;
    font-size: 0.75rem;
    font-weight: 600;
    color: #94a3b8;
  }

  .stat-value {
    margin-top: 0.35rem;
    font-size: 1.6rem;
    font-weight: 800;
  }

  .text-emerald {
    color: #0b7c3d;
  }
  .text-blue {
    color: #2563eb;
  }
  .text-purple {
    color: #7c3aed;
  }

  /* IQRO SELECTOR */
  .section-title-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 1rem;
    margin-bottom: 1.25rem;
  }

  .title-with-icon {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .section-title {
    margin: 0;
    font-size: 1.15rem;
    font-weight: 800;
    color: #0f172a;
  }

  .section-desc {
    margin: 0.2rem 0 0;
    font-size: 0.85rem;
    color: #64748b;
  }

  .current-iqro-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.35rem 0.85rem;
    background: #ecfdf5;
    border: 1px solid #a7f3d0;
    color: #065f46;
    border-radius: 9999px;
    font-size: 0.85rem;
  }

  .locked-mini-tag {
    font-size: 0.72rem;
    font-weight: 700;
    background: #fef3c7;
    color: #92400e;
    padding: 0.1rem 0.45rem;
    border-radius: 9999px;
    border: 1px solid #fde68a;
  }

  /* LOCK WARNING BANNER */
  .jilid-lock-banner {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    background: #fffbeb;
    border: 1.5px solid #fde68a;
    border-radius: 12px;
    padding: 0.85rem 1.15rem;
    margin-bottom: 1.25rem;
    flex-wrap: wrap;
  }

  .lock-banner-left {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex: 1;
    min-width: 260px;
  }

  .lock-banner-icon {
    font-size: 1.4rem;
  }

  .lock-banner-left strong {
    display: block;
    font-size: 0.9rem;
    color: #92400e;
  }

  .lock-banner-left p {
    margin: 0.15rem 0 0;
    font-size: 0.8rem;
    color: #b45309;
    line-height: 1.35;
  }

  .btn-reset-jilid {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    background: #ffffff;
    border: 1.5px solid #f59e0b;
    color: #b45309;
    padding: 0.45rem 0.85rem;
    border-radius: 8px;
    font-size: 0.8rem;
    font-weight: 700;
    cursor: pointer;
    transition: all 0.2s;
    white-space: nowrap;
  }

  .btn-reset-jilid:hover {
    background: #fef3c7;
    color: #92400e;
  }

  .iqro-selector-grid {
    display: grid;
    grid-template-columns: repeat(6, 1fr);
    gap: 0.75rem;
    margin-bottom: 1.25rem;
  }

  @media (max-width: 768px) {
    .iqro-selector-grid {
      grid-template-columns: repeat(3, 1fr);
    }
  }

  .iqro-btn {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 0.85rem 0.5rem;
    background: #ffffff;
    border: 1.5px solid #e2e8f0;
    border-radius: 12px;
    cursor: pointer;
    transition: all 0.2s ease;
    gap: 0.2rem;
  }

  .iqro-btn:hover:not(:disabled) {
    border-color: #0b7c3d;
    background: #f8fafc;
    transform: translateY(-2px);
  }

  .iqro-btn.active {
    background: #0b7c3d;
    border-color: #096833;
    color: #ffffff;
    box-shadow: 0 4px 10px rgba(11, 124, 61, 0.25);
  }

  .iqro-btn.locked-btn {
    opacity: 0.45;
    cursor: not-allowed;
    background: #f8fafc;
    border-color: #e2e8f0;
  }

  .iqro-btn-num {
    font-size: 1rem;
    font-weight: 800;
  }

  .iqro-btn-label {
    font-size: 0.78rem;
    color: #64748b;
  }
  .iqro-btn.active .iqro-btn-label {
    color: rgba(255, 255, 255, 0.9);
  }

  .iqro-check-badge {
    margin-top: 0.25rem;
    font-size: 0.7rem;
    font-weight: 700;
    background: rgba(255, 255, 255, 0.25);
    padding: 0.1rem 0.5rem;
    border-radius: 9999px;
  }

  .iqro-lock-badge {
    margin-top: 0.25rem;
    font-size: 0.68rem;
    font-weight: 700;
    color: #92400e;
    background: #fef3c7;
    padding: 0.1rem 0.4rem;
    border-radius: 9999px;
  }

  .iqro-info-banner {
    padding: 0.85rem 1rem;
    background: #f8fafc;
    border-radius: 10px;
    border: 1px solid #e2e8f0;
    font-size: 0.85rem;
  }

  .info-tag {
    font-weight: 700;
    color: #334155;
    margin-right: 0.5rem;
  }
  .info-val {
    color: #0f172a;
  }

  /* TABLE (DESKTOP) */
  .table-header-info {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
    flex-wrap: wrap;
    gap: 0.75rem;
  }

  .score-guide-pill {
    padding: 0.35rem 0.85rem;
    background: #f1f5f9;
    border-radius: 9999px;
    font-size: 0.82rem;
    color: #475569;
    font-weight: 600;
  }

  .table-responsive {
    overflow-x: auto;
  }

  .tahsin-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.9rem;
  }

  .tahsin-table th {
    background: #f8fafc;
    padding: 0.75rem 1rem;
    text-align: left;
    font-weight: 700;
    color: #475569;
    border-bottom: 2px solid #e2e8f0;
  }

  .tahsin-table td {
    padding: 0.75rem 1rem;
    border-bottom: 1px solid #f1f5f9;
    vertical-align: middle;
  }

  .tahsin-table tr:hover {
    background: #f8fafc;
  }

  .tahsin-table tr.row-filled {
    background: #fafffc;
  }

  .kode-tag {
    display: inline-block;
    padding: 0.2rem 0.5rem;
    background: #f1f5f9;
    border-radius: 6px;
    font-weight: 700;
    font-size: 0.8rem;
    color: #0f172a;
  }

  .indikator-label {
    font-weight: 600;
    color: #1e293b;
    line-height: 1.4;
  }

  .input-cell-wrapper {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
  }

  .score-number-input {
    width: 90px;
    padding: 0.5rem 0.65rem;
    font-size: 1rem;
    font-weight: 800;
    text-align: center;
    border-radius: 8px;
    border: 1.5px solid #cbd5e1;
    outline: none;
    transition: all 0.2s;
  }

  .score-number-input:focus {
    border-color: #0b7c3d;
    box-shadow: 0 0 0 3px rgba(11, 124, 61, 0.15);
  }

  .score-number-input.input-filled {
    border-color: #a7f3d0;
    background: #f0fdf4;
    color: #065f46;
  }

  .inline-status {
    font-size: 0.75rem;
    font-weight: 700;
    color: #64748b;
  }
  .inline-status.status-saved {
    color: #059669;
  }

  /* MOBILE CARDS VIEW (LIKE TAHFIZ) */
  .mobile-cards-view {
    display: flex;
    flex-direction: column;
    gap: 0.85rem;
  }

  .mobile-indicator-card {
    background: #ffffff;
    border: 1.5px solid #e2e8f0;
    border-radius: 12px;
    padding: 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    transition: all 0.2s;
  }

  .mobile-indicator-card.card-filled {
    border-color: #a7f3d0;
    background: #fafffc;
    border-left: 4px solid #0b7c3d;
  }

  .card-top-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .card-badge-group {
    display: flex;
    align-items: center;
    gap: 0.45rem;
  }

  .card-idx-badge {
    font-weight: 800;
    font-size: 0.85rem;
    color: #94a3b8;
  }

  .card-kode-badge {
    padding: 0.2rem 0.55rem;
    background: #f1f5f9;
    border-radius: 6px;
    font-weight: 800;
    font-size: 0.8rem;
    color: #0f172a;
  }

  .card-status-badge {
    font-size: 0.75rem;
    font-weight: 700;
    color: #64748b;
  }
  .card-status-badge.status-saved {
    color: #059669;
  }

  .card-indicator-text {
    font-size: 0.95rem;
    font-weight: 600;
    color: #1e293b;
    line-height: 1.45;
  }

  .card-input-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-top: 0.65rem;
    border-top: 1px solid #f1f5f9;
    gap: 0.75rem;
  }

  .card-input-label {
    font-size: 0.85rem;
    font-weight: 700;
    color: #475569;
  }

  .mobile-score-input {
    width: 100px;
    padding: 0.55rem 0.75rem;
    font-size: 1.1rem;
    font-weight: 800;
    text-align: center;
    border-radius: 10px;
    border: 1.5px solid #cbd5e1;
    outline: none;
    background: #fff;
    transition: all 0.2s;
  }

  .mobile-score-input:focus {
    border-color: #0b7c3d;
    box-shadow: 0 0 0 3px rgba(11, 124, 61, 0.15);
  }

  .mobile-score-input.input-filled {
    border-color: #a7f3d0;
    background: #f0fdf4;
    color: #065f46;
  }

  .loading-box {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 2.5rem;
    gap: 0.75rem;
    color: #64748b;
  }

  .spinner {
    width: 30px;
    height: 30px;
    border: 3px solid #e2e8f0;
    border-top-color: #0b7c3d;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .empty-indikator {
    text-align: center;
    padding: 2.5rem;
    color: #64748b;
  }
</style>
