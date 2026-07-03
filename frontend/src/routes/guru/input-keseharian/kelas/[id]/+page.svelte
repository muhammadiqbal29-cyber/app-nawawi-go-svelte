<script>
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import Icon from '$lib/components/ui/Icon.svelte';
  import { guruState } from '$lib/guruState.svelte.js';

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

  let kelasId = $derived($page.params.id);
  let sekolahId = $derived(guruState.activeSekolahId);
  let namaKelas = $derived($page.url.searchParams.get('kelas') || 'Kelas');
  let rombel = $derived($page.url.searchParams.get('rombel') || '');
  let displayKelasName = $derived(
    rombel && rombel !== '__kosong__' ? `${namaKelas} - Rombel ${rombel}` : namaKelas
  );

  let indikatorList = $state([]);
  let selectedIndikator = $state('');

  let siswaData = $state([]);
  let rekapData = $state({});
  let loading = $state(false);
  let error = $state('');
  let saveStatus = $state('');

  $effect(() => {
    if (sekolahId) {
      loadIndikator();
    }
  });

  $effect(() => {
    if (sekolahId && kelasId && selectedIndikator) {
      fetchRekap();
    }
  });

  async function loadIndikator() {
    try {
      const res = await fetch(`${apiUrl}/api/guru/sekolah/${sekolahId}/keseharian/indikator`, {
        credentials: 'include'
      });
      if (res.ok) {
        const body = await res.json();
        indikatorList = body.data || [];
      }
    } catch (err) {
      console.error(err);
    }
  }

  async function fetchRekap() {
    if (!kelasId || !selectedIndikator) return;

    loading = true;
    error = '';
    saveStatus = '';
    siswaData = [];
    rekapData = {};

    try {
      const url = new URL(`${apiUrl}/api/guru/kelas/${kelasId}/keseharian/rekap`);
      url.searchParams.append('sekolah_id', sekolahId);
      url.searchParams.append('indikator_id', selectedIndikator);

      const res = await fetch(url.toString(), { credentials: 'include' });
      if (!res.ok) throw new Error('Gagal memuat rekap keseharian.');

      const body = await res.json();
      siswaData = body.data?.siswa || [];
      const rekaps = body.data?.rekap || [];

      // Map rekaps to dict by siswa_id
      const map = {};
      for (const r of rekaps) {
        map[r.siswa_id] = {
          id: r.id,
          total_sesi: r.total_sesi,
          hadir_sesi: r.hadir_sesi,
          poin_murni: r.poin_murni,
          poin_penyesuaian: r.poin_penyesuaian,
          poin_akhir: r.poin_akhir,
          saving: false
        };
      }
      rekapData = map;
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function handlePenyesuaian(siswaId, newPoin) {
    const r = rekapData[siswaId];
    if (!r || !r.id) return; // Cannot adjust if no data yet

    r.saving = true;
    try {
      const res = await fetch(`${apiUrl}/api/guru/keseharian/penyesuaian/${r.id}`, {
        method: 'PUT',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ poin_penyesuaian: Number(newPoin) })
      });
      if (!res.ok) throw new Error('Gagal simpan');

      const body = await res.json();
      const updated = body.data;
      rekapData[siswaId].poin_penyesuaian = updated.poin_penyesuaian;
      rekapData[siswaId].poin_akhir = updated.poin_akhir;

      saveStatus = 'Tersimpan otomatis!';
      setTimeout(() => (saveStatus = ''), 2000);
    } catch (err) {
      alert(err.message);
    } finally {
      r.saving = false;
    }
  }
</script>

<svelte:head>
  <title>Finalisasi Keseharian - SIAKAD</title>
</svelte:head>

<div class="page-container">
  <button class="btn-back" onclick={() => goto('/guru/input-keseharian')}>
    <Icon name="chevronLeft" size={18} /> Kembali ke Daftar Kelas
  </button>

  <div class="header-action" style="margin-top: 1.5rem;">
    <div>
      <h2>Input Keseharian: {displayKelasName}</h2>
      <p class="subtitle">
        Lihat rata-rata kehadiran/ceklis keseharian dan berikan nilai penyesuaian akhir.
      </p>
    </div>
  </div>

  <div class="filter-card">
    <div class="form-group">
      <label for="indikator">Pilih Indikator Keseharian</label>
      <select id="indikator" class="input" bind:value={selectedIndikator}>
        <option value="">-- Pilih Indikator --</option>
        {#each indikatorList as ind}
          <option value={ind.id}>{ind.nama}</option>
        {/each}
      </select>
    </div>
  </div>

  {#if loading}
    <div class="loading-state">Memuat data rekap...</div>
  {:else if error}
    <div class="alert alert-error">{error}</div>
  {:else if siswaData.length > 0}
    <div class="table-container">
      {#if saveStatus}
        <div class="toast-save">{saveStatus}</div>
      {/if}
      <table class="data-table">
        <thead>
          <tr>
            <th width="5%">No</th>
            <th width="30%">Nama Siswa</th>
            <th width="15%">Total Ceklis (Sistem)</th>
            <th width="15%">Poin Murni (Otomatis)</th>
            <th width="15%">Penyesuaian (± Poin)</th>
            <th width="15%">Nilai Akhir (Rapor)</th>
          </tr>
        </thead>
        <tbody>
          {#each siswaData as s, i}
            {@const r = rekapData[s.siswa_id]}
            <tr>
              <td>{i + 1}</td>
              <td class="font-medium"
                >{s.nama_lengkap}
                <div class="text-sm text-gray-500">{s.nisn}</div>
              </td>
              <td>
                {#if r}
                  <span class="badge-blue">{r.hadir_sesi} / {r.total_sesi} kali</span>
                {:else}
                  <span class="badge-gray">Belum ada data</span>
                {/if}
              </td>
              <td>
                {#if r}
                  <span class="font-semibold">{Number(r.poin_murni).toFixed(1)}</span>
                {:else}
                  -
                {/if}
              </td>
              <td>
                {#if r}
                  <input
                    type="number"
                    class="input input-sm text-center"
                    value={r.poin_penyesuaian}
                    onchange={(e) => handlePenyesuaian(s.siswa_id, e.target.value)}
                    disabled={r.saving}
                    placeholder="Contoh: 10 atau -5"
                  />
                {:else}
                  <input type="text" class="input input-sm" disabled value="Kosong" />
                {/if}
              </td>
              <td>
                {#if r}
                  <div class="final-score" class:saving={r.saving}>
                    {Number(r.poin_akhir).toFixed(1)}
                  </div>
                {:else}
                  -
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {:else if selectedIndikator}
    <div class="empty-state">Tidak ada data siswa ditemukan untuk kombinasi ini.</div>
  {/if}
</div>

<style>
  .page-container {
    padding: 1rem 0;
  }
  .header-action {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 2rem;
  }
  .header-action h2 {
    margin: 0;
    color: #1e293b;
    font-size: 1.5rem;
  }
  .subtitle {
    margin: 0.25rem 0 0;
    color: #64748b;
    font-size: 0.9rem;
  }

  .filter-card {
    background: white;
    padding: 1.5rem;
    border-radius: 12px;
    border: 1px solid #e2e8f0;
    display: flex;
    gap: 1.5rem;
    margin-bottom: 2rem;
  }
  .form-group {
    flex: 1;
    max-width: 300px;
  }
  .form-group label {
    display: block;
    font-size: 0.85rem;
    font-weight: 600;
    color: #475569;
    margin-bottom: 0.5rem;
  }
  .input {
    width: 100%;
    padding: 0.6rem 0.75rem;
    border: 1px solid #cbd5e1;
    border-radius: 8px;
    background: #f8fafc;
    font-family: inherit;
    font-size: 0.9rem;
    color: #1e293b;
  }
  .input:focus {
    outline: none;
    border-color: #6366f1;
    background: white;
    box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
  }
  .input-sm {
    padding: 0.4rem 0.5rem;
    font-size: 0.85rem;
    border-radius: 6px;
  }

  .table-container {
    background: white;
    border-radius: 12px;
    border: 1px solid #e2e8f0;
    overflow: hidden;
    position: relative;
  }
  .data-table {
    width: 100%;
    border-collapse: collapse;
  }
  .data-table th {
    background: #f8fafc;
    text-align: left;
    padding: 0.75rem 1rem;
    font-size: 0.85rem;
    font-weight: 600;
    color: #475569;
    border-bottom: 1px solid #e2e8f0;
  }
  .data-table td {
    padding: 0.75rem 1rem;
    font-size: 0.9rem;
    color: #1e293b;
    border-bottom: 1px solid #f1f5f9;
    vertical-align: middle;
  }

  .font-medium {
    font-weight: 500;
  }
  .font-semibold {
    font-weight: 600;
    color: #6366f1;
  }
  .text-sm {
    font-size: 0.8rem;
  }
  .text-gray-500 {
    color: #64748b;
  }

  .badge-blue {
    background: #e0e7ff;
    color: #4338ca;
    padding: 0.2rem 0.5rem;
    border-radius: 4px;
    font-size: 0.8rem;
    font-weight: 600;
  }
  .badge-gray {
    background: #f1f5f9;
    color: #94a3b8;
    padding: 0.2rem 0.5rem;
    border-radius: 4px;
    font-size: 0.8rem;
  }

  .final-score {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    background: #10b981;
    color: white;
    font-weight: 700;
    border-radius: 50%;
    font-size: 0.9rem;
    transition: opacity 0.2s;
  }
  .final-score.saving {
    opacity: 0.5;
  }

  .toast-save {
    position: absolute;
    top: -40px;
    right: 0;
    background: #10b981;
    color: white;
    padding: 0.5rem 1rem;
    border-radius: 8px;
    font-size: 0.85rem;
    font-weight: 600;
    animation:
      fadein 0.3s,
      fadeout 0.3s 1.7s;
  }

  @keyframes fadein {
    from {
      opacity: 0;
      transform: translateY(10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
  @keyframes fadeout {
    from {
      opacity: 1;
    }
    to {
      opacity: 0;
    }
  }

  .btn-back {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 1rem;
    border-radius: 8px;
    background: white;
    border: 1px solid #e2e8f0;
    color: #475569;
    font-size: 0.9rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }
  .btn-back:hover {
    background: #f8fafc;
    color: #1e293b;
    border-color: #cbd5e1;
  }
</style>
