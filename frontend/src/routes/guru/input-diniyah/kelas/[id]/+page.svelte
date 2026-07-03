<script>
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import Icon from '$lib/components/ui/Icon.svelte';
  import { goto } from '$app/navigation';

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

  // URL params
  let kelasId = $derived($page.params.id);
  let kelasLabel = $derived($page.url.searchParams.get('kelas') || '');
  let rombelParam = $derived($page.url.searchParams.get('rombel') || '__kosong__');
  let sekolahId = $derived($page.url.searchParams.get('sekolah_id') || '');

  // State
  let mapelList = $state([]);
  let selectedMapel = $state(null);
  let siswaList = $state([]);
  let nilaiMap = $state({});
  let loadingMapel = $state(false);
  let loadingNilai = $state(false);
  let error = $state('');
  let savingCell = $state('');
  let uhColumnCount = $state(4);

  const KKM = 65;

  onMount(loadMapel);

  async function loadMapel() {
    loadingMapel = true;
    try {
      const res = await fetch(`${apiUrl}/api/guru/sekolah/${sekolahId}/mapel-diniyah`, {
        credentials: 'include'
      });
      if (!res.ok) throw new Error('Gagal memuat daftar mapel diniyah.');
      mapelList = await res.json();
      if (mapelList.length > 0) {
        selectedMapel = mapelList[0];
        await loadNilai();
      }
    } catch (err) {
      error = err.message;
    } finally {
      loadingMapel = false;
    }
  }

  async function loadNilai() {
    if (!selectedMapel) return;
    loadingNilai = true;
    error = '';
    try {
      const params = new URLSearchParams({
        mapel_diniyah_id: String(selectedMapel.id),
        rombel: rombelParam
      });
      const res = await fetch(
        `${apiUrl}/api/guru/kelas/${kelasId}/nilai-diniyah?${params.toString()}`,
        { credentials: 'include' }
      );
      if (!res.ok) throw new Error('Gagal memuat data nilai.');
      const body = await res.json();
      siswaList = body.siswa || [];

      // Build nilaiMap: siswa_id -> nilaiData
      const map = {};
      for (const n of body.nilai || []) {
        map[n.siswa_id] = {
          id: n.id,
          uh_scores: n.uh_scores || [],
          rata_uh: n.rata_uh || 0,
          murni_pts: n.murni_pts || 0,
          remidi_pts: n.remidi_pts,
          na_pts: n.na_pts || 0,
          pas: n.pas || 0,
          nilai_rapor: n.nilai_rapor || 0
        };
      }
      nilaiMap = map;

      // Determine UH column count from existing data
      // Lock to 4 columns explicitly for ulangan harian
      uhColumnCount = 4;
    } catch (err) {
      error = err.message;
    } finally {
      loadingNilai = false;
    }
  }

  function getSiswaData(siswaId) {
    return (
      nilaiMap[siswaId] || {
        uh_scores: [],
        rata_uh: 0,
        murni_pts: 0,
        remidi_pts: null,
        na_pts: 0,
        pas: 0,
        nilai_rapor: 0
      }
    );
  }

  function getUhScore(siswaId, idx) {
    const data = getSiswaData(siswaId);
    return data.uh_scores[idx] ?? '';
  }

  function clamp(val) {
    const n = parseInt(val, 10);
    if (isNaN(n) || n < 0) return 0;
    if (n > 100) return 100;
    return n;
  }

  // Client-side calculation
  function recalculate(siswaId) {
    const data = getSiswaData(siswaId);
    const uhArr = data.uh_scores || [];

    // Rata-rata UH
    const filled = uhArr.filter((v) => v > 0);
    const rataUh = filled.length > 0 ? filled.reduce((a, b) => a + b, 0) / filled.length : 0;

    // NA PTS
    const m = data.murni_pts || 0;
    const r = data.remidi_pts;
    let naPts = m;
    if (m < KKM && r !== null && r !== undefined && r > 0) {
      naPts = KKM;
    }

    // NR
    const pas = data.pas || 0;
    const nr = rataUh * 0.4 + naPts * 0.3 + pas * 0.3;

    data.rata_uh = Math.round(rataUh * 100) / 100;
    data.na_pts = naPts;
    data.nilai_rapor = Math.round(nr * 100) / 100;

    nilaiMap[siswaId] = { ...data };
    nilaiMap = { ...nilaiMap };
  }

  function updateUh(siswaId, idx, value) {
    const data = getSiswaData(siswaId);
    if (!data.uh_scores) data.uh_scores = [];
    // Ensure array is long enough
    while (data.uh_scores.length <= idx) data.uh_scores.push(0);
    data.uh_scores[idx] = clamp(value);
    nilaiMap[siswaId] = { ...data };
    nilaiMap = { ...nilaiMap };
    recalculate(siswaId);
  }

  function updateField(siswaId, field, value) {
    const data = getSiswaData(siswaId);
    if (field === 'remidi_pts') {
      const v = value === '' || value === null ? null : clamp(value);
      data.remidi_pts = v;
    } else {
      data[field] = clamp(value);
    }
    nilaiMap[siswaId] = { ...data };
    nilaiMap = { ...nilaiMap };
    recalculate(siswaId);
  }

  async function saveNilai(siswaId) {
    const data = getSiswaData(siswaId);
    savingCell = String(siswaId);
    try {
      const payload = {
        siswa_id: siswaId,
        sekolah_id: Number(sekolahId),
        mapel_diniyah_id: selectedMapel.id,
        uh_scores: data.uh_scores || [],
        murni_pts: data.murni_pts || 0,
        remidi_pts: data.remidi_pts,
        pas: data.pas || 0
      };
      const res = await fetch(`${apiUrl}/api/guru/nilai-diniyah`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.message || 'Gagal menyimpan');
      }
      const saved = await res.json();
      nilaiMap[siswaId] = {
        id: saved.id,
        uh_scores: saved.uh_scores || [],
        rata_uh: saved.rata_uh || 0,
        murni_pts: saved.murni_pts || 0,
        remidi_pts: saved.remidi_pts,
        na_pts: saved.na_pts || 0,
        pas: saved.pas || 0,
        nilai_rapor: saved.nilai_rapor || 0
      };
      nilaiMap = { ...nilaiMap };
    } catch (err) {
      error = err.message;
    } finally {
      savingCell = '';
    }
  }

  function addUhColumn() {
    uhColumnCount += 1;
  }

  function removeUhColumn() {
    if (uhColumnCount <= 1) return;
    uhColumnCount -= 1;
    // Truncate all existing data
    for (const key in nilaiMap) {
      const data = nilaiMap[key];
      if (data.uh_scores && data.uh_scores.length > uhColumnCount) {
        data.uh_scores = data.uh_scores.slice(0, uhColumnCount);
      }
    }
    nilaiMap = { ...nilaiMap };
  }

  function onSelectMapel(e) {
    const id = Number(e.target.value);
    selectedMapel = mapelList.find((m) => m.id === id) || null;
    if (selectedMapel) loadNilai();
  }

  function formatDecimal(val) {
    if (val === 0 || val === null || val === undefined) return '-';
    return Number(val).toFixed(2);
  }

  function downloadTemplate() {
    if (!selectedMapel) return;
    const params = new URLSearchParams({ rombel: rombelParam });
    window.location.href = `${apiUrl}/api/guru/kelas/${kelasId}/mapel-diniyah/${selectedMapel.id}/template?${params.toString()}`;
  }

  function handlePaste(e) {
    const text = e.clipboardData.getData('text/plain');
    if (!text) return;

    const target = e.target;
    if (!target.classList.contains('input-cell')) return;

    e.preventDefault();

    const rows = text
      .trim()
      .split('\n')
      .map((row) => row.split('\t'));
    const td = target.closest('td');
    const tr = target.closest('tr');
    const tbody = tr.closest('tbody');

    const startRowIdx = Array.from(tbody.children).indexOf(tr);
    const startColIdx = Array.from(tr.children).indexOf(td);

    let currentSiswaIdxToSave = new Set();

    for (let r = 0; r < rows.length; r++) {
      const targetTr = tbody.children[startRowIdx + r];
      if (!targetTr) break;

      const siswaIdStr = targetTr.getAttribute('data-siswa-id');
      if (!siswaIdStr) continue;
      const siswaId = parseInt(siswaIdStr, 10);

      for (let c = 0; c < rows[r].length; c++) {
        const valStr = rows[r][c].trim().replace(',', '.');
        const val = parseFloat(valStr);
        if (isNaN(val)) continue;

        const targetTd = targetTr.children[startColIdx + c];
        if (!targetTd) break;

        const input = targetTd.querySelector('input.input-cell');
        if (input) {
          const field = input.getAttribute('data-field');
          if (field === 'uh') {
            const idx = parseInt(input.getAttribute('data-idx'), 10);
            updateUh(siswaId, idx, val);
          } else if (field === 'murni_pts' || field === 'remidi_pts' || field === 'pas') {
            updateField(siswaId, field, val);
          }
          currentSiswaIdxToSave.add(siswaId);
        }
      }
    }

    const ids = Array.from(currentSiswaIdxToSave);
    (async () => {
      for (const id of ids) {
        await saveNilai(id);
      }
    })();
  }

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
  <title>Input Nilai Diniyah - {kelasLabel} - SIAKAD</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <button class="btn-back" onclick={() => goto('/guru/input-diniyah')}>
      <Icon name="chevronLeft" size={16} /> Kembali
    </button>
    <div>
      <h1 class="page-title">Input Nilai Diniyah</h1>
      <p class="page-subtitle">Kelas {kelasLabel}</p>
    </div>
  </div>

  {#if error}
    <div class="alert-error">
      {error} <button class="btn-dismiss" onclick={() => (error = '')}>✕</button>
    </div>
  {/if}

  <!-- Mapel Selector -->
  <div class="toolbar">
    <div class="toolbar-left">
      <label for="select-mapel" class="toolbar-label">Mata Pelajaran:</label>
      {#if loadingMapel}
        <span class="text-muted">Memuat...</span>
      {:else if mapelList.length === 0}
        <span class="text-muted">Belum ada mapel diniyah dikonfigurasi admin.</span>
      {:else}
        <select id="select-mapel" class="select-mapel" onchange={onSelectMapel}>
          {#each mapelList as m (m.id)}
            <option value={m.id} selected={selectedMapel?.id === m.id}>
              {m.nama_mapel}
              {m.nama_mapel_arab ? `(${m.nama_mapel_arab})` : ''}
            </option>
          {/each}
        </select>
      {/if}
    </div>

    <div class="toolbar-right">
      <button class="btn-sm btn-primary" onclick={downloadTemplate}>
        <Icon name="download" size={14} /> Template Excel
      </button>
      <span class="uh-info">{uhColumnCount} kolom UH</span>
    </div>
  </div>

  <!-- Tabel Input -->
  {#if loadingNilai}
    <div class="table-loading">Memuat data nilai...</div>
  {:else if selectedMapel && siswaList.length > 0}
    <div class="table-wrapper">
      <table class="nilai-table">
        <thead>
          <tr>
            <th class="col-no" rowspan="2">No</th>
            <th class="col-nisn" rowspan="2">NISN</th>
            <th class="col-nama" rowspan="2">Nama Siswa</th>
            {#each Array(uhColumnCount) as _, i (i)}
              <th class="col-uh">UH {i + 1}</th>
            {/each}
            <th class="col-calc">Rata UH</th>
            <th class="col-pts" colspan="2">PTS</th>
            <th class="col-calc">NA PTS</th>
            <th class="col-input">PAS</th>
            <th class="col-calc col-nr">NR</th>
          </tr>
          <tr>
            {#each Array(uhColumnCount) as _, i (i)}
              <th class="col-uh-sub"></th>
            {/each}
            <th class="col-calc-sub"></th>
            <th class="col-pts-sub">M</th>
            <th class="col-pts-sub">R</th>
            <th class="col-calc-sub"></th>
            <th class="col-input-sub"></th>
            <th class="col-calc-sub col-nr"></th>
          </tr>
        </thead>
        <tbody onpaste={handlePaste}>
          {#each siswaList as siswa, idx (siswa.siswa_id)}
            {@const data = getSiswaData(siswa.siswa_id)}
            <tr class:saving={savingCell === String(siswa.siswa_id)} data-siswa-id={siswa.siswa_id}>
              <td class="text-center">{idx + 1}</td>
              <td class="text-center text-mono">{siswa.nisn}</td>
              <td class="col-nama-cell">{siswa.nama_lengkap}</td>
              {#each Array(uhColumnCount) as _, i (i)}
                <td>
                  <input
                    type="number"
                    class="input-cell"
                    min="0"
                    max="100"
                    data-field="uh"
                    data-idx={i}
                    data-cell={`${idx}-${i}`}
                    value={getUhScore(siswa.siswa_id, i) || ''}
                    onblur={(e) => {
                      updateUh(siswa.siswa_id, i, e.target.value);
                      saveNilai(siswa.siswa_id);
                    }}
                    onchange={(e) => updateUh(siswa.siswa_id, i, e.target.value)}
                    onkeydown={(e) => moveCell(e, idx, i)}
                  />
                </td>
              {/each}
              <td class="cell-calc">{formatDecimal(data.rata_uh)}</td>
              <td>
                <input
                  type="number"
                  class="input-cell"
                  min="0"
                  max="100"
                  data-field="murni_pts"
                  data-cell={`${idx}-${uhColumnCount}`}
                  value={data.murni_pts || ''}
                  onblur={(e) => {
                    updateField(siswa.siswa_id, 'murni_pts', e.target.value);
                    saveNilai(siswa.siswa_id);
                  }}
                  onchange={(e) => updateField(siswa.siswa_id, 'murni_pts', e.target.value)}
                  onkeydown={(e) => moveCell(e, idx, uhColumnCount)}
                />
              </td>
              <td>
                <input
                  type="number"
                  class="input-cell input-remidi"
                  min="0"
                  max="100"
                  data-field="remidi_pts"
                  data-cell={`${idx}-${uhColumnCount + 1}`}
                  value={data.remidi_pts ?? ''}
                  placeholder="-"
                  onblur={(e) => {
                    updateField(siswa.siswa_id, 'remidi_pts', e.target.value);
                    saveNilai(siswa.siswa_id);
                  }}
                  onchange={(e) => updateField(siswa.siswa_id, 'remidi_pts', e.target.value)}
                  onkeydown={(e) => moveCell(e, idx, uhColumnCount + 1)}
                />
              </td>
              <td class="cell-calc">{data.na_pts || '-'}</td>
              <td>
                <input
                  type="number"
                  class="input-cell"
                  min="0"
                  max="100"
                  data-field="pas"
                  data-cell={`${idx}-${uhColumnCount + 2}`}
                  value={data.pas || ''}
                  onblur={(e) => {
                    updateField(siswa.siswa_id, 'pas', e.target.value);
                    saveNilai(siswa.siswa_id);
                  }}
                  onchange={(e) => updateField(siswa.siswa_id, 'pas', e.target.value)}
                  onkeydown={(e) => moveCell(e, idx, uhColumnCount + 2)}
                />
              </td>
              <td class="cell-nr">{formatDecimal(data.nilai_rapor)}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    <div class="legend">
      <p>
        <strong>Keterangan:</strong> UH = Ulangan Harian, M = Murni PTS, R = Remidi PTS, NA PTS = Nilai
        Akhir PTS, PAS = Penilaian Akhir Semester, NR = Nilai Rapor
      </p>
      <p>
        <strong>Rumus:</strong> NR = (Rata UH × 40%) + (NA PTS × 30%) + (PAS × 30%) | KKM = {KKM}
      </p>
      <p>
        <strong>Autosave:</strong> Nilai otomatis tersimpan saat kursor keluar dari kolom input.
      </p>
    </div>
  {:else if selectedMapel}
    <div class="table-loading">Belum ada siswa di kelas ini.</div>
  {/if}
</div>

<style>
  @import url('https://fonts.googleapis.com/css2?family=Amiri:wght@400;700&display=swap');

  .page-container {
    display: flex;
    flex-direction: column;
    width: 100%;
    padding-bottom: 2rem;
  }

  .page-header {
    margin-bottom: 1.5rem;
  }

  .btn-back {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    background: none;
    border: none;
    color: #6366f1;
    font-weight: 600;
    cursor: pointer;
    font-size: 0.9rem;
    font-family: inherit;
    margin-bottom: 0.75rem;
    padding: 0;
  }
  .btn-back:hover {
    color: #4338ca;
  }

  .page-title {
    margin: 0;
    font-size: 1.5rem;
    color: #1e293b;
    font-weight: 700;
  }
  .page-subtitle {
    margin: 0.25rem 0 0;
    font-size: 0.9rem;
    color: #64748b;
  }

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 1rem;
    margin-bottom: 1rem;
    background: #fff;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    padding: 0.85rem 1rem;
  }

  .toolbar-left {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .toolbar-label {
    font-weight: 700;
    font-size: 0.9rem;
    color: #334155;
    white-space: nowrap;
  }

  .select-mapel {
    padding: 0.5rem 0.75rem;
    border: 1px solid #cbd5e1;
    border-radius: 8px;
    font-family: inherit;
    font-size: 0.9rem;
    min-width: 220px;
    background: #f8fafc;
  }

  .toolbar-right {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .btn-sm {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
    padding: 0.4rem 0.7rem;
    font-size: 0.82rem;
    font-weight: 600;
    border: 1px solid #cbd5e1;
    border-radius: 6px;
    background: #fff;
    cursor: pointer;
    font-family: inherit;
    color: #475569;
    transition: all 0.15s;
  }
  .btn-sm:hover {
    background: #f1f5f9;
    border-color: #94a3b8;
  }
  .btn-sm:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  .btn-primary {
    background: #6366f1;
    color: white;
    border: 1px solid #4f46e5;
  }
  .btn-primary:hover {
    background: #4f46e5;
    color: white;
  }

  .uh-info {
    font-size: 0.82rem;
    color: #64748b;
    font-weight: 600;
  }

  .table-wrapper {
    overflow-x: auto;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    background: #fff;
  }

  .nilai-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.82rem;
  }

  .nilai-table th,
  .nilai-table td {
    border: 1px solid #e2e8f0;
    padding: 0.4rem 0.5rem;
    text-align: center;
    vertical-align: middle;
  }

  .nilai-table thead th {
    background: #f1f5f9;
    font-weight: 700;
    color: #334155;
    position: sticky;
    top: 0;
    z-index: 1;
  }

  .col-no {
    width: 36px;
  }
  .col-nisn {
    width: 85px;
  }
  .col-nama {
    min-width: 160px;
    text-align: left;
  }
  .col-nama-cell {
    text-align: left;
    font-weight: 600;
    color: #1e293b;
    white-space: nowrap;
  }
  .col-uh,
  .col-uh-sub {
    width: 55px;
  }
  .col-calc,
  .col-calc-sub {
    width: 60px;
    background: #f8fafc;
  }
  .col-pts,
  .col-pts-sub {
    width: 55px;
  }
  .col-input,
  .col-input-sub {
    width: 55px;
  }
  .col-nr {
    background: #eff6ff !important;
  }

  .text-center {
    text-align: center;
  }
  .text-mono {
    font-family: monospace;
    font-size: 0.78rem;
  }
  .text-muted {
    color: #94a3b8;
  }

  .cell-calc {
    background: #f8fafc;
    color: #475569;
    font-weight: 700;
  }
  .cell-nr {
    background: #eff6ff;
    color: #1d4ed8;
    font-weight: 800;
    font-size: 0.88rem;
  }

  .input-cell {
    width: 100%;
    max-width: 55px;
    padding: 0.3rem 0.25rem;
    text-align: center;
    border: 1px solid transparent;
    border-radius: 4px;
    font-family: inherit;
    font-size: 0.82rem;
    background: transparent;
    transition: all 0.15s;
  }
  .input-cell:focus {
    outline: none;
    border-color: #6366f1;
    background: #fff;
    box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
  }
  .input-cell:hover {
    background: #f8fafc;
  }

  .input-remidi {
    color: #d97706;
  }

  tr.saving {
    background: #fefce8;
  }

  /* Remove number input arrows */
  .input-cell::-webkit-outer-spin-button,
  .input-cell::-webkit-inner-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }
  .input-cell {
    -moz-appearance: textfield;
  }

  .table-loading {
    color: #64748b;
    font-size: 0.9rem;
    font-weight: 700;
    border: 1px dashed #cbd5e1;
    border-radius: 8px;
    padding: 2rem 1rem;
    background: #f8fafc;
    text-align: center;
  }

  .legend {
    margin-top: 1rem;
    padding: 1rem;
    background: #f8fafc;
    border: 1px solid #e2e8f0;
    border-radius: 10px;
    font-size: 0.82rem;
    color: #475569;
    line-height: 1.7;
  }
  .legend p {
    margin: 0.25rem 0;
  }

  .alert-error {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
    padding: 0.75rem 1rem;
    border-radius: 8px;
    font-size: 0.875rem;
    font-weight: 700;
    background: #fef2f2;
    color: #dc2626;
    border: 1px solid #fecaca;
  }
  .btn-dismiss {
    background: none;
    border: none;
    color: #dc2626;
    cursor: pointer;
    font-size: 1rem;
  }

  @media (max-width: 720px) {
    .toolbar {
      flex-direction: column;
      align-items: stretch;
    }
    .toolbar-right {
      justify-content: flex-end;
    }
  }
</style>
