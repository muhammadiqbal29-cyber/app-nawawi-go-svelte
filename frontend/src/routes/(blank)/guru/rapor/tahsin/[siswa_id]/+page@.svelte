<script>
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import Icon from '$lib/components/ui/Icon.svelte';

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

  let siswaId = $derived($page.params.siswa_id);
  let sekolahId = $derived($page.url.searchParams.get('sekolah_id') || '');
  let paramTaId = $derived($page.url.searchParams.get('tahun_ajaran_id') || '');
  let paramSemId = $derived($page.url.searchParams.get('semester_id') || '');

  let loading = $state(true);
  let error = $state('');

  let siswa = $state(null);
  let sekolah = $state(null);
  let riwayat = $state(null);
  let kelasObj = $state(null);
  let semester = $state(null);
  let tahunAjaran = $state(null);
  let nilaiHeader = $state(null);
  let details = $state([]);
  let allIndikators = $state([]);
  let jilidList = $state([]);
  let catatan = $state('');

  let currentDate = new Date().toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric'
  });

  // Dynamic Print Settings
  let customKota = $state('Palangka Raya');
  let customTanggalRapor = $state(currentDate);
  let customKepalaSekolah = $state('M. HUSAINI, S.Pd.I.');
  let customGuruPenguji = $state('');

  // Kop & Logo
  let customLogoUrl = $state('');
  let logoX = $state(600);
  let logoY = $state(20);
  let logoSize = $state(80);
  let kopLines = $state([]);
  let dividerThickness = $state(3);

  // Local Signature Settings
  let ttdKepala = $state('');
  let ttdPenguji = $state('');
  let ttdKepalaSize = $state(100);
  let ttdPengujiSize = $state(100);

  onMount(() => {
    ttdKepala = localStorage.getItem('ttdKepala_global') || '';
    ttdPenguji = localStorage.getItem('ttdPenguji_tahsin_global') || localStorage.getItem('ttdMusyrif_global') || '';
    ttdKepalaSize = parseInt(localStorage.getItem('ttdKepalaSize_global')) || 100;
    ttdPengujiSize = parseInt(localStorage.getItem('ttdPengujiSize_tahsin_global')) || 100;
  });

  $effect(() => {
    if (ttdKepalaSize !== undefined) localStorage.setItem('ttdKepalaSize_global', ttdKepalaSize.toString());
    if (ttdPengujiSize !== undefined) localStorage.setItem('ttdPengujiSize_tahsin_global', ttdPengujiSize.toString());
  });

  function handleTtdKepala(e) {
    const file = e.target.files[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (event) => {
      ttdKepala = event.target.result;
      localStorage.setItem('ttdKepala_global', ttdKepala);
    };
    reader.readAsDataURL(file);
  }

  function handleTtdPenguji(e) {
    const file = e.target.files[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (event) => {
      ttdPenguji = event.target.result;
      localStorage.setItem('ttdPenguji_tahsin_global', ttdPenguji);
    };
    reader.readAsDataURL(file);
  }

  $effect(() => {
    if (siswaId) {
      loadData();
    }
  });

  async function loadData() {
    loading = true;
    error = '';
    try {
      const params = new URLSearchParams({
        mata_pelajaran: 'Tahsin',
        jenis_penilaian: 'semester'
      });
      if (sekolahId) params.append('sekolah_id', sekolahId);
      if (paramTaId) params.append('tahun_ajaran_id', paramTaId);
      if (paramSemId) params.append('semester_id', paramSemId);

      // Coba fetch Tahsin V2 terlebih dahulu
      let res = await fetch(`${apiUrl}/api/guru/siswa/${siswaId}/nilai-tahsin-v2?${params.toString()}`, {
        credentials: 'include'
      });

      // Jika belum ada di V2, fallback ke V1
      if (!res.ok) {
        res = await fetch(`${apiUrl}/api/guru/siswa/${siswaId}/nilai-tahsin?${params.toString()}`, {
          credentials: 'include'
        });
      }

      if (!res.ok) {
        throw new Error('Gagal memuat data rapor tahsin siswa');
      }

      const body = await res.json();
      siswa = body.siswa || null;
      sekolah = body.sekolah || null;
      riwayat = body.riwayat || null;
      kelasObj = body.kelas || null;
      semester = body.semester || null;
      tahunAjaran = body.tahun_ajaran || null;
      nilaiHeader = body.nilai || null;
      details = body.details || [];
      allIndikators = body.indikator_list || [];
      jilidList = body.jilid_list || [];

      // Setup Kop & Logo dari Sekolah
      if (sekolah) {
        if (sekolah.kepala_sekolah) {
          customKepalaSekolah = sekolah.kepala_sekolah;
        }
        customLogoUrl = sekolah.logo_base64 || '';
        logoX = sekolah.logo_x || 600;
        logoY = sekolah.logo_y || 20;
        logoSize = sekolah.logo_size || 80;
        dividerThickness = sekolah.divider_thickness || 3;
        let localKop = [];
        if (sekolah.kop_surat) {
          try {
            const parsed = JSON.parse(sekolah.kop_surat);
            if (Array.isArray(parsed) && parsed.length > 0) {
              localKop = parsed;
            }
          } catch (e) {}
        }
        if (localKop.length === 0) {
          localKop = [
            {
              text: sekolah.nama_sekolah || 'SMP ISLAM IMAM NAWAWI',
              color: '#000000',
              size: 18,
              bold: true
            },
            {
              text: sekolah.alamat || 'Jl. Tjilik Riwut KM 8, Palangka Raya',
              color: '#000000',
              size: 12,
              bold: false
            }
          ];
        }
        kopLines = localKop;
      }

      // Load Catatan Rapor khusus Tahsin
      const paramsCatatan = new URLSearchParams({
        jenis_rapor: 'tahsin'
      });
      if (sekolahId) paramsCatatan.append('sekolah_id', sekolahId);
      if (paramTaId) paramsCatatan.append('tahun_ajaran_id', paramTaId);
      if (paramSemId) paramsCatatan.append('semester_id', paramSemId);

      const resCatatan = await fetch(`${apiUrl}/api/guru/siswa/${siswaId}/catatan-rapor?${paramsCatatan.toString()}`, {
        credentials: 'include'
      });
      if (resCatatan.ok) {
        const catBody = await resCatatan.json();
        catatan = catBody.catatan || (nilaiHeader?.catatan || '');
      } else {
        catatan = nilaiHeader?.catatan || '';
      }
    } catch (err) {
      error = err.message || 'Gagal memuat data rapor';
    } finally {
      loading = false;
    }
  }

  // Active Iqro Jilid
  let activeJilid = $derived(nilaiHeader?.jilid || 1);

  // Indikator-indikator jilid aktif
  let currentIndikators = $derived(
    allIndikators
      .filter((ind) => ind.jilid === activeJilid)
      .sort((a, b) => (a.urutan || 0) - (b.urutan || 0))
  );

  let totalIndikatorJilid = $derived(currentIndikators.length);

  // Map skor dari detail
  let detailMap = $derived.by(() => {
    const map = {};
    for (const d of details) {
      map[d.kode] = d.nilai;
    }
    return map;
  });

  // Perhitungan Nilai Rata-rata dari total indikator di jilidnya
  let totalPoin = $derived.by(() => {
    return currentIndikators.reduce((sum, ind) => {
      const val = detailMap[ind.kode];
      return sum + (val !== undefined && val !== null ? Number(val) : 0);
    }, 0);
  });

  let nilaiAkhirCalculated = $derived.by(() => {
    if (totalIndikatorJilid > 0) {
      return Number((totalPoin / totalIndikatorJilid).toFixed(2));
    }
    return nilaiHeader?.nilai_akhir || 0;
  });

  let predikatCalculated = $derived.by(() => {
    const score = nilaiAkhirCalculated;
    if (score <= 0) return '-';
    if (score <= 5) {
      if (score >= 4.5) return 'A (Sangat Baik)';
      if (score >= 3.5) return 'B (Baik)';
      if (score >= 2.5) return 'C (Cukup)';
      if (score >= 1.5) return 'D (Kurang)';
      return 'E (Perlu Bimbingan)';
    }
    if (score >= 90) return 'A (Sangat Baik)';
    if (score >= 80) return 'B (Baik)';
    if (score >= 70) return 'C (Cukup)';
    if (score >= 60) return 'D (Kurang)';
    return 'E (Perlu Bimbingan)';
  });
</script>

<svelte:head>
  <title>Rapor Ujian Tahsin - {siswa?.nama_lengkap || 'Siswa'}</title>
</svelte:head>

{#if loading}
  <div class="flex justify-center items-center p-20 text-gray-500">
    <p>Memuat rapor tahsin siswa...</p>
  </div>
{:else if error}
  <div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded m-8">
    <strong>Error:</strong> {error}
  </div>
{:else}
  <div class="print-wrapper bg-gray-100 py-8 font-sans text-sm text-black min-h-screen">
    <!-- PRINT ACTIONS -->
    <div class="print-actions no-print">
      <button class="btn-back" onclick={() => history.back()}>
        <Icon name="chevronLeft" size={18} />
        Kembali
      </button>
      <button class="btn-print" onclick={() => window.print()}>
        <Icon name="printer" size={18} />
        Cetak Rapor
      </button>
    </div>

    <!-- PENGATURAN CETAK (NO PRINT) -->
    <div class="settings-box no-print">
      <h4 class="font-bold mb-2">Pengaturan Cetak (Tidak ikut terprint)</h4>
      <div class="settings-grid">
        <label>
          Kota:
          <input type="text" class="form-input" bind:value={customKota} />
        </label>
        <label>
          Tanggal Rapor:
          <input type="text" class="form-input" bind:value={customTanggalRapor} />
        </label>
        <label>
          Kepala Sekolah:
          <input type="text" class="form-input" bind:value={customKepalaSekolah} />
        </label>
        <label>
          Guru Penguji / Pembimbing:
          <input
            type="text"
            class="form-input"
            bind:value={customGuruPenguji}
            placeholder="Nama Guru Penguji"
          />
        </label>
        <label>
          Ttd Kepala Sekolah (Kanan):
          <input type="file" class="form-input" accept="image/*" onchange={handleTtdKepala} />
        </label>
        <label>
          Ukuran Ttd Kepala ({ttdKepalaSize}%):
          <input type="range" min="10" max="300" class="form-input" bind:value={ttdKepalaSize} />
        </label>
        <label>
          Ttd Penguji (Kiri):
          <input type="file" class="form-input" accept="image/*" onchange={handleTtdPenguji} />
        </label>
        <label>
          Ukuran Ttd Penguji ({ttdPengujiSize}%):
          <input type="range" min="10" max="300" class="form-input" bind:value={ttdPengujiSize} />
        </label>
      </div>
    </div>

    <!-- A4 CONTAINER (MIRIP RAPOR KESEHARIAN) -->
    <div class="a4-container" style="position:relative;">
      <!-- KOP SURAT -->
      <div class="header-section" style="text-align: center; margin-bottom: 16px;">
        {#if customLogoUrl}
          <img
            src={customLogoUrl}
            alt="Logo"
            style="position:absolute; left: {logoX}px; top: {logoY}px; width: {logoSize}px; z-index: 50;"
          />
        {/if}
        {#each kopLines as line}
          <div
            style="color: {line.color}; font-size: {line.size}px; font-weight: {line.bold
              ? 'bold'
              : 'normal'}; margin: {line.bold ? '4px 0' : '0'}; letter-spacing: 1px;"
          >
            {line.text}
          </div>
        {/each}
      </div>

      <!-- DOUBLE BORDER LINE -->
      <div
        style="border-top: {dividerThickness}px double black; border-bottom: none; height: 0; margin-bottom: 24px;"
      ></div>

      <!-- TITLE SECTION -->
      <div class="title-section">
        <h3 class="title-main">RAPOR UJIAN TAHSIN AL-QUR'AN</h3>
        <p class="title-sub">METODE BACAAN IQRO (JILID {activeJilid})</p>
      </div>

      <!-- STUDENT INFO -->
      <div class="student-info" style="display: flex; justify-content: space-between;">
        <table class="info-table" style="width: auto;">
          <tbody>
            <tr>
              <td class="info-label" style="width: 110px;">NAMA SISWA</td>
              <td class="info-sep">:</td>
              <td style="font-weight: 700;">{siswa?.nama_lengkap || '-'}</td>
            </tr>
            <tr>
              <td class="info-label">NISN</td>
              <td class="info-sep">:</td>
              <td>{siswa?.nisn || '-'}</td>
            </tr>
            <tr>
              <td class="info-label">KELAS</td>
              <td class="info-sep">:</td>
              <td>{kelasObj?.nama_kelas || kelasObj?.tingkat || riwayat?.kelas_asal || '-'}</td>
            </tr>
          </tbody>
        </table>

        <table class="info-table" style="width: auto;">
          <tbody>
            <tr>
              <td class="info-label" style="width: 140px;">IQRO DIUJIKAN</td>
              <td class="info-sep">:</td>
              <td style="font-weight: 700;">Jilid {activeJilid}</td>
            </tr>
            <tr>
              <td class="info-label">SEMESTER</td>
              <td class="info-sep">:</td>
              <td>{semester?.nama_semester || semester?.tipe_semester || 'Ganjil'}</td>
            </tr>
            <tr>
              <td class="info-label">TAHUN AJARAN</td>
              <td class="info-sep">:</td>
              <td>{tahunAjaran?.nama_tahun || '2026/2027'}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- TABEL INDIKATOR PENILAIAN -->
      <div class="mb-4">
        <table class="rapor-table">
          <thead>
            <tr>
              <th class="w-12 text-center" style="width: 45px;">NO</th>
              <th>INDIKATOR PENILAIAN</th>
              <th class="w-24 text-center" style="width: 90px;">NILAI</th>
            </tr>
          </thead>
          <tbody>
            {#if currentIndikators.length > 0}
              {#each currentIndikators as ind, i (ind.id || ind.kode)}
                {@const score = detailMap[ind.kode]}
                <tr>
                  <td class="text-center">{i + 1}</td>
                  <td>{ind.indikator}</td>
                  <td class="text-center font-bold">
                    {score !== undefined && score !== null ? score : '-'}
                  </td>
                </tr>
              {/each}
            {:else}
              <tr>
                <td colspan="3" class="text-center py-4 text-gray-500">
                  Belum ada indikator yang terdaftar untuk Iqro {activeJilid}.
                </td>
              </tr>
            {/if}

            <!-- TOTAL POIN -->
            <tr class="font-bold bg-gray-50 print:bg-transparent">
              <td colspan="2" class="text-right pr-4">TOTAL POIN</td>
              <td class="text-center">{totalPoin}</td>
            </tr>

            <!-- NILAI RATA-RATA -->
            <tr class="font-bold bg-gray-50 print:bg-transparent">
              <td colspan="2" class="text-right pr-4">
                NILAI AKHIR (RATA-RATA DIBAGI {totalIndikatorJilid} INDIKATOR)
              </td>
              <td class="text-center" style="font-size: 14px; color: black;">
                {nilaiAkhirCalculated}
              </td>
            </tr>

            <!-- PREDIKAT -->
            <tr class="font-bold bg-gray-50 print:bg-transparent">
              <td colspan="2" class="text-right pr-4">PREDIKAT</td>
              <td class="text-center">{predikatCalculated}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- CATATAN GURU PEMBIMBING -->
      <div class="catatan-box">
        <strong style="display: block; margin-bottom: 4px; font-size: 13px;">
          CATATAN / EVALUASI GURU PEMBIMBING:
        </strong>
        <p style="margin: 0; line-height: 1.5; font-size: 13px; min-height: 40px;">
          {catatan || 'Alhamdulillah, siswa telah menyelesaikan ujian tahsin pada jilid ini dengan baik. Terus tingkatkan kelancaran dan ketepatan kaidah tajwid.'}
        </p>
      </div>

      <!-- TANDA TANGAN (SIGNATURES) -->
      <div class="signature-section">
        <!-- PENGUJI (KIRI) -->
        <div class="sig-col">
          <div class="sig-invisible">Tempat, Tanggal</div>
          <div>Guru Penguji / Pembimbing,</div>
          <div class="sig-space" style="display: flex; align-items: center; justify-content: center;">
            {#if ttdPenguji}
              <img src={ttdPenguji} alt="Ttd Penguji" style="max-height: 70px; width: {ttdPengujiSize}%;" />
            {/if}
          </div>
          <p class="sig-name">{customGuruPenguji || '( .................................................... )'}</p>
        </div>

        <!-- KEPALA SEKOLAH (KANAN) -->
        <div class="sig-col">
          <div>{customKota}, {customTanggalRapor}</div>
          <div>Mengetahui,</div>
          <div style="font-size: 12px; margin-top: -2px;">Kepala Sekolah,</div>
          <div class="sig-space" style="display: flex; align-items: center; justify-content: center;">
            {#if ttdKepala}
              <img src={ttdKepala} alt="Ttd Kepala Sekolah" style="max-height: 70px; width: {ttdKepalaSize}%;" />
            {/if}
          </div>
          <p class="sig-name">{customKepalaSekolah || 'M. HUSAINI, S.Pd.I.'}</p>
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  :global(body) {
    background-color: #f3f4f6;
    margin: 0;
  }

  .print-actions {
    display: flex;
    justify-content: center;
    gap: 1rem;
    align-items: center;
    max-width: 210mm;
    margin: 0 auto 1.5rem;
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
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
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

  .settings-box {
    max-width: 210mm;
    margin: 0 auto 1.5rem;
    background: white;
    padding: 1.25rem;
    border-radius: 12px;
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
    gap: 0.35rem;
  }
  .form-input {
    padding: 0.5rem;
    border: 1px solid #cbd5e1;
    border-radius: 6px;
    font-family: inherit;
    font-size: 0.85rem;
  }

  /* A4 CONTAINER */
  .a4-container {
    background: white;
    width: 100%;
    max-width: 210mm;
    min-height: auto;
    margin: 0 auto;
    padding: 1.2cm !important;
    box-sizing: border-box;
    box-shadow: 0 8px 30px rgba(0, 0, 0, 0.12);
    font-family: Arial, Helvetica, sans-serif;
    font-size: 13.5px;
    color: black;
  }

  .header-section {
    position: relative;
    text-align: center;
    margin-bottom: 16px;
  }

  .title-section {
    text-align: center;
    margin-bottom: 20px;
  }
  .title-main {
    font-weight: 800;
    font-size: 16px;
    text-decoration: underline;
    text-underline-offset: 4px;
    margin: 0;
    letter-spacing: 0.5px;
  }
  .title-sub {
    font-weight: 700;
    font-size: 13px;
    margin: 4px 0 0 0;
    color: #334155;
  }

  .student-info {
    display: flex;
    justify-content: space-between;
    margin-bottom: 16px;
    font-size: 13.5px;
  }

  .info-table {
    border: none;
    width: auto;
  }
  .info-table td {
    border: none;
    padding: 3px 0;
    vertical-align: top;
  }
  .info-label {
    color: #334155;
  }
  .info-sep {
    padding: 3px 8px;
  }

  .rapor-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 13px;
  }

  .rapor-table th,
  .rapor-table td {
    border: 1px solid black;
    padding: 6px 10px;
    vertical-align: middle;
  }

  .rapor-table th {
    font-weight: 700;
    background: #f8fafc;
    text-align: left;
  }

  .text-center {
    text-align: center !important;
  }
  .text-right {
    text-align: right !important;
  }

  .catatan-box {
    margin-top: 16px;
    border: 1px solid black;
    padding: 10px 12px;
    min-height: 60px;
  }

  .signature-section {
    display: flex;
    justify-content: space-between;
    margin-top: 32px;
    text-align: center;
    font-size: 13.5px;
    padding: 0 24px;
  }

  .sig-col {
    display: flex;
    flex-direction: column;
    align-items: center;
    min-width: 220px;
  }
  .sig-invisible {
    visibility: hidden;
  }
  .sig-space {
    height: 75px;
  }
  .sig-name {
    font-weight: 700;
    text-decoration: underline;
    margin: 0;
  }

  @media print {
    :global(body) {
      background: white !important;
      margin: 0 !important;
    }
    .no-print {
      display: none !important;
    }
    .print-wrapper {
      padding: 0 !important;
      background: white !important;
    }
    .a4-container {
      box-shadow: none !important;
      margin: 0 !important;
      padding: 0 !important;
      width: 100% !important;
      max-width: none !important;
    }
  }

  @page {
    size: A4 portrait;
    margin: 1cm;
  }
</style>
