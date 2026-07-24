<script>
  import { SvelteMap } from 'svelte/reactivity';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import Icon from '$lib/components/ui/Icon.svelte';

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';
  const mataPelajaran = 'Tahfiz';

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
  let pageTitle = $derived(jenisPenilaian === 'kelulusan' ? 'Ujian Kelulusan' : 'Ujian Tahfiz');

  let backUrl = $derived(
    kelasId
      ? `/guru/input-nilai/tahfiz/kelas/${kelasId}?kelas=${encodeURIComponent(kelasLabel)}&rombel=${encodeURIComponent(rombel)}&sekolah_id=${sekolahId}&jenis_penilaian=${encodeURIComponent(jenisPenilaian)}${tahunAjaranId ? `&tahun_ajaran_id=${tahunAjaranId}` : ''}${semesterId ? `&semester_id=${semesterId}` : ''}`
      : '/guru/input-nilai/tahfiz'
  );

  let rows = $state([]);
  let materiTidakAda = $state(false);

  let loading = $state(false);
  let loadError = $state('');
  let lastSavedSummary = $state(null);
  const saveTimers = new SvelteMap();

  onMount(init);

  function clampScore(value) {
    if (value === '' || value === null || value === undefined) return '';
    const numberValue = Number(value);
    if (Number.isNaN(numberValue)) return '';
    return Math.min(5, Math.max(0, numberValue));
  }

  function total(row) {
    const fields = [row.hafalan, row.kelancaran, row.mad, row.tajwid, row.makharijulHuruf];
    if (fields.some((value) => value === '')) return '';
    return fields.reduce((sum, value) => sum + Number(value), 0);
  }

  function toInputValue(value) {
    if (value === null || value === undefined) return '';
    return Number(value);
  }

  function scoreNumber(value) {
    if (value === '' || value === null || value === undefined) return null;
    const numberValue = Number(value);
    return Number.isFinite(numberValue) ? numberValue : null;
  }

  async function init() {
    loading = true;
    loadError = '';
    try {
      // 1. Fetch daftar materi aktif dari backend
      const materiRes = await fetch(`${apiUrl}/api/guru/sekolah/${sekolahId}/materi-tahfiz`, {
        credentials: 'include'
      });
      if (!materiRes.ok) {
        const body = await materiRes.json().catch(() => ({}));
        throw new Error(body.message || 'Gagal memuat daftar materi.');
      }
      const materiData = await materiRes.json();

      if (!materiData || materiData.length === 0) {
        materiTidakAda = true;
        return;
      }

      // 2. Bangun rows dari materi
      rows = materiData.map((m) => ({
        namaMateri: m.nama_materi,
        juz: m.juz || 0,
        urutan: m.urutan,
        hafalan: '',
        kelancaran: '',
        mad: '',
        tajwid: '',
        makharijulHuruf: '',
        saveStatus: ''
      }));

      // 3. Fetch nilai yang sudah ada
      await loadNilaiTahfiz();
    } catch (err) {
      loadError = err.message || 'Gagal memuat halaman.';
    } finally {
      loading = false;
    }
  }

  async function loadNilaiTahfiz() {
    try {
      const params = new URLSearchParams({
        mata_pelajaran: mataPelajaran,
        jenis_penilaian: jenisPenilaian
      });
      if (tahunAjaranId) params.append('tahun_ajaran_id', tahunAjaranId);
      if (semesterId) params.append('semester_id', semesterId);
      const res = await fetch(
        `${apiUrl}/api/guru/siswa/${siswaId}/nilai-tahfiz?${params.toString()}`,
        {
          credentials: 'include'
        }
      );
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.message || 'Gagal memuat nilai tahfiz.');
      }

      const body = await res.json();
      const detailMap = new Map((body.details || []).map((detail) => [detail.nama_materi, detail]));
      rows = rows.map((row) => {
        const detail = detailMap.get(row.namaMateri);
        if (!detail) return row;
        return {
          ...row,
          hafalan: toInputValue(detail.skor_hafalan),
          kelancaran: toInputValue(detail.skor_kelancaran),
          mad: toInputValue(detail.skor_mad),
          tajwid: toInputValue(detail.skor_tajwid),
          makharijulHuruf: toInputValue(detail.skor_makharijul_huruf),
          saveStatus: 'Tersimpan'
        };
      });
      lastSavedSummary = body.nilai || null;
    } catch (err) {
      loadError = err.message || 'Gagal memuat nilai tahfiz.';
    }
  }

  function handleScoreInput(event, rowIndex, field) {
    let val = event.target.value;
    if (val.length > 1 && val.startsWith('0')) {
      val = Number(val).toString();
      event.target.value = val;
    }
    rows[rowIndex][field] = clampScore(val);
    scheduleAutosave(rowIndex);
  }

  function scheduleAutosave(rowIndex) {
    const existingTimer = saveTimers.get(rowIndex);
    if (existingTimer) clearTimeout(existingTimer);

    rows[rowIndex].saveStatus = 'Menunggu...';
    const timer = setTimeout(() => {
      saveTimers.delete(rowIndex);
      autosaveRow(rowIndex);
    }, 650);
    saveTimers.set(rowIndex, timer);
  }

  async function autosaveRow(rowIndex) {
    const row = rows[rowIndex];
    row.saveStatus = 'Menyimpan...';

    try {
      const res = await fetch(`${apiUrl}/api/guru/siswa/${siswaId}/nilai-tahfiz/detail`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          mata_pelajaran: mataPelajaran,
          jenis_penilaian: jenisPenilaian,
          tahun_ajaran_id: tahunAjaranId ? Number(tahunAjaranId) : undefined,
          semester_id: semesterId ? Number(semesterId) : undefined,
          nama_materi: row.namaMateri,
          juz: row.juz,
          urutan: rowIndex + 1,
          skor_hafalan: scoreNumber(row.hafalan),
          skor_kelancaran: scoreNumber(row.kelancaran),
          skor_mad: scoreNumber(row.mad),
          skor_tajwid: scoreNumber(row.tajwid),
          skor_makharijul_huruf: scoreNumber(row.makharijulHuruf)
        })
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.message || 'Gagal menyimpan.');
      }

      const body = await res.json();
      lastSavedSummary = body.nilai || null;
      row.saveStatus = 'Tersimpan';
    } catch (err) {
      row.saveStatus = err.message || 'Gagal';
    }
  }

  let submittedRows = $derived(rows.filter((row) => total(row) !== ''));
  let totalPoin = $derived(submittedRows.reduce((sum, row) => sum + total(row), 0));
  let rataRataPoin = $derived(
    submittedRows.length > 0 ? Number((totalPoin / submittedRows.length).toFixed(2)) : ''
  );

  function moveCell(event, rowIndex, colIndex) {
    const keys = {
      ArrowUp: [rowIndex - 1, colIndex],
      ArrowDown: [rowIndex + 1, colIndex],
      ArrowLeft: [rowIndex, colIndex - 1],
      ArrowRight: [rowIndex, colIndex + 1]
    };

    if (!keys[event.key]) return;

    const [nextRow, nextCol] = keys[event.key];
    const nextCell = document.querySelector(`[data-cell="${nextRow}-${nextCol}"]`);
    if (!nextCell) return;

    event.preventDefault();
    nextCell.focus();
    nextCell.select();
  }
</script>

<svelte:head>
  <title>{pageTitle} - {namaSiswa}</title>
</svelte:head>

<main class="sheet-page">
  <div class="back-button-wrapper">
    <button type="button" class="back-button" onclick={() => goto(backUrl)}>
      <Icon name="arrowLeft" size={16} /> Kembali
    </button>
  </div>

  <section class="sheet-header">
    <div>
      <h1>{pageTitle}</h1>
      <p>{namaSiswa} · {kelas}</p>
    </div>
    <div class="meta-box">
      <span>Juz 30</span>
      <strong>{lastSavedSummary?.jumlah_materi ?? submittedRows.length} Materi Disetor</strong>
      <small>{loading ? 'Memuat...' : 'Autosave aktif'}</small>
    </div>
  </section>

  {#if loadError}
    <div class="sheet-alert">{loadError}</div>
  {/if}

  {#if loading}
    <div class="sheet-loading">Memuat data...</div>
  {:else if materiTidakAda}
    <div class="sheet-empty">
      <div class="sheet-empty-icon">📋</div>
      <strong>Materi tahfiz belum dikonfigurasi</strong>
      <p>Hubungi admin untuk menambahkan daftar materi sebelum input nilai dapat dilakukan.</p>
    </div>
  {:else}
    <section class="sheet-wrap" aria-label="Tabel ujian tahfiz">
      <table class="excel-table">
        <thead>
          <tr>
            <th class="col-no">No</th>
            <th class="col-surah">Nama Surah</th>
            <th class="col-juz">Juz</th>
            <th>Hafalan (1-5)</th>
            <th>Kelancaran (1-5)</th>
            <th>Mad (1-5)</th>
            <th>Tajwid (1-5)</th>
            <th>Makharijul Huruf (1-5)</th>
            <th class="col-total">Total</th>
          </tr>
        </thead>
        <tbody>
          {#each rows as row, index (row.namaMateri)}
            <tr>
              <td class="locked text-center">{index + 1}</td>
              <td class="locked">{row.namaMateri}</td>
              <td class="locked text-center">{row.juz > 0 ? row.juz : '—'}</td>
              <td>
                <input
                  type="number"
                  min="0"
                  max="5"
                  data-cell={`${index}-0`}
                  bind:value={row.hafalan}
                  oninput={(e) => handleScoreInput(e, index, 'hafalan')}
                  onkeydown={(event) => moveCell(event, index, 0)}
                />
              </td>
              <td>
                <input
                  type="number"
                  min="0"
                  max="5"
                  data-cell={`${index}-1`}
                  bind:value={row.kelancaran}
                  oninput={(e) => handleScoreInput(e, index, 'kelancaran')}
                  onkeydown={(event) => moveCell(event, index, 1)}
                />
              </td>
              <td>
                <input
                  type="number"
                  min="0"
                  max="5"
                  data-cell={`${index}-2`}
                  bind:value={row.mad}
                  oninput={(e) => handleScoreInput(e, index, 'mad')}
                  onkeydown={(event) => moveCell(event, index, 2)}
                />
              </td>
              <td>
                <input
                  type="number"
                  min="0"
                  max="5"
                  data-cell={`${index}-3`}
                  bind:value={row.tajwid}
                  oninput={(e) => handleScoreInput(e, index, 'tajwid')}
                  onkeydown={(event) => moveCell(event, index, 3)}
                />
              </td>
              <td>
                <input
                  type="number"
                  min="0"
                  max="5"
                  data-cell={`${index}-4`}
                  bind:value={row.makharijulHuruf}
                  oninput={(e) => handleScoreInput(e, index, 'makharijulHuruf')}
                  onkeydown={(event) => moveCell(event, index, 4)}
                />
              </td>
              <td class="locked total-cell text-center">
                <strong>{total(row)}</strong>
                {#if row.saveStatus}
                  <span class="save-state">{row.saveStatus}</span>
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
        <tfoot>
          <tr>
            <td colspan="8" class="summary-label">Jumlah Materi Disetor</td>
            <td class="summary-value text-center">{submittedRows.length}</td>
          </tr>
          <tr>
            <td colspan="8" class="summary-label">Jumlah Poin</td>
            <td class="summary-value text-center">{totalPoin}</td>
          </tr>
          <tr>
            <td colspan="8" class="summary-label">Preview Rata-rata Poin</td>
            <td class="summary-value final-score text-center">
              {lastSavedSummary?.nilai_akhir ?? rataRataPoin}
            </td>
          </tr>
        </tfoot>
      </table>
    </section>
  {/if}
</main>

<style>
  :global(body) {
    background: #f3f4f6;
  }

  .sheet-loading {
    max-width: 1320px;
    margin: 0 auto;
    padding: 2rem;
    text-align: center;
    color: #6b7280;
    font-size: 13px;
    border: 1px solid #e5e7eb;
    background: #fff;
  }

  .sheet-empty {
    max-width: 1320px;
    margin: 0 auto;
    padding: 3rem 2rem;
    text-align: center;
    background: #fff;
    border: 1px solid #e5e7eb;
    color: #374151;
  }

  .sheet-empty-icon {
    font-size: 2.5rem;
    margin-bottom: 0.75rem;
  }

  .sheet-empty strong {
    display: block;
    font-size: 1rem;
    margin-bottom: 0.5rem;
  }

  .sheet-empty p {
    font-size: 13px;
    color: #6b7280;
    margin: 0;
  }

  .sheet-page {
    min-height: 100vh;
    padding: 24px;
    color: #111827;
    font-family: Arial, Helvetica, sans-serif;
  }

  .sheet-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    max-width: 1320px;
    margin: 0 auto 12px;
  }

  .back-button-wrapper {
    max-width: 1320px;
    margin: 0 auto;
    display: flex;
    justify-content: flex-start;
  }

  .back-button {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    margin: 0 0 16px;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    background: #fff;
    color: #475569;
    padding: 8px 16px;
    font-weight: 600;
    font-size: 13px;
    cursor: pointer;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
    transition: all 0.2s ease;
  }

  .back-button:hover {
    background: #f8fafc;
    border-color: #cbd5e1;
    color: #0f172a;
  }

  .sheet-header h1 {
    margin: 0;
    font-size: 22px;
    line-height: 1.2;
  }

  .sheet-header p {
    margin: 4px 0 0;
    color: #4b5563;
    font-size: 13px;
  }

  .meta-box {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 160px;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    background: #fff;
    padding: 12px 16px;
    text-align: right;
    font-size: 12px;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  }

  .meta-box span {
    color: #4b5563;
  }

  .meta-box small {
    color: #16a34a;
    font-size: 11px;
    font-weight: 700;
  }

  .sheet-alert {
    max-width: 1320px;
    margin: 0 auto 12px;
    border: 1px solid #fca5a5;
    background: #fef2f2;
    color: #991b1b;
    padding: 8px 10px;
    font-size: 12px;
    font-weight: 700;
  }

  .sheet-wrap {
    max-width: 1320px;
    margin: 0 auto;
    overflow: auto;
    background: #fff;
    border: 1px solid #cbd5e1;
    border-radius: 8px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  }

  .excel-table {
    width: 100%;
    min-width: 1100px;
    border-collapse: collapse;
    table-layout: fixed;
    font-size: 13px;
  }

  .excel-table th,
  .excel-table td {
    border: 1px solid #cbd5e1;
    height: 36px;
    padding: 0;
  }

  .excel-table th {
    background: #d9ead3;
    color: #111827;
    font-weight: 700;
    text-align: center;
    padding: 6px;
  }

  .excel-table tbody tr:nth-child(even) .locked {
    background: #f8fafc;
  }

  .excel-table input {
    width: 100%;
    height: 32px;
    border: 0;
    padding: 0 8px;
    box-sizing: border-box;
    text-align: center;
    font: inherit;
    background: #fff;
    outline: none;
  }

  .excel-table input:focus {
    box-shadow: inset 0 0 0 2px #16a34a;
    background: #f0fdf4;
  }

  .locked {
    background: #eef2ff;
    color: #111827;
    font-weight: 600;
    padding: 0 8px;
  }

  .total-cell {
    background: #fef3c7;
    font-weight: 800;
  }

  .total-cell strong,
  .save-state {
    display: block;
  }

  .save-state {
    margin-top: 1px;
    color: #475569;
    font-size: 10px;
    font-weight: 700;
    line-height: 1.1;
  }

  .summary-label,
  .summary-value {
    background: #d9ead3;
    font-weight: 800;
    padding: 0 8px;
  }

  .summary-label {
    text-align: right;
  }

  .summary-value {
    background: #fef3c7;
  }

  .final-score {
    color: #14532d;
    font-size: 13px;
  }

  .text-center {
    text-align: center;
  }

  .col-no {
    width: 48px;
  }

  .col-surah {
    width: 190px;
  }

  .col-juz {
    width: 70px;
  }

  .col-total {
    width: 90px;
  }

  @media (max-width: 720px) {
    .sheet-page {
      padding: 12px;
    }

    .sheet-header {
      align-items: stretch;
      flex-direction: column;
    }

    .meta-box {
      text-align: left;
    }
  }
</style>
