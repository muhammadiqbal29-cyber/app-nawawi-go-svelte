<script>
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import Icon from '$lib/components/ui/Icon.svelte';
  import { toastStore } from '$lib/toast.svelte.js';

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

  let kelasId = $derived($page.params.kelas_id);
  let rombel = $derived($page.url.searchParams.get('rombel') || '');
  let sekolahId = $derived($page.url.searchParams.get('sekolah_id') || '');
  let paramTaId = $derived($page.url.searchParams.get('tahun_ajaran_id') || '');
  let paramSemId = $derived($page.url.searchParams.get('semester_id') || '');

  let loading = $state(true);
  let error = $state('');
  let raporList = $state([]);
  let filterOnlyScored = $state(true);

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

  // PDF Download State
  let isDownloading = $state(false);
  let dlProgress = $state(0);
  let dlTotal = $state(0);
  let dlCurrentName = $state('');
  let dlStage = $state('');

  onMount(() => {
    ttdKepala = localStorage.getItem('ttdKepala_global') || '';
    ttdPenguji = localStorage.getItem('ttdPenguji_tahsin_global') || localStorage.getItem('ttdMusyrif_global') || '';
    ttdKepalaSize = parseInt(localStorage.getItem('ttdKepalaSize_global')) || 100;
    ttdPengujiSize = parseInt(localStorage.getItem('ttdPengujiSize_tahsin_global')) || 100;
    loadData();
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

      // Coba fetch Tahsin V2 Masal
      let res = await fetch(`${apiUrl}/api/guru/kelas/${kelasId}/tahsin-v2-masal?${params.toString()}`, {
        credentials: 'include'
      });

      // Fallback ke V1 Masal
      if (!res.ok) {
        res = await fetch(`${apiUrl}/api/guru/kelas/${kelasId}/tahsin-masal?${params.toString()}`, {
          credentials: 'include'
        });
      }

      if (!res.ok) {
        throw new Error('Gagal memuat data rapor ujian tahsin masal.');
      }

      const rawData = await res.json();
      const sortedData = (rawData || []).sort((a, b) => {
        const nameA = a.siswa?.nama_lengkap || '';
        const nameB = b.siswa?.nama_lengkap || '';
        return nameA.localeCompare(nameB);
      });

      // Ambil Catatan Rapor Khusus Tahsin
      const paramsCatatan = new URLSearchParams({
        jenis_rapor: 'tahsin'
      });
      if (sekolahId) paramsCatatan.append('sekolah_id', sekolahId);
      if (rombel && rombel !== '__kosong__') paramsCatatan.append('rombel', rombel);
      if (paramTaId) paramsCatatan.append('tahun_ajaran_id', paramTaId);
      if (paramSemId) paramsCatatan.append('semester_id', paramSemId);

      const resCatatan = await fetch(`${apiUrl}/api/guru/kelas/${kelasId}/catatan-rapor?${paramsCatatan.toString()}`, {
        credentials: 'include'
      });

      let catMap = {};
      if (resCatatan.ok) {
        const catBody = await resCatatan.json();
        const catList = catBody.catatan || [];
        catList.forEach((c) => {
          catMap[c.siswa_id] = c.catatan;
        });
      }

      raporList = sortedData.map((item) => {
        return {
          ...item,
          catatan: catMap[item.siswa?.id] || item.catatan || item.nilai?.catatan || ''
        };
      });

      // Ambil metadata kop surat dari item pertama yang punya sekolah
      const firstSekolah = raporList.find((r) => r.sekolah)?.sekolah;
      if (firstSekolah) {
        if (firstSekolah.kepala_sekolah) {
          customKepalaSekolah = firstSekolah.kepala_sekolah;
        }
        customLogoUrl = firstSekolah.logo_base64 || '';
        logoX = firstSekolah.logo_x || 600;
        logoY = firstSekolah.logo_y || 20;
        logoSize = firstSekolah.logo_size || 80;
        dividerThickness = firstSekolah.divider_thickness || 3;
        let localKop = [];
        if (firstSekolah.kop_surat) {
          try {
            const parsed = JSON.parse(firstSekolah.kop_surat);
            if (Array.isArray(parsed) && parsed.length > 0) {
              localKop = parsed;
            }
          } catch (e) {}
        }
        if (localKop.length === 0) {
          localKop = [
            {
              text: firstSekolah.nama_sekolah || 'SMP ISLAM IMAM NAWAWI',
              color: '#000000',
              size: 18,
              bold: true
            },
            {
              text: firstSekolah.alamat || 'Jl. Tjilik Riwut KM 8, Palangka Raya',
              color: '#000000',
              size: 12,
              bold: false
            }
          ];
        }
        kopLines = localKop;
      }
    } catch (err) {
      error = err.message || 'Gagal memuat data rapor masal';
    } finally {
      loading = false;
    }
  }

  // Filter siswa: jika filterOnlyScored aktif, hanya tampilkan yang sudah ada nilai/jilid
  let scoredCount = $derived(
    raporList.filter((r) => r.nilai && r.nilai.jilid && r.details && r.details.length > 0).length
  );
  let totalCount = $derived(raporList.length);

  let displayedRaporList = $derived(
    filterOnlyScored
      ? raporList.filter((r) => r.nilai && r.nilai.jilid && r.details && r.details.length > 0)
      : raporList
  );

  function getActiveJilid(rapor) {
    return rapor.nilai?.jilid || 1;
  }

  function getIndikatorsForStudent(rapor) {
    const jilid = getActiveJilid(rapor);
    const list = rapor.indikator_list || [];
    return list
      .filter((ind) => ind.jilid === jilid)
      .sort((a, b) => (a.urutan || 0) - (b.urutan || 0));
  }

  function getDetailMap(rapor) {
    const map = {};
    const details = rapor.details || [];
    for (const d of details) {
      map[d.kode] = d.nilai;
    }
    return map;
  }

  function calculatePoin(rapor, indikators, map) {
    return indikators.reduce((sum, ind) => {
      const val = map[ind.kode];
      return sum + (val !== undefined && val !== null ? Number(val) : 0);
    }, 0);
  }

  function calculateNilaiAkhir(rapor, totalPoin, totalIndikator) {
    if (totalIndikator > 0) {
      return Number((totalPoin / totalIndikator).toFixed(2));
    }
    return rapor.nilai?.nilai_akhir || 0;
  }

  function calculatePredikat(score, rapor) {
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
  }

  async function downloadSemuaPDF() {
    if (displayedRaporList.length === 0) {
      toastStore.show('Tidak ada data siswa untuk di-download.', 'error');
      return;
    }
    isDownloading = true;
    dlProgress = 0;
    dlTotal = displayedRaporList.length;
    dlCurrentName = '';
    dlStage = 'pdf';

    try {
      let activeTaName = 'TA';
      try {
        const taRes = await fetch(`${apiUrl}/api/guru/tahun-ajaran`, { credentials: 'include' });
        if (taRes.ok) {
          const resData = await taRes.json();
          const tas = resData.data || resData;
          const ta = tas.find((t) => (paramTaId ? t.id == paramTaId : t.status === 'Active'));
          if (ta && ta.nama_tahun) activeTaName = ta.nama_tahun.replace(/\//g, '-');
        }
      } catch (e) {}

      const html2pdf = (await import('html2pdf.js')).default;
      const JSZip = (await import('jszip')).default;
      const { saveAs } = await import('file-saver');

      const zip = new JSZip();
      let semCleanGlobal = 'Sem';

      const noPrintEls = document.querySelectorAll('.no-print');
      noPrintEls.forEach((el) => {
        if (!el.classList.contains('dl-overlay')) el.style.display = 'none';
      });

      for (let idx = 0; idx < displayedRaporList.length; idx++) {
        const rapor = displayedRaporList[idx];
        dlProgress = idx + 1;
        dlCurrentName = rapor.siswa?.nama_lengkap || `Siswa_${idx + 1}`;
        const elId = `rapor-container-${rapor.siswa?.id}`;
        const element = document.getElementById(elId);

        if (element) {
          const taClean = rapor.tahun_ajaran
            ? (rapor.tahun_ajaran.nama_tahun || '').replace(/\//g, '-')
            : activeTaName;
          const semClean = rapor.semester ? rapor.semester.tipe_semester || 'Sem' : 'Sem';
          semCleanGlobal = semClean;
          const namaClean = (rapor.siswa?.nama_lengkap || 'Siswa').replace(/\s+/g, '_');

          const origStyles = {
            padding: element.style.padding,
            margin: element.style.margin,
            boxShadow: element.style.boxShadow,
            width: element.style.width,
            maxWidth: element.style.maxWidth
          };

          element.style.padding = '0';
          element.style.margin = '0';
          element.style.boxShadow = 'none';
          element.style.width = '210mm';
          element.style.maxWidth = 'none';

          const opt = {
            margin: 0,
            filename: `Rapor_Tahsin_${namaClean}_${semClean}_${taClean}.pdf`,
            image: { type: 'jpeg', quality: 1 },
            html2canvas: { scale: 2, useCORS: true, width: 794, windowWidth: 794 },
            jsPDF: { unit: 'mm', format: 'a4', orientation: 'portrait' }
          };

          const pdfBlob = await html2pdf().set(opt).from(element).outputPdf('blob');
          zip.file(opt.filename, pdfBlob);

          Object.assign(element.style, origStyles);
        }
      }

      noPrintEls.forEach((el) => (el.style.display = ''));

      dlStage = 'zip';
      dlCurrentName = 'Mengompresi ke file ZIP...';

      const zipBlob = await zip.generateAsync({ type: 'blob' });
      const zipFilename = `Rapor_Masal_Tahsin_Kelas_${kelasId}_${semCleanGlobal}_${activeTaName}.zip`;
      saveAs(zipBlob, zipFilename);

      dlStage = 'done';
      dlCurrentName = zipFilename;
      await new Promise((r) => setTimeout(r, 1500));

      toastStore.show('Berhasil mengunduh semua file PDF Rapor Tahsin!', 'success');
    } catch (err) {
      console.error(err);
      toastStore.show('Gagal memproses file PDF: ' + err.message, 'error');
    } finally {
      isDownloading = false;
      dlStage = '';
      document.querySelectorAll('.no-print').forEach((el) => (el.style.display = ''));
    }
  }
</script>

<svelte:head>
  <title>Cetak Rapor Ujian Tahsin Masal - SIAKAD</title>
</svelte:head>

{#if loading}
  <div class="flex justify-center items-center p-20 text-gray-500 font-sans">
    <p>Memuat rapor ujian tahsin masal...</p>
  </div>
{:else if error}
  <div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded m-8 font-sans">
    <strong>Error:</strong> {error}
  </div>
{:else}
  <div class="print-wrapper bg-gray-100 py-8 font-sans text-sm text-black min-h-screen">
    <!-- PRINT ACTIONS -->
    <div class="print-actions no-print">
      <button class="btn-back" onclick={() => goto('/guru/rapor/tahsin')}>
        <Icon name="chevronLeft" size={18} />
        Kembali
      </button>
      <button class="btn-print" onclick={() => window.print()}>
        <Icon name="printer" size={18} />
        Cetak Semua Rapor (Print)
      </button>
      <button class="btn-download" onclick={downloadSemuaPDF} disabled={isDownloading}>
        <Icon name="download" size={18} />
        {isDownloading ? '⏳ Memproses PDF...' : 'Download ZIP PDF'}
      </button>
    </div>

    <!-- PENGATURAN CETAK (NO PRINT) -->
    <div class="settings-box no-print">
      <div class="flex justify-between items-center mb-2">
        <h4 class="font-bold">Pengaturan Cetak Masal (Tidak ikut terprint)</h4>
        <label class="filter-toggle">
          <input type="checkbox" bind:checked={filterOnlyScored} />
          <span>Hanya cetak siswa yang telah dinilai ({scoredCount} dari {totalCount} siswa)</span>
        </label>
      </div>

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

    {#if displayedRaporList.length === 0}
      <div class="empty-box no-print">
        <Icon name="info" size={32} />
        <p class="font-semibold mt-2">Tidak ada data siswa untuk dicetak.</p>
        {#if filterOnlyScored && scoredCount === 0}
          <p class="text-xs text-gray-500 mt-1">
            Belum ada siswa di rombel ini yang memiliki nilai ujian tahsin. Anda dapat menonaktifkan filter di atas untuk melihat seluruh siswa.
          </p>
        {/if}
      </div>
    {:else}
      <!-- RENDER EACH STUDENT RAPOR -->
      {#each displayedRaporList as rapor (rapor.siswa?.id || Math.random())}
        {@const activeJilid = getActiveJilid(rapor)}
        {@const indikators = getIndikatorsForStudent(rapor)}
        {@const detailMap = getDetailMap(rapor)}
        {@const totalPoin = calculatePoin(rapor, indikators, detailMap)}
        {@const totalIndikator = indikators.length}
        {@const nilaiAkhir = calculateNilaiAkhir(rapor, totalPoin, totalIndikator)}
        {@const predikat = calculatePredikat(nilaiAkhir, rapor)}

        <div
          class="a4-container mb-8"
          id="rapor-container-{rapor.siswa?.id}"
          style="position:relative;"
        >
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
                  <td style="font-weight: 700;">{rapor.siswa?.nama_lengkap || '-'}</td>
                </tr>
                <tr>
                  <td class="info-label">NISN</td>
                  <td class="info-sep">:</td>
                  <td>{rapor.siswa?.nisn || '-'}</td>
                </tr>
                <tr>
                  <td class="info-label">KELAS</td>
                  <td class="info-sep">:</td>
                  <td>
                    {rapor.kelas?.nama_kelas || rapor.kelas?.tingkat || rapor.riwayat?.kelas_asal || '-'}
                  </td>
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
                  <td>{rapor.semester?.nama_semester || 'Ganjil'}</td>
                </tr>
                <tr>
                  <td class="info-label">TAHUN AJARAN</td>
                  <td class="info-sep">:</td>
                  <td>{rapor.tahun_ajaran?.nama_tahun || '2026/2027'}</td>
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
                {#if indikators.length > 0}
                  {#each indikators as ind, i (ind.id || ind.kode)}
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
                    NILAI AKHIR (RATA-RATA DIBAGI {totalIndikator} INDIKATOR)
                  </td>
                  <td class="text-center" style="font-size: 14px; color: black;">
                    {nilaiAkhir}
                  </td>
                </tr>

                <!-- PREDIKAT -->
                <tr class="font-bold bg-gray-50 print:bg-transparent">
                  <td colspan="2" class="text-right pr-4">PREDIKAT</td>
                  <td class="text-center">{predikat}</td>
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
              {rapor.catatan || 'Alhamdulillah, siswa telah menyelesaikan ujian tahsin pada jilid ini dengan baik. Terus tingkatkan kelancaran dan ketepatan kaidah tajwid.'}
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
      {/each}
    {/if}
  </div>
{/if}

<!-- DOWNLOAD PROGRESS OVERLAY -->
{#if isDownloading}
  <div class="dl-overlay no-print">
    <div class="dl-modal">
      <div class="dl-icon">
        <div class="dl-spinner"></div>
      </div>
      <h3 class="dl-title">Mengunduh PDF Rapor Tahsin</h3>
      <p class="dl-subtitle">
        {#if dlStage === 'pdf'}
          Memproses siswa ({dlProgress}/{dlTotal}):<br />
          <strong>{dlCurrentName}</strong>
        {:else if dlStage === 'zip'}
          {dlCurrentName}
        {:else if dlStage === 'done'}
          Selesai! Mengunduh <strong>{dlCurrentName}</strong>
        {/if}
      </p>
      <div class="dl-progress-wrap">
        <div
          class="dl-progress-bar"
          class:dl-progress-indeterminate={dlStage === 'zip'}
          style="width: {dlStage === 'pdf' ? (dlProgress / dlTotal) * 100 : 100}%"
        ></div>
      </div>
      <p class="dl-progress-text">
        {dlStage === 'pdf' ? `${Math.round((dlProgress / dlTotal) * 100)}% selesai` : 'Harap tunggu...'}
      </p>
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
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .btn-print {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    background: #0b7c3d;
    color: white;
    border: none;
    border-radius: 8px;
    padding: 0.6rem 1.75rem;
    font-weight: 700;
    font-size: 0.95rem;
    cursor: pointer;
    box-shadow: 0 4px 6px -1px rgba(11, 124, 61, 0.3);
  }

  .btn-download {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    background: #2563eb;
    color: white;
    border: none;
    border-radius: 8px;
    padding: 0.6rem 1.5rem;
    font-weight: 700;
    font-size: 0.95rem;
    cursor: pointer;
    box-shadow: 0 4px 6px -1px rgba(37, 99, 235, 0.3);
  }

  .btn-download:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .settings-box {
    background: white;
    border-radius: 12px;
    max-width: 210mm;
    margin: 0 auto 1.5rem;
    padding: 1.25rem;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    border: 1px solid #e5e7eb;
  }

  .filter-toggle {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.85rem;
    font-weight: 600;
    color: #374151;
    cursor: pointer;
  }

  .empty-box {
    background: white;
    border-radius: 12px;
    max-width: 210mm;
    margin: 0 auto;
    padding: 3rem 1.5rem;
    text-align: center;
    color: #6b7280;
    border: 1px dashed #d1d5db;
  }

  .settings-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 0.75rem 1rem;
    font-size: 0.85rem;
    font-weight: 600;
    color: #4b5563;
  }

  .form-input {
    display: block;
    width: 100%;
    margin-top: 0.25rem;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    padding: 0.4rem 0.6rem;
    font-size: 0.85rem;
    font-weight: normal;
  }

  /* A4 CONTAINER MIRIP KESEHARIAN */
  .a4-container {
    background: white;
    width: 100%;
    max-width: 210mm;
    margin: 0 auto 2.5rem;
    padding: 1.2cm !important;
    box-sizing: border-box;
    box-shadow: 0 8px 30px rgba(0, 0, 0, 0.12);
    font-family: Arial, Helvetica, sans-serif;
    font-size: 13.5px;
    color: black;
    page-break-after: always;
    break-after: page;
  }

  .title-section {
    text-align: center;
    margin-bottom: 20px;
  }

  .title-main {
    font-weight: 700;
    font-size: 16px;
    text-decoration: underline;
    text-underline-offset: 4px;
    margin: 0;
  }

  .title-sub {
    font-size: 12.5px;
    font-weight: 700;
    margin: 4px 0 0;
    color: #374151;
    letter-spacing: 0.5px;
  }

  .student-info {
    display: flex;
    justify-content: space-between;
    margin-bottom: 16px;
    font-size: 13px;
    font-weight: 600;
  }

  .info-table {
    border: none;
    width: auto;
  }

  .info-table td {
    border: none;
    padding: 2px 0;
    vertical-align: top;
  }

  .info-label {
    width: 120px;
  }

  .info-sep {
    padding: 2px 8px;
  }

  .rapor-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 12.5px;
    border-top: 1px solid black;
    border-left: 1px solid black;
  }

  .rapor-table th,
  .rapor-table td {
    border: none;
    border-bottom: 1px solid black;
    border-right: 1px solid black;
    padding: 5px 8px;
    vertical-align: middle;
  }

  .rapor-table th {
    font-weight: 700;
    text-align: center;
    background-color: #f9fafb;
  }

  .catatan-box {
    border: 1px solid black;
    padding: 8px 12px;
    margin-top: 12px;
    margin-bottom: 20px;
  }

  .signature-section {
    display: flex;
    justify-content: space-between;
    margin-top: 24px;
    text-align: center;
    font-size: 13px;
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

  /* DOWNLOAD PROGRESS OVERLAY */
  .dl-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(6px);
    z-index: 99999;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .dl-modal {
    background: white;
    border-radius: 20px;
    padding: 2.5rem 3rem;
    max-width: 420px;
    width: 90%;
    text-align: center;
    box-shadow: 0 25px 60px rgba(0, 0, 0, 0.3);
  }

  .dl-icon {
    margin-bottom: 1.25rem;
    display: flex;
    justify-content: center;
  }

  .dl-spinner {
    width: 48px;
    height: 48px;
    border: 4px solid #e2e8f0;
    border-top-color: #2563eb;
    border-radius: 50%;
    animation: dlSpin 0.8s linear infinite;
  }

  @keyframes dlSpin {
    to {
      transform: rotate(360deg);
    }
  }

  .dl-title {
    font-size: 1.25rem;
    font-weight: 700;
    color: #1e293b;
    margin: 0 0 0.5rem;
  }

  .dl-subtitle {
    font-size: 0.9rem;
    color: #64748b;
    margin: 0 0 1.25rem;
    line-height: 1.5;
  }

  .dl-progress-wrap {
    width: 100%;
    height: 10px;
    background: #e2e8f0;
    border-radius: 999px;
    overflow: hidden;
  }

  .dl-progress-bar {
    height: 100%;
    background: linear-gradient(90deg, #2563eb, #3b82f6);
    border-radius: 999px;
    transition: width 0.4s ease;
  }

  .dl-progress-indeterminate {
    width: 40% !important;
    animation: dlIndeterminate 1.5s ease-in-out infinite;
  }

  @keyframes dlIndeterminate {
    0% {
      transform: translateX(-100%);
    }
    100% {
      transform: translateX(350%);
    }
  }

  .dl-progress-text {
    font-size: 0.8rem;
    color: #94a3b8;
    margin: 0.75rem 0 0;
    font-weight: 600;
  }

  @media print {
    :global(body) {
      background: white !important;
    }
    .print-wrapper {
      background: white !important;
      padding: 0 !important;
    }
    .no-print {
      display: none !important;
    }
    .a4-container {
      box-shadow: none !important;
      margin: 0 auto !important;
      padding: 0.8cm !important;
      max-width: 100% !important;
      page-break-after: always !important;
      break-after: page !important;
    }
  }
</style>
