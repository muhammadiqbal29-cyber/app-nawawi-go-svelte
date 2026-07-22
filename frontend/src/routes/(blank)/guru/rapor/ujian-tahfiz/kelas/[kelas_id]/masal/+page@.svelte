<script>
  import { toastStore } from '$lib/toast.svelte.js';
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import Icon from '$lib/components/ui/Icon.svelte';

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

  let kelasId = $derived($page.params.kelas_id);
  let rombel = $derived($page.url.searchParams.get('rombel') || '');
  let sekolahId = $derived($page.url.searchParams.get('sekolah_id') || '1');
  let jenisPenilaian = 'semester';

  let loading = $state(true);
  let error = $state('');
  let raporList = $state([]);
  let kriteriaPenilaian = $state([]);

  let currentDate = new Date().toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric'
  });

  // Dynamic Print Settings
  let customTanggalRapor = $state(currentDate);
  let customKota = $state('Jakarta');
  let customGuruPembimbing = $state('Ust. Muhammad Rasyid, S.Pd.I');
  let customOrangTua = $state('____________________');
  let customSubJudul = $state("Evaluasi Pembelajaran Juz 30 (Juz 'Amma)");

  let logoX = $state(600);
  let logoY = $state(20);
  let logoSize = $state(80);
  let dividerThickness = $state(3);
  let customLogoUrl = $state('');
  let kopLines = $state([]);

  // Local Signature Settings
  let ttdMusyrif = $state('');
  let ttdMusyrifSize = $state(100);

  $effect(() => {
    if (ttdMusyrifSize !== undefined) localStorage.setItem('ttdMusyrifSize_global', ttdMusyrifSize.toString());
  });

  function handleTtdMusyrif(e) {
    const file = e.target.files[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (event) => {
      ttdMusyrif = event.target.result;
      localStorage.setItem('ttdMusyrif_global', ttdMusyrif);
    };
    reader.readAsDataURL(file);
  }

  $effect(() => {
    if (raporList.length > 0 && raporList[0].sekolah) {
      const sekolah = raporList[0].sekolah;
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
            text: sekolah.alamat || 'Jl. Tjilik Riwut KM 8',
            color: '#000000',
            size: 12,
            bold: false
          }
        ];
      }
      kopLines = localKop;
    }
  });

  onMount(async () => {
    try {
      await Promise.all([loadKonfigurasiTahfiz(), loadDataMasal()]);
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  });

  async function loadKonfigurasiTahfiz() {
    try {
      const res = await fetch(`${apiUrl}/api/guru/sekolah/${sekolahId}/konfigurasi-nilai-tahfiz`, {
        credentials: 'include'
      });
      if (res.ok) {
        const body = await res.json();
        kriteriaPenilaian = body.kriteria_penilaian || [];
      }
    } catch {
      // Use fallback if fails
    }
  }

  async function loadDataMasal() {
    const params = new URLSearchParams();
    if (sekolahId) params.append('sekolah_id', sekolahId);
    if (rombel && rombel !== '__kosong__') params.append('rombel', rombel);
    params.append('jenis_penilaian', jenisPenilaian);

    const res = await fetch(
      `${apiUrl}/api/guru/kelas/${kelasId}/tahfiz-masal?${params.toString()}`,
      {
        credentials: 'include'
      }
    );
    if (!res.ok) throw new Error('Gagal memuat rapor tahfiz masal.');
    const data = await res.json();
    raporList = data || [];

    // Set default print settings based on first data if available
    customTanggalRapor = formatTanggal();
    ttdMusyrif = localStorage.getItem('ttdMusyrif_global') || '';
    
    const mSize = localStorage.getItem('ttdMusyrifSize_global');
    if (mSize) ttdMusyrifSize = parseInt(mSize) || 100;

    if (raporList.length > 0 && raporList[0].sekolah) {
      const s = raporList[0].sekolah;
      // Removed custom state mapping, we'll use raporData.sekolah directly in the template
    }
  }

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
    return `${now.getDate()} ${months[now.getMonth()]} ${now.getFullYear()}`;
  }

  function getPredicate(score) {
    if (!kriteriaPenilaian || kriteriaPenilaian.length === 0) {
      if (score >= 90) return 'مُمْتَاز';
      if (score >= 80) return 'جَيِّد جِدًّا';
      if (score >= 70) return 'جَيِّد';
      if (score >= 60) return 'مَقْبُول';
      return 'رَاسِب';
    }
    for (const k of kriteriaPenilaian) {
      if (score >= k.min && score <= k.max) {
        return k.arab;
      }
    }
    return '';
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
            filename: `Rapor_Tahfiz_${rapor.siswa.nama_lengkap.replace(/\s+/g, '_')}.pdf`,
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
</script>

<svelte:head>
  <title>Cetak Rapor Ujian Tahfiz Masal - SIAKAD</title>
</svelte:head>

<div class="page-wrapper">
  {#if loading}
    <div class="loading-state">Memuat data rapor masal...</div>
  {:else if error}
    <div class="alert-error">{error}</div>
  {:else if raporList.length > 0}
    <div
      class="controls no-print"
      style="position: static; display: flex; justify-content: center; gap: 1rem; align-items: center; max-width: 210mm; margin: 0 auto 1.5rem;"
    >
      <button class="btn-secondary" onclick={() => goto('/guru/rapor/ujian-tahfiz')}>
        <Icon name="arrowLeft" size={16} /> Kembali
      </button>
      <button class="btn-primary" onclick={() => window.print()}>
        <Icon name="printer" size={16} /> Cetak Rapor (Print)
      </button>
      <button class="btn-download" onclick={downloadSemuaPDF} disabled={isDownloading}>
        {isDownloading ? 'Memproses...' : '⬇️ Download PDF (Per Siswa)'}
      </button>
    </div>

    <!-- Edit Settings -->
    <div
      class="settings-box no-print"
      style="max-width: 210mm; margin: 0 auto 1.5rem; background: white; padding: 1rem; border-radius: 8px; box-shadow: 0 4px 12px rgba(0,0,0,0.05); border: 1px solid #e2e8f0;"
    >
      <h4 style="margin-top: 0;">Pengaturan Cetak (Tidak ikut terprint)</h4>
      <div
        style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem;"
      >
        <label
          style="display: flex; flex-direction: column; font-size: 0.85rem; font-weight: 600; color: #475569; gap: 0.25rem;"
        >
          Kota:
          <input type="text" class="form-input" bind:value={customKota} />
        </label>
        <label
          style="display: flex; flex-direction: column; font-size: 0.85rem; font-weight: 600; color: #475569; gap: 0.25rem;"
        >
          Tanggal Rapor (contoh: 21 Juni 2026):
          <input type="text" class="form-input" bind:value={customTanggalRapor} />
        </label>
        <label
          style="display: flex; flex-direction: column; font-size: 0.85rem; font-weight: 600; color: #475569; gap: 0.25rem;"
        >
          Guru Pembimbing Tahfiz:
          <input type="text" class="form-input" bind:value={customGuruPembimbing} />
        </label>
        <label
          style="display: flex; flex-direction: column; font-size: 0.85rem; font-weight: 600; color: #475569; gap: 0.25rem;"
        >
          Keterangan Evaluasi (Sub-Judul):
          <input type="text" class="form-input" bind:value={customSubJudul} />
        </label>
        <label
          style="display: flex; flex-direction: column; font-size: 0.85rem; font-weight: 600; color: #475569; gap: 0.25rem;"
        >
          Ttd Guru Pembimbing:
          <input type="file" class="form-input" accept="image/*" onchange={handleTtdMusyrif} />
        </label>
        <label
          style="display: flex; flex-direction: column; font-size: 0.85rem; font-weight: 600; color: #475569; gap: 0.25rem;"
        >
          Ukuran Ttd Guru Pembimbing ({ttdMusyrifSize}%):
          <input type="range" min="10" max="300" class="form-input" bind:value={ttdMusyrifSize} />
        </label>
      </div>
    </div>

    <!-- Render all students -->
    {#each raporList as raporData}
      {@const details = raporData.details || []}
      {@const totalPoin = details.reduce((sum, d) => sum + (d.total_poin || 0), 0)}
      {@const maxPoin = details.length * 25}
      {@const averageScore =
        raporData.nilai?.nilai_akhir != null
          ? Math.round(raporData.nilai.nilai_akhir)
          : details.length > 0
            ? Math.round(totalPoin / (details.length * 5))
            : 0}

      <!-- A4 Paper Container -->
      <div class="page" id="rapor-container-{raporData.siswa.id}">
        <div style="position:relative;">
          <div class="report-header">
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
            <!-- Double Border Line -->
            <div
              style="border-top: {dividerThickness}px double black; border-bottom: none; height: 0; margin-bottom: 24px;"
            ></div>
            <div class="report-title">
              <h2>RAPOR HAFALAN AL-QUR'AN</h2>
              <h3>{customSubJudul}</h3>
            </div>
          </div>

          <div class="student-summary">
            <table id="student-info-table">
              <tbody>
                <tr>
                  <td class="label">Nama Siswa</td>
                  <td>:</td>
                  <td class="value">{raporData.siswa?.nama_lengkap || '-'}</td>
                </tr>
                <tr>
                  <td class="label">NIS / NISN</td>
                  <td>:</td>
                  <td class="value">{raporData.siswa?.nisn || '-'}</td>
                </tr>
                <tr>
                  <td class="label">Kelas</td>
                  <td>:</td>
                  <td class="value">{raporData.kelas?.nama_kelas || '-'}</td>
                </tr>
              </tbody>
            </table>

            <div class="predicate-box">
              <div class="predicate-label">PREDIKAT</div>
              <div class="arabic">{getPredicate(averageScore)}</div>
            </div>
          </div>

          {#if details.length > 0}
            <table id="tahfiz-table">
              <thead>
                <tr>
                  <th rowspan="2">No</th>
                  <th rowspan="2" class="text-left">Nama Surah</th>
                  <th colspan="5">Kategori Penilaian</th>
                  <th rowspan="2">Total Poin</th>
                </tr>
                <tr>
                  <th class="col-kategori">Hafalan</th>
                  <th class="col-kategori">Kelancaran</th>
                  <th class="col-kategori">Mad</th>
                  <th class="col-kategori">Tajwid</th>
                  <th class="col-kategori">Makharij</th>
                </tr>
              </thead>
              <tbody>
                {#each details as row, index (row.id || index)}
                  <tr>
                    <td>{index + 1}</td>
                    <td class="text-left">{row.nama_materi}</td>
                    <td>{row.skor_hafalan || 0}</td>
                    <td>{row.skor_kelancaran || 0}</td>
                    <td>{row.skor_mad || 0}</td>
                    <td>{row.skor_tajwid || 0}</td>
                    <td>{row.skor_makharijul_huruf || 0}</td>
                    <td>{row.total_poin || 0}</td>
                  </tr>
                {/each}
              </tbody>
              <tfoot>
                <tr class="tfoot-row">
                  <td colspan="7" class="tfoot-label">Total Poin Diperoleh</td>
                  <td class="tfoot-value">{totalPoin} / {maxPoin}</td>
                </tr>
                <tr class="tfoot-row">
                  <td colspan="7" class="tfoot-label">Rata-rata Nilai (Skala 100)</td>
                  <td class="tfoot-value">{averageScore}</td>
                </tr>
              </tfoot>
            </table>
          {:else}
            <div class="empty-state-message">
              <p>Belum ada data nilai tahfiz untuk siswa ini.</p>
            </div>
          {/if}

          <!-- Catatan -->
          <div class="catatan-box">
            <p style="font-weight: 700; margin-bottom: 4px;">CATATAN:</p>
            <p style="white-space: pre-wrap;">{raporData.nilai?.catatan || '-'}</p>
          </div>

          <div class="report-footer">
            <div class="note-box">
              <p class="note-title">Kriteria Penilaian Rapor:</p>
              <table>
                <tbody>
                  {#if kriteriaPenilaian && kriteriaPenilaian.length > 0}
                    {#each kriteriaPenilaian as k (k.id || k.min)}
                      <tr>
                        <td>{k.min === 0 ? `< ${k.max + 1}` : `${k.min} - ${k.max}`}</td>
                        <td>: {k.arab} ({k.indo})</td>
                      </tr>
                    {/each}
                  {:else}
                    <tr><td>90 - 100</td><td>: مُمْتَاز (Istimewa)</td></tr>
                    <tr><td>80 - 89</td><td>: جَيِّد جِدًّا (Sangat Baik)</td></tr>
                    <tr><td>70 - 79</td><td>: جَيِّد (Baik)</td></tr>
                    <tr><td>60 - 69</td><td>: مَقْبُول (Cukup)</td></tr>
                    <tr><td>&lt; 60</td><td>: رَاسِب (Mengulang)</td></tr>
                  {/if}
                </tbody>
              </table>
            </div>
            <div class="signatures">
              <div>
                <p class="signature-label">Mengetahui,</p>
                <p class="signature-title">Orang Tua / Wali Siswa</p>
                <div class="signature-line">( {customOrangTua} )</div>
              </div>
              <div>
                <p>{customKota}, {customTanggalRapor}</p>
                <p class="signature-title">Guru Pembimbing Tahfiz</p>
                {#if ttdMusyrif}
                  <div style="height: 3rem; margin-top: 0.5rem; display: flex; justify-content: center; align-items: center;">
                    <img src={ttdMusyrif} alt="Ttd Musyrif" style="max-height: 100%; max-width: 100%; object-fit: contain; transform: scale({ttdMusyrifSize / 100}); transform-origin: center;" />
                  </div>
                  <div class="signature-line" style="margin-top: 0.5rem;">( {customGuruPembimbing} )</div>
                {:else}
                  <div class="signature-line">( {customGuruPembimbing} )</div>
                {/if}
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="page-break" style="page-break-after: always;"></div>
    {/each}
  {:else}
    <div class="empty-state-message" style="margin-top: 100px;">Data rapor masal kosong.</div>
  {/if}
</div>

<style>
  @import url('https://fonts.googleapis.com/css2?family=Aref+Ruqaa&family=Cinzel:wght@600&family=Inter:wght@400;600;700&display=swap');

  :root {
    --print-width: 210mm;
    --print-height: 297mm;
  }

  :global(body) {
    background-color: #f3f4f6;
    font-family: 'Inter', sans-serif;
    color: #000;
    margin: 0;
    padding: 0;
  }

  .form-input {
    padding: 0.5rem;
    border: 1px solid #cbd5e1;
    border-radius: 6px;
    font-family: inherit;
  }
  .btn-download {
    background: #10b981;
    color: white;
    border: none;
    border-radius: 0.5rem;
    padding: 0.625rem 1.25rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
  }
  .btn-download:hover:not(:disabled) {
    background: #059669;
  }
  .btn-download:disabled {
    background: #94a3b8;
    cursor: not-allowed;
  }

  .controls {
    display: flex;
    gap: 0.5rem;
  }

  .btn-primary,
  .btn-secondary {
    border: none;
    border-radius: 0.5rem;
    padding: 0.625rem 1.25rem;
    font-weight: 600;
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
  }

  .btn-primary {
    background: #111827;
    color: #fff;
  }

  .btn-secondary {
    background: #f8fafc;
    color: #111827;
    border: 1px solid #d1d5db;
  }

  .loading-state {
    text-align: center;
    padding: 3rem 1rem;
    color: #64748b;
    font-size: 0.95rem;
    font-weight: 600;
  }

  .alert-error {
    margin: 0.75rem 0;
    padding: 1rem 1.2rem;
    border-radius: 0.5rem;
    color: #b91c1c;
    background: #fee2e2;
    border: 1px solid #fecaca;
    font-size: 0.875rem;
    text-align: center;
    max-width: 400px;
    margin: 100px auto;
  }

  .empty-state-message {
    text-align: center;
    padding: 2rem;
    border: 1px dashed #cbd5e1;
    border-radius: 0.5rem;
    color: #64748b;
    font-size: 0.9rem;
    margin-bottom: 1.5rem;
  }

  .empty-state-message p {
    margin: 0;
  }

  .page {
    width: var(--print-width);
    min-height: var(--print-height);
    padding: 1cm !important;
    margin: 20px auto;
    background: white;
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
    box-sizing: border-box;
    position: relative;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
  }

  .report-header {
    text-align: center;
    margin-bottom: 1.5rem;
  }

  .header-section {
    position: relative;
    text-align: center;
    margin-bottom: 16px;
  }

  .header-line {
    border-top: 3px double black;
    border-bottom: none;
    height: 0;
    margin-bottom: 24px;
  }

  .report-title {
    margin-bottom: 0.25rem;
  }

  .report-title h2 {
    margin: 0;
    font-size: 1.15rem;
    letter-spacing: 0.08em;
    font-weight: 700;
    text-transform: uppercase;
    color: #000;
    text-decoration: underline;
    text-underline-offset: 4px;
  }

  .report-title h3 {
    margin: 0.25rem 0 0;
    font-size: 0.95rem;
    font-weight: 600;
    color: #404040;
  }

  .student-summary {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 1.5rem;
    margin-bottom: 1.5rem;
  }

  #student-info-table {
    width: 55%;
    border-collapse: collapse;
    font-size: 0.875rem;
  }

  #student-info-table td {
    border: none;
    padding: 4px 0;
    vertical-align: top;
  }

  .label {
    width: 35%;
    font-weight: 600;
    color: #525252;
  }

  .value {
    font-weight: 700;
    color: #111827;
    border-bottom: 1px solid #d1d5db;
  }

  .predicate-box {
    width: 16rem;
    padding: 0;
    margin-top: 4px;
    text-align: center;
    display: flex;
    flex-direction: column;
    justify-content: flex-start;
    align-items: center;
  }

  .predicate-label {
    font-size: 0.75rem;
    font-weight: 800;
    letter-spacing: 0.15em;
    text-transform: uppercase;
    width: 100%;
    padding-bottom: 0.25rem;
  }

  .arabic {
    font-family: 'Aref Ruqaa', serif;
    font-size: 3.5rem;
    font-weight: 700;
    line-height: 1.3;
    padding-top: 0;
    margin-top: -0.5rem;
  }

  #tahfiz-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 9.5pt;
    margin-bottom: 1.5rem;
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  }

  #tahfiz-table th,
  #tahfiz-table td {
    border: 1px solid #000;
    padding: 4px 6px;
    text-align: center;
    color: #000;
  }

  #tahfiz-table th {
    background: #f9fafb;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-size: 8.5pt;
  }

  #tahfiz-table th.col-kategori {
    width: 12%;
    font-size: 7.5pt;
  }

  .tfoot-row td {
    border: 1px solid #000;
    padding: 5px 8px;
    font-weight: 700;
    background: #f9fafb;
  }

  .tfoot-label {
    text-align: right;
    font-size: 9pt;
    text-transform: uppercase;
    letter-spacing: 0.03em;
  }

  .tfoot-value {
    text-align: center;
    font-size: 10pt;
    font-weight: 800;
    color: #111827;
  }

  .text-left {
    text-align: left !important;
  }

  .report-footer {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    gap: 1rem;
    margin-top: 1.5rem;
  }

  .note-box {
    width: 41.666667%;
    border: 1px solid #000;
    padding: 0.625rem;
    font-size: 8pt;
  }

  .catatan-box {
    margin-top: 1rem;
    margin-bottom: 0.5rem;
    border: 1px solid #000;
    padding: 0.5rem;
    min-height: 80px;
    font-size: 9pt;
  }
  .catatan-box p {
    margin: 0;
  }

  .note-title {
    margin: 0 0 0.5rem;
    font-weight: 800;
  }

  .note-box table {
    width: 100%;
    border-collapse: collapse;
    font-size: 7.5pt;
  }

  .note-box td {
    border: none;
    padding: 1px;
    text-align: left;
  }

  .signatures {
    width: 58.333333%;
    display: flex;
    justify-content: space-between;
    padding-left: 2rem;
    font-size: 0.75rem;
    text-align: center;
  }

  .signatures > div {
    width: 50%;
  }

  .signature-label,
  .signature-title {
    margin: 0;
  }

  .signature-title {
    font-weight: 600;
    margin-top: 0;
    color: #000;
  }

  .signature-line {
    display: inline-block;
    margin-top: 3rem;
    border-bottom: 1px solid #000;
    width: 10rem;
    padding-bottom: 0.25rem;
    font-weight: 500;
    color: #000;
  }

  .signatures > div:last-child .signature-line {
    width: 11rem;
    font-weight: 700;
  }

  @media print {
    :global(html),
    :global(body) {
      background: white;
      margin: 0;
      padding: 0;
      -webkit-print-color-adjust: exact;
      print-color-adjust: exact;
    }
    .page {
      margin: 0;
      padding: 1cm !important;
      box-shadow: none;
      width: 100%;
      min-height: auto !important;
      height: auto !important;
    }
    .no-print {
      display: none !important;
    }
    thead {
      display: table-header-group;
    }
    tr {
      page-break-inside: avoid;
    }
  }

  @media screen and (max-width: 900px) {
    .page {
      margin-left: 1rem;
      margin-right: 1rem;
    }
  }
  @page {
    size: A4 portrait;
    margin: 0;
  }
</style>
