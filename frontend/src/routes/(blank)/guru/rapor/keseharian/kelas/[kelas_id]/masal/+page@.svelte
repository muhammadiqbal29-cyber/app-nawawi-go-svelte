<script>
  import { toastStore } from '$lib/toast.svelte.js';
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import Icon from '$lib/components/ui/Icon.svelte';

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

  let kelasId = $derived($page.params.kelas_id);
  let rombel = $derived($page.url.searchParams.get('rombel') || '');
  let sekolahId = $derived($page.url.searchParams.get('sekolah_id') || '');

  let loading = $state(true);
  let error = $state('');
  let raporList = $state([]);

  let currentDate = new Date().toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric'
  });

  // Dynamic Print Settings
  let customTanggalRapor = $state(currentDate);
  let customKepalaAsrama = $state('AJI SYAFAAT');
  let customMusyrif = $state('');

  onMount(async () => {
    try {
      const params = new URLSearchParams();
      if (sekolahId) params.append('sekolah_id', sekolahId);
      if (rombel && rombel !== '__kosong__') params.append('rombel', rombel);

      const res = await fetch(
        `${apiUrl}/api/guru/kelas/${kelasId}/keseharian-masal?${params.toString()}`,
        {
          credentials: 'include'
        }
      );
      if (!res.ok) throw new Error('Gagal memuat rapor keseharian masal.');
      const data = await res.json();
      raporList = data || [];
      
      const paramsCatatan = new URLSearchParams({
        jenis_rapor: 'keseharian',
        sekolah_id: sekolahId,
        rombel: rombel && rombel !== '__kosong__' ? rombel : ''
      });
      const resCatatan = await fetch(`${apiUrl}/api/guru/kelas/${kelasId}/catatan-rapor?${paramsCatatan.toString()}`, {
        credentials: 'include'
      });
      if (resCatatan.ok) {
        const catBody = await resCatatan.json();
        const catList = catBody.catatan || [];
        const catMap = {};
        catList.forEach(c => {
          catMap[c.siswa_id] = c.catatan;
        });
        raporList = raporList.map(r => ({
          ...r,
          catatan: catMap[r.siswa.id] || ''
        }));
      }

      customTanggalRapor = formatTanggal();
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  });

  function formatTanggal() {
    const months = [
      'Januari',
      'Februari',
      'Maret',
      'April',
      'Mei',
      'Juni',
      'Juli',
      'Agustus',
      'September',
      'Oktober',
      'November',
      'Desember'
    ];
    const now = new Date();
    return `Palangka Raya, ${now.getDate()} ${months[now.getMonth()]} ${now.getFullYear()}`;
  }

  function getNilaiForIndikator(raporData, indId) {
    if (!raporData || !raporData.rekaps) return 100;
    const rekap = raporData.rekaps.find((r) => r.indikator_id === indId);
    return rekap ? rekap.poin_akhir : 100;
  }

  function formatDate(dateStr) {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    return date.toLocaleDateString('id-ID', { day: '2-digit', month: '2-digit', year: 'numeric' });
  }

  let isDownloading = $state(false);

  async function downloadSemuaPDF() {
    isDownloading = true;
    try {
      const html2pdf = (await import('html2pdf.js')).default;
      for (const rapor of raporList) {
        const elId = `rapor-container-${rapor.siswa.id}`;
        const element = document.getElementById(elId);
        if (element) {
          const opt = {
            margin: [10, 15],
            filename: `Rapor_Keseharian_${rapor.siswa.nama_lengkap.replace(/\s+/g, '_')}.pdf`,
            image: { type: 'jpeg', quality: 1 },
            html2canvas: { scale: 2, useCORS: true },
            jsPDF: { unit: 'mm', format: 'a4', orientation: 'portrait' }
          };

          await html2pdf().set(opt).from(element).save();
          await new Promise((r) => setTimeout(r, 800));
        }
      }
      toastStore.show('Berhasil menyimpan semua file PDF per siswa!', 'info');
    } catch (err) {
      console.error(err);
      toastStore.show('Terjadi kesalahan saat download PDF: ' + err.message, 'info');
    } finally {
      isDownloading = false;
    }
  }

  function getRiwayatGabungan(rd) {
    if (!rd) return [];
    const p = (rd.pelanggaran || []).map(x => ({ ...x, tipe: 'Pelanggaran' }));
    const iq = (rd.iqob || []).map(x => ({ ...x, tipe: 'Iqob' }));
    return [...p, ...iq].sort((a, b) => new Date(a.tanggal) - new Date(b.tanggal));
  }

  function getRiwayatSideBySide(rd) {
    if (!rd) return [];
    const p = rd.pelanggaran || [];
    const iq = rd.iqob || [];
    const maxLen = Math.max(p.length, iq.length);
    const result = [];
    for (let i = 0; i < maxLen; i++) {
      result.push({
        pelanggaran: p[i] || null,
        iqob: iq[i] || null
      });
    }
    return result;
  }

</script>

<svelte:head>
  <title>Cetak Rapor Keseharian Masal - SIAKAD</title>
</svelte:head>

<div class="page-wrapper">
  {#if loading}
    <div class="loading-box">Memuat rapor keseharian...</div>
  {:else if error}
    <div class="error-box">{error}</div>
  {:else if raporList.length > 0}
    <div class="print-actions no-print">
      <button class="btn-back" onclick={() => goto('/guru/rapor/keseharian')}>
        <Icon name="chevronLeft" size={16} /> Kembali
      </button>
      <button class="btn-print" onclick={() => window.print()}> 🖨️ Cetak Rapor (Print) </button>
      <button class="btn-download" onclick={downloadSemuaPDF} disabled={isDownloading}>
        {isDownloading ? 'Memproses...' : '⬇️ Download PDF (Per Siswa)'}
      </button>
    </div>

    <!-- Edit Settings -->
    <div class="settings-box no-print">
      <h4 style="margin-top: 0;">Pengaturan Cetak (Tidak ikut terprint)</h4>
      <div class="settings-grid">
        <label>
          Tanggal Rapor:
          <input type="text" class="form-input" bind:value={customTanggalRapor} />
        </label>
        <label>
          Kepala Asrama:
          <input type="text" class="form-input" bind:value={customKepalaAsrama} />
        </label>
        <label>
          Musyrif/Musyrifah:
          <input
            type="text"
            class="form-input"
            bind:value={customMusyrif}
            placeholder="Nama Musyrif"
          />
        </label>
      </div>
    </div>

    <!-- Render all students -->
    {#each raporList as raporData}
      <!-- A4 Paper Container -->
      <div class="a4-container" id="rapor-container-{raporData.siswa.id}">
        <!-- Header -->
        <div class="header-section">
          <h1 class="header-sekolah">
            {raporData.sekolah?.nama_sekolah || 'SMP ISLAM IMAM NAWAWI'}
          </h1>
          <p class="header-address">
            {raporData.sekolah?.alamat ||
              'Jl. Tjilik Riwut KM 8 (Jl. Gajah Mada) Kel. Bukit Tunggal Kec. Jekan Raya Kota Palangka Raya'}
          </p>
          <p class="header-contact">Email: smpislamnawawi@gmail.com Telp. 0898 1567 898</p>
        </div>

        <!-- Double Border Line -->
        <div class="header-line"></div>

        <!-- Title -->
        <div class="title-section">
          <h3 class="title-main">RAPOR KESEHARIAN & POIN SANTRI</h3>
        </div>

        <!-- Student Info -->
        <div class="student-info">
          <table class="info-table">
            <tbody>
              <tr>
                <td class="info-label">NAMA SISWA</td>
                <td class="info-sep">:</td>
                <td>{raporData.siswa?.nama_lengkap || '-'}</td>
              </tr>
              <tr>
                <td>KAMAR</td>
                <td class="info-sep">:</td>
                <td>{raporData.riwayat?.kamar || '-'}</td>
              </tr>
              <tr>
                <td>MUSYRIF/MUSYRIFAH</td>
                <td class="info-sep">:</td>
                <td>{customMusyrif || '-'}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Bagian A: RAPOR KESEHARIAN -->
        <div style="margin-bottom: 0.5rem;">
          <h4 style="font-weight: 700; font-size: 1rem; margin-bottom: 0.5rem;">
            A. RAPOR KESEHARIAN
          </h4>
          <table class="rapor-table">
            <thead>
              <tr>
                <th style="width: 3rem; text-align: center;">NO</th>
                <th>KEGIATAN</th>
                <th style="width: 6rem; text-align: center;">NILAI</th>
              </tr>
            </thead>
            <tbody>
              {#each raporData.indikators || [] as ind, i (ind.id)}
                <tr>
                  <td class="text-center">{i + 1}</td>
                  <td>{ind.nama}</td>
                  <td class="text-center">{getNilaiForIndikator(raporData, ind.id)}</td>
                </tr>
              {/each}
              <!-- Rata-rata -->
              <tr class="group-header">
                <td colspan="2" class="text-right" style="padding-right: 1rem;">NILAI RATA-RATA</td>
                <td class="text-center">
                  {raporData.indikators && raporData.indikators.length > 0
                    ? (
                        raporData.indikators.reduce(
                          (acc, ind) => acc + getNilaiForIndikator(raporData, ind.id),
                          0
                        ) / raporData.indikators.length
                      ).toFixed(1)
                    : 100}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Bagian B: RAPOR PELANGGARAN & IQOB -->
      <div class="mb-2 mt-4">
        <h4 class="font-bold text-base mb-2">B. RAPOR PELANGGARAN & IQOB</h4>
        <table class="rapor-table">
          <thead>
            <tr>
              <th class="w-10 text-center">NO</th>
              <th class="w-24 text-center">TANGGAL</th>
              <th>KETERANGAN PELANGGARAN</th>
              <th class="w-20 text-center">POIN PELANGGARAN</th>
              <th>KETERANGAN IQOB</th>
              <th class="w-20 text-center">POIN IQOB</th>
            </tr>
          </thead>
          <tbody>
            {#if true}
              {@const riwayat = getRiwayatSideBySide(raporData)}
              {#if riwayat.length > 0}
                {#each riwayat as r, i (i)}
                  <tr style="color: black;">
                    <td class="text-center">{i + 1}</td>
                    <td class="text-center">
                      {r.pelanggaran ? formatDate(r.pelanggaran.tanggal) : ''}
                    </td>
                    <td>{r.pelanggaran ? r.pelanggaran.pelanggaran : ''}</td>
                    <td class="text-center">{r.pelanggaran ? r.pelanggaran.poin : ''}</td>
                    <td>
                      {#if r.iqob}
                        {r.iqob.nama_iqob}
                      {/if}
                    </td>
                    <td class="text-center">{r.iqob ? r.iqob.poin : ''}</td>
                  </tr>
                {/each}
              {:else}
                <tr>
                  <td colspan="6" class="text-center py-4 text-gray-500 italic">
                    Tidak ada catatan pelanggaran atau iqob.
                  </td>
                </tr>
              {/if}
            {/if}
          </tbody>
        </table>

        <!-- Ringkasan Poin Akhir -->
        <div class="font-bold mt-4 ml-2 border p-4 bg-gray-50">
          {#if true}
            {@const tPlg = (raporData.pelanggaran || []).reduce((acc, p) => acc + (p.poin || 0), 0)}
            {@const tIqob = (raporData.iqob || []).reduce((acc, p) => acc + (p.poin || 0), 0)}
            {@const tSisa = tPlg > tIqob ? tPlg - tIqob : 0}
            <table class="border-none w-auto text-sm text-black" style="border:none; color: black;">
              <tbody>
                <tr>
                  <td class="border-none py-1 px-0 w-48" style="border:none">TOTAL POIN PELANGGARAN</td>
                  <td class="info-sep" style="padding:0 8px;">=</td>
                  <td>{tPlg}</td>
                </tr>
                <tr>
                  <td style="border:none">TOTAL PENGURANGAN IQOB</td>
                  <td class="info-sep" style="border:none; padding:0 8px;">=</td>
                  <td class="border-none py-1 px-0" style="border:none">{tIqob}</td>
                </tr>
                <tr style="border-top: 1px solid #ccc;">
                  <td style="padding-top: 8px;">SISA POIN PELANGGARAN</td>
                  <td class="info-sep" style="padding-top: 8px; padding-left: 8px; padding-right: 8px;">=</td>
                  <td class="border-none py-1 px-0" style="border:none; padding-top: 8px;">{tSisa}</td>
                </tr>
                <tr>
                  <td style="font-size: 1.1em;">NILAI KEDISIPLINAN AKHIR</td>
                  <td class="info-sep" style="padding:0 8px; font-size: 1.1em;">=</td>
                  <td class="border-none py-1 px-0 font-bold" style="border:none; font-size: 1.1em;">{100 - tSisa}</td>
                </tr>
              </tbody>
            </table>
          {/if}
        </div>
      </div>
      
      <!-- Catatan Keseharian -->
      <div class="catatan-box mt-4 border border-black p-2" style="min-height: 80px;">
        <p class="font-bold text-sm m-0 mb-1">CATATAN:</p>
        <p class="text-sm m-0 whitespace-pre-wrap">{raporData.catatan || '-'}</p>
      </div>

      <!-- Signatures -->
        <div class="signature-section">
          <div class="sig-col">
            <p class="sig-invisible">{customTanggalRapor}</p>
            <!-- Spacer -->
            <p>Kepala Asrama</p>
            <div class="sig-space"></div>
            <!-- Space for signature -->
            <p class="sig-name">{customKepalaAsrama || '____________________'}</p>
          </div>
          <div class="sig-col">
            <p>{customTanggalRapor}</p>
            <p>Musyrif/Musyrifah Asrama</p>
            <div class="sig-space"></div>
            <!-- Space for signature -->
            <p class="sig-name">{customMusyrif || '_________________________'}</p>
          </div>
        </div>
      </div>

      <div class="page-break" style="page-break-after: always;"></div>
    {/each}
  {:else}
    <div class="error-box">Data rapor masal kosong.</div>
  {/if}
</div>

<style>
  @import url('https://fonts.googleapis.com/css2?family=Amiri:wght@400;700&display=swap');

  :global(body) {
    background-color: #f3f4f6;
  }

  .arabic {
    font-family: 'Amiri', 'Traditional Arabic', serif;
    font-size: 18px !important;
    line-height: 1.4 !important;
  }

  @media print {
    :global(body) {
      background-color: white !important;
      margin: 0;
      padding: 0;
      -webkit-print-color-adjust: exact;
      print-color-adjust: exact;
    }
    .no-print {
      display: none !important;
    }
    .page-wrapper {
      padding: 0 !important;
      margin: 0 !important;
    }
    .a4-container {
      box-shadow: none !important;
      margin: 0 !important;
      padding: 1cm 2cm 2cm 2cm !important;
      width: 100% !important;
      max-width: none !important;
    }
    @page {
      size: A4 portrait;
      margin: 0;
    }
  }

  .page-wrapper {
    padding: 2rem 0;
  }

  .print-actions {
    display: flex;
    justify-content: center;
    gap: 1rem;
    align-items: center;
    max-width: 210mm;
    margin: 0 auto 1.5rem;
  }

  .settings-box {
    max-width: 210mm;
    margin: 0 auto 1.5rem;
    background: white;
    padding: 1rem;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
    border: 1px solid #e2e8f0;
  }
  .settings-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
  }
  .settings-grid label {
    display: flex;
    flex-direction: column;
    font-size: 0.85rem;
    font-weight: 600;
    color: #475569;
    gap: 0.25rem;
  }
  .form-input {
    padding: 0.5rem;
    border: 1px solid #cbd5e1;
    border-radius: 6px;
    font-family: inherit;
  }

  .btn-back {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    background: #4b5563;
    color: white;
    border: none;
    border-radius: 8px;
    padding: 0.6rem 1.5rem;
    font-weight: 700;
    font-size: 0.95rem;
    cursor: pointer;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    transition: all 0.2s;
  }
  .btn-back:hover {
    background: #374151;
  }

  .btn-print {
    background: #2563eb;
    color: white;
    border: none;
    border-radius: 8px;
    padding: 0.6rem 1.5rem;
    font-weight: 700;
    font-size: 0.95rem;
    cursor: pointer;
    box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
    transition: all 0.2s;
  }
  .btn-print:hover {
    background: #1d4ed8;
  }

  .btn-download {
    background: #10b981;
    color: white;
    border: none;
    border-radius: 8px;
    padding: 0.6rem 1.5rem;
    font-weight: 700;
    font-size: 0.95rem;
    cursor: pointer;
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
    transition: all 0.2s;
  }
  .btn-download:hover:not(:disabled) {
    background: #059669;
  }
  .btn-download:disabled {
    background: #94a3b8;
    cursor: not-allowed;
    box-shadow: none;
  }

  .a4-container {
    background: white;
    width: 100%;
    max-width: 210mm;
    min-height: auto !important;
    height: auto !important;
    margin: 0 auto;
    padding: 1cm 2cm 2cm 2cm !important;
    box-shadow: 0 8px 30px rgba(0, 0, 0, 0.12);
    font-family: sans-serif;
    font-size: 14px;
    color: black;
  }

  .header-section {
    text-align: center;
    margin-bottom: 16px;
  }
  .header-sekolah {
    font-size: 24px;
    font-weight: 700;
    letter-spacing: 3px;
    margin-top: 4px;
    margin-bottom: 8px;
  }
  .header-address,
  .header-contact {
    font-size: 12px;
    margin: 4px 0 0;
  }

  .header-line {
    border-top: 3px double black;
    margin-bottom: 24px;
  }

  .title-section {
    text-align: center;
    margin-bottom: 24px;
  }
  .title-main {
    font-weight: 700;
    font-size: 16px;
    text-decoration: underline;
    text-underline-offset: 4px;
    margin: 0;
  }

  .student-info {
    display: flex;
    justify-content: space-between;
    margin-bottom: 16px;
    font-size: 14px;
    font-weight: 600;
  }

  .info-table {
    border: none;
    width: auto;
  }
  .info-table td {
    border: none;
    padding: 4px 0;
    vertical-align: top;
  }
  .info-label {
    width: 120px;
  }
  .info-sep {
    padding: 4px 8px;
  }

  .rapor-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 12px;
  }

  .rapor-table th,
  .rapor-table td {
    border: 1px solid black;
    padding: 4px 8px;
    vertical-align: middle;
  }

  .rapor-table th {
    font-weight: 700;
    text-align: center;
  }

  .text-center {
    text-align: center;
  }
  .text-right {
    text-align: right;
  }

  .group-header {
    font-weight: 700;
    background: #f9fafb;
  }
  @media print {
    .group-header {
      background: transparent;
    }
  }

  .signature-section {
    display: flex;
    justify-content: space-between;
    margin-top: 32px;
    text-align: center;
    font-size: 14px;
    padding: 0 32px;
  }

  .sig-col {
    display: flex;
    flex-direction: column;
    align-items: center;
  }
  .sig-invisible {
    visibility: hidden;
  }
  .sig-space {
    height: 80px;
  }
  .sig-name {
    font-weight: 700;
    text-decoration: underline;
    margin: 0;
  }

  .loading-box,
  .error-box {
    padding: 3rem;
    text-align: center;
    border-radius: 8px;
    max-width: 400px;
    margin: 100px auto;
  }
  .loading-box {
    color: #64748b;
    border: 1px dashed #cbd5e1;
    background: #f8fafc;
  }
  .error-box {
    color: #dc2626;
    background: #fef2f2;
    border: 1px solid #fecaca;
  }
</style>
