<script>
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import Icon from '$lib/components/ui/Icon.svelte';
  import { goto } from '$app/navigation';
  import { toastStore } from '$lib/toast.svelte.js';

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

  let kelasId = $derived($page.params.kelas_id);
  let kelasLabel = $derived($page.url.searchParams.get('kelas') || `Kelas #${kelasId}`);
  let rombel = $derived($page.url.searchParams.get('rombel') || '__kosong__');
  let rombelText = $derived(rombel === '__kosong__' ? 'Tanpa Rombel' : `Rombel ${rombel}`);
  let sekolahId = $derived($page.url.searchParams.get('sekolah_id') || '1');
  let paramTaId = $derived($page.url.searchParams.get('tahun_ajaran_id') || '');
  let paramSemId = $derived($page.url.searchParams.get('semester_id') || '');

  let loading = $state(false);
  let saving = $state(false);
  let error = $state('');

  let siswaData = $state([]);
  let searchQuery = $state('');
  let filterOnlyScored = $state(false);

  let filteredSiswa = $derived(
    siswaData
      .filter((s) => (s.nama || '').toLowerCase().includes(searchQuery.toLowerCase()))
      .filter((s) => (filterOnlyScored ? s.hasNilai : true))
  );

  onMount(loadData);

  async function loadData() {
    loading = true;
    error = '';
    try {
      const params = new URLSearchParams({
        mata_pelajaran: 'Tahsin',
        jenis_penilaian: 'semester'
      });
      if (sekolahId) params.append('sekolah_id', sekolahId);
      if (rombel && rombel !== '__kosong__') params.append('rombel', rombel);
      if (paramTaId) params.append('tahun_ajaran_id', paramTaId);
      if (paramSemId) params.append('semester_id', paramSemId);

      // 1. Ambil data rapor tahsin masal (V2 or V1)
      let res = await fetch(`${apiUrl}/api/guru/kelas/${kelasId}/tahsin-v2-masal?${params.toString()}`, {
        credentials: 'include'
      });
      if (!res.ok) {
        res = await fetch(`${apiUrl}/api/guru/kelas/${kelasId}/tahsin-masal?${params.toString()}`, {
          credentials: 'include'
        });
      }

      if (!res.ok) throw new Error('Gagal memuat data ujian tahsin siswa');
      const rawData = await res.json();

      // 2. Ambil catatan rapor yang sudah ada
      const paramsCatatan = new URLSearchParams({
        jenis_rapor: 'tahsin',
        sekolah_id: sekolahId
      });
      if (rombel && rombel !== '__kosong__') paramsCatatan.append('rombel', rombel);
      if (paramTaId) paramsCatatan.set('tahun_ajaran_id', paramTaId);
      if (paramSemId) paramsCatatan.set('semester_id', paramSemId);

      const resCatatan = await fetch(
        `${apiUrl}/api/guru/kelas/${kelasId}/catatan-rapor?${paramsCatatan.toString()}`,
        {
          credentials: 'include'
        }
      );

      let catMap = {};
      if (resCatatan.ok) {
        const body = await resCatatan.json();
        (body.catatan || []).forEach((c) => {
          catMap[c.siswa_id] = c.catatan;
        });
      }

      // 3. Process siswa
      siswaData = (rawData || [])
        .map((item) => {
          const s = item.siswa || {};
          const header = item.nilai;
          const hasNilai = !!(header && header.jilid && item.details && item.details.length > 0);
          const jilid = header?.jilid || 1;

          // Hitung rata-rata
          const allInd = item.indikator_list || [];
          const currentInd = allInd.filter((ind) => ind.jilid === jilid);
          const totalInd = currentInd.length;

          const detailMap = {};
          (item.details || []).forEach((d) => {
            detailMap[d.kode] = d.nilai;
          });

          const totalPoin = currentInd.reduce((sum, ind) => {
            const v = detailMap[ind.kode];
            return sum + (v !== undefined && v !== null ? Number(v) : 0);
          }, 0);

          let avg = 0;
          if (totalInd > 0) {
            avg = Number((totalPoin / totalInd).toFixed(2));
          } else {
            avg = header?.nilai_akhir || 0;
          }

          let predikat = '-';
          if (hasNilai && avg > 0) {
            if (avg <= 5) {
              if (avg >= 4.5) predikat = 'A (Sangat Baik)';
              else if (avg >= 3.5) predikat = 'B (Baik)';
              else if (avg >= 2.5) predikat = 'C (Cukup)';
              else if (avg >= 1.5) predikat = 'D (Kurang)';
              else predikat = 'E (Perlu Bimbingan)';
            } else {
              if (avg >= 90) predikat = 'A (Sangat Baik)';
              else if (avg >= 80) predikat = 'B (Baik)';
              else if (avg >= 70) predikat = 'C (Cukup)';
              else if (avg >= 60) predikat = 'D (Kurang)';
              else predikat = 'E (Perlu Bimbingan)';
            }
          }

          return {
            siswa_id: s.id,
            nama: s.nama_lengkap,
            nisn: s.nisn,
            jilid,
            hasNilai,
            rata_rata: avg,
            predikat,
            catatan: catMap[s.id] || item.catatan || header?.catatan || ''
          };
        })
        .sort((a, b) => (a.nama || '').localeCompare(b.nama || ''));
    } catch (err) {
      error = err.message || 'Gagal memuat data';
    } finally {
      loading = false;
    }
  }

  function generateAutoCatatan() {
    let count = 0;
    siswaData = siswaData.map((s) => {
      if (!s.hasNilai) return s;
      let text = '';
      const jilidStr = `Iqro Jilid ${s.jilid}`;
      const score = s.rata_rata;

      const isV1 = score <= 5;
      const isA = isV1 ? score >= 4.5 : score >= 90;
      const isB = isV1 ? score >= 3.5 : score >= 80;
      const isC = isV1 ? score >= 2.5 : score >= 70;
      const isD = isV1 ? score >= 1.5 : score >= 60;

      if (isA) {
        text = `Alhamdulillah, ananda telah menyelesaikan ujian ${jilidStr} dengan sangat baik, lancar, dan makhraj yang tepat. Siap melanjutkan ke jilid berikutnya.`;
      } else if (isB) {
        text = `Alhamdulillah, ananda telah menyelesaikan ujian ${jilidStr} dengan baik. Terus pertahankan dan rajin berlatih agar semakin fasih.`;
      } else if (isC) {
        text = `Ananda telah menyelesaikan ujian ${jilidStr} dengan cukup baik. Perlu memperbanyak latihan mandiri terutama pada panjang-pendek bacaan.`;
      } else if (isD) {
        text = `Ananda telah mengikuti ujian ${jilidStr}. Masih memerlukan bimbingan rutin guru dalam pengucapan makhraj dan kaidah tajwid.`;
      } else {
        text = `Ananda masih membutuhkan pendampingan dan bimbingan intensif pada materi ${jilidStr} sebelum melanjutkan ke tingkat berikutnya.`;
      }

      count++;
      return {
        ...s,
        catatan: text
      };
    });

    toastStore.show(
      `Berhasil men-generate template catatan otomatis untuk ${count} siswa yang telah dinilai.`,
      'success'
    );
  }

  async function simpanSemua() {
    saving = true;
    try {
      const payload = {
        tahun_ajaran_id: parseInt(paramTaId) || 0,
        semester_id: parseInt(paramSemId) || 0,
        sekolah_id: Number(sekolahId),
        jenis_rapor: 'tahsin',
        data: siswaData.map((s) => ({
          siswa_id: s.siswa_id,
          catatan: s.catatan || ''
        }))
      };

      const res = await fetch(`${apiUrl}/api/guru/catatan-rapor`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });

      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.message || 'Gagal menyimpan catatan masal');
      }

      toastStore.show('Catatan rapor ujian tahsin berhasil disimpan', 'success');
      setTimeout(() => {
        goto('/guru/rapor/tahsin');
      }, 1000);
    } catch (err) {
      toastStore.show(err.message, 'error');
    } finally {
      saving = false;
    }
  }
</script>

<svelte:head>
  <title>Catatan Masal Rapor Tahsin - {kelasLabel}</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <button class="btn-back" onclick={() => goto('/guru/rapor/tahsin')}>
      <Icon name="chevronLeft" size={16} /> Kembali
    </button>
    <div>
      <p class="eyebrow">Rapor Ujian Tahsin</p>
      <h1>Isi Catatan Evaluasi Masal</h1>
      <p class="subtitle">{kelasLabel} • {rombelText}</p>
    </div>
  </div>

  {#if error}
    <div class="alert-error">{error}</div>
  {/if}

  <div class="top-bar">
    <div class="search-bar">
      <Icon name="search" size={16} class="search-icon" />
      <input
        type="text"
        bind:value={searchQuery}
        placeholder="Cari nama siswa..."
        class="search-input"
      />
    </div>

    <div class="action-buttons">
      <label class="toggle-scored">
        <input type="checkbox" bind:checked={filterOnlyScored} />
        <span>Hanya yang sudah dinilai</span>
      </label>

      <button class="btn-auto" onclick={generateAutoCatatan} type="button">
        <Icon name="sparkles" size={16} /> Generate Catatan Otomatis
      </button>

      <button class="btn-save" onclick={simpanSemua} disabled={saving} type="button">
        <Icon name="check" size={16} />
        {saving ? 'Menyimpan...' : 'Simpan Semua'}
      </button>
    </div>
  </div>

  {#if loading}
    <div class="loading-box">Memuat data siswa...</div>
  {:else if siswaData.length === 0}
    <div class="empty-state">
      <Icon name="users" size={48} />
      <h3>Tidak ada siswa di rombel ini</h3>
    </div>
  {:else}
    <div class="table-container">
      <table class="data-table">
        <thead>
          <tr>
            <th class="w-12 text-center">No</th>
            <th>Siswa</th>
            <th class="w-32 text-center">Iqro</th>
            <th class="w-32 text-center">Nilai Rata-rata</th>
            <th class="w-40 text-center">Predikat</th>
            <th>Catatan / Evaluasi Guru Pembimbing</th>
          </tr>
        </thead>
        <tbody>
          {#each filteredSiswa as item, idx (item.siswa_id)}
            <tr class:row-unscored={!item.hasNilai}>
              <td class="text-center text-gray-500 font-semibold">{idx + 1}</td>
              <td>
                <div class="student-name">{item.nama}</div>
                <div class="student-nisn">NISN: {item.nisn || '-'}</div>
              </td>
              <td class="text-center">
                {#if item.hasNilai}
                  <span class="badge-jilid">Jilid {item.jilid}</span>
                {:else}
                  <span class="badge-empty">Belum Dinilai</span>
                {/if}
              </td>
              <td class="text-center">
                {#if item.hasNilai}
                  <span class="score-text">{item.rata_rata}</span>
                {:else}
                  <span class="text-gray-400">-</span>
                {/if}
              </td>
              <td class="text-center">
                {#if item.hasNilai}
                  <span class="predikat-badge">{item.predikat}</span>
                {:else}
                  <span class="text-gray-400">-</span>
                {/if}
              </td>
              <td>
                <textarea
                  bind:value={item.catatan}
                  placeholder={item.hasNilai
                    ? 'Tulis evaluasi kemajuan bacaan Iqro siswa...'
                    : 'Siswa belum memiliki nilai ujian tahsin'}
                  rows="2"
                  class="catatan-textarea"
                ></textarea>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<style>
  .page-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 1.5rem;
  }

  .page-header {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .btn-back {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    background: #f1f5f9;
    color: #475569;
    border: 1px solid #cbd5e1;
    border-radius: 8px;
    padding: 0.5rem 0.9rem;
    font-size: 0.85rem;
    font-weight: 600;
    cursor: pointer;
    margin-top: 0.25rem;
    transition: all 0.15s;
  }

  .btn-back:hover {
    background: #e2e8f0;
    color: #1e293b;
  }

  .eyebrow {
    margin: 0 0 0.25rem;
    color: #0b7c3d;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    font-size: 0.75rem;
  }

  .page-header h1 {
    margin: 0;
    font-size: 1.5rem;
    color: #0f172a;
    font-weight: 800;
  }

  .subtitle {
    margin: 0.25rem 0 0;
    color: #64748b;
    font-size: 0.9rem;
  }

  .alert-error {
    background: #fef2f2;
    border: 1px solid #fecaca;
    color: #dc2626;
    padding: 0.75rem 1rem;
    border-radius: 8px;
    margin-bottom: 1rem;
  }

  .top-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 1rem;
    margin-bottom: 1.25rem;
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

  .action-buttons {
    display: flex;
    align-items: center;
    gap: 0.75rem;
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

  .btn-auto {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    background: #f8fafc;
    color: #4f46e5;
    border: 1px solid #c7d2fe;
    border-radius: 8px;
    padding: 0.55rem 1rem;
    font-size: 0.85rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s;
  }

  .btn-auto:hover {
    background: #eef2ff;
    border-color: #818cf8;
  }

  .btn-save {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    background: #0b7c3d;
    color: white;
    border: none;
    border-radius: 8px;
    padding: 0.55rem 1.25rem;
    font-size: 0.85rem;
    font-weight: 700;
    cursor: pointer;
    box-shadow: 0 2px 4px rgba(11, 124, 61, 0.2);
    transition: all 0.15s;
  }

  .btn-save:hover {
    background: #096330;
  }

  .btn-save:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .table-container {
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    overflow: hidden;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  }

  .data-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.875rem;
  }

  .data-table th {
    background: #f8fafc;
    color: #475569;
    font-weight: 700;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid #e2e8f0;
    text-align: left;
  }

  .data-table td {
    padding: 0.75rem 1rem;
    border-bottom: 1px solid #f1f5f9;
    vertical-align: middle;
  }

  .row-unscored {
    background: #fafafa;
    opacity: 0.8;
  }

  .student-name {
    font-weight: 700;
    color: #1e293b;
  }

  .student-nisn {
    font-size: 0.75rem;
    color: #64748b;
  }

  .badge-jilid {
    display: inline-block;
    padding: 0.2rem 0.6rem;
    background: #dbeafe;
    color: #1d4ed8;
    border-radius: 999px;
    font-size: 0.75rem;
    font-weight: 700;
  }

  .badge-empty {
    display: inline-block;
    padding: 0.2rem 0.6rem;
    background: #f1f5f9;
    color: #64748b;
    border-radius: 999px;
    font-size: 0.75rem;
    font-weight: 600;
  }

  .score-text {
    font-weight: 800;
    font-size: 0.95rem;
    color: #0b7c3d;
  }

  .predikat-badge {
    font-weight: 600;
    font-size: 0.8rem;
    color: #334155;
  }

  .catatan-textarea {
    width: 100%;
    box-sizing: border-box;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    padding: 0.45rem 0.65rem;
    font-size: 0.825rem;
    line-height: 1.4;
    font-family: inherit;
    resize: vertical;
    outline: none;
  }

  .catatan-textarea:focus {
    border-color: #0b7c3d;
    box-shadow: 0 0 0 2px rgba(11, 124, 61, 0.15);
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
