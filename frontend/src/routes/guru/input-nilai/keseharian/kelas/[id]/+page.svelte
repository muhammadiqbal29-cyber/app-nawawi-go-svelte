<script>
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import Icon from '$lib/components/ui/Icon.svelte';
  import { guruState } from '$lib/guruState.svelte.js';
  import { toastStore } from '$lib/toast.svelte.js';
  import { modalStore } from '$lib/modal.svelte.js';

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

  let kelasId = $derived($page.params.id);
  let sekolahId = $derived(guruState.activeSekolahId);
  let namaKelas = $derived($page.url.searchParams.get('kelas') || 'Kelas');
  let rombel = $derived($page.url.searchParams.get('rombel') || '');
  let displayKelasName = $derived(
    rombel && rombel !== '__kosong__' ? `${namaKelas} - Rombel ${rombel}` : namaKelas
  );


  // ===== TAB STATE =====
  let activeTab = $state('keseharian');

  // ===== MODAL STATE =====
  let showStudentModal = $state(false);
  let selectedStudent = $state(null);
  let modalActiveTab = $state('pelanggaran');

  function openStudentModal(student) {
    selectedStudent = student;
    plgFormSiswaId = student.siswa_id;
    iqobFormSiswaId = student.siswa_id;
    catatanFormSiswaId = student.siswa_id;
    showStudentModal = true;
    modalActiveTab = 'pelanggaran';
    
    // reset form inputs
    plgFormJenisId = '';
    plgFormPoin = 0;
    plgSearchQuery = '';
    plgShowDropdown = true;
    plgSelectedNama = '';
    iqobFormJenisId = '';
    iqobFormPoin = 0;
    iqobSearchQuery = '';
    iqobShowDropdown = true;
    iqobSelectedNama = '';
    
    // load existing catatan
    const ex = catatanRaporMap[student.siswa_id];
    catatanFormText = ex ? ex.catatan : '';
  }

  function closeStudentModal() {
    showStudentModal = false;
    selectedStudent = null;
  }


  // ===== KESEHARIAN POIN STATE =====
  let indikatorList = $state([]);
  let selectedIndikator = $state('');
  let siswaData = $state([]);
  let searchQuery = $state('');
  let filteredSiswaData = $derived(
    siswaData.filter(s => (s.nama_lengkap || '').toLowerCase().includes(searchQuery.toLowerCase()))
  );
  let rekapData = $state({});
  let loading = $state(false);
  let error = $state('');
  let saveStatus = $state('');

  // ===== PELANGGARAN STATE =====
  let jenisPelanggaranList = $state([]);
  let pelanggaranSiswaList = $state([]);
  let filteredPelanggaranSiswaList = $derived(
    pelanggaranSiswaList.filter(s => (s.nama_lengkap || '').toLowerCase().includes(searchQuery.toLowerCase()))
  );
  let pelanggaranCatatanMap = $state({});
  let loadingPelanggaran = $state(false);
  let errorPelanggaran = $state('');

  // Form input pelanggaran
  let showFormPelanggaran = $state(false);
  let plgFormSiswaId = $state('');
  let plgFormJenisId = $state('');
  let plgFormTanggal = $state(new Date().toISOString().split('T')[0]);
  let plgFormPoin = $state(0);
  let plgFormIqob = $state('');
  
  let plgSearchQuery = $state('');
  let plgShowDropdown = $state(true);
  let plgSelectedNama = $state('');
  let plgFilteredOptions = $derived(
    jenisPelanggaranList.filter(j => j.nama.toLowerCase().includes(plgSearchQuery.toLowerCase()))
  );
  let plgFormLoading = $state(false);
  let plgFormError = $state('');
  let pendingPelanggaranList = $state([]);


  // ===== IQOB STATE =====
  let jenisIqobList = $state([]);
  let iqobSiswaList = $state([]);
  let iqobCatatanMap = $state({});
  let loadingIqob = $state(false);
  let errorIqob = $state('');

  // Form input iqob
  let showFormIqob = $state(false);
  let iqobFormSiswaId = $state('');
  let iqobFormJenisId = $state('');
  let iqobFormTanggal = $state(new Date().toISOString().split('T')[0]);
  let iqobFormPoin = $state(0);
  
  let iqobSearchQuery = $state('');
  let iqobShowDropdown = $state(true);
  let iqobSelectedNama = $state('');
  let iqobFilteredOptions = $derived(
    jenisIqobList.filter(j => j.nama.toLowerCase().includes(iqobSearchQuery.toLowerCase()))
  );
  let iqobFormLoading = $state(false);
  let iqobFormError = $state('');
  let pendingIqobList = $state([]);
  let pendingIqobPoints = $derived(
    pendingIqobList.reduce((sum, item) => sum + Number(item.poin), 0)
  );
  let recommendedIqob = $derived((() => {
    if (!selectedStudent || !jenisIqobList || jenisIqobList.length === 0) return null;
    const currentPoints = getPoinAkhir(selectedStudent.siswa_id) - pendingIqobPoints;
    if (currentPoints <= 0) return null;
    
    const available = [...jenisIqobList].filter(j => j.aktif !== false).sort((a, b) => a.poin - b.poin);
    if (available.length === 0) return null;

    const exact = available.find(j => j.poin === currentPoints);
    if (exact) return exact;

    if (currentPoints < available[0].poin) return available[0];

    let closest = available[0];
    let minDiff = Math.abs(available[0].poin - currentPoints);
    for (const j of available) {
      const diff = Math.abs(j.poin - currentPoints);
      if (diff < minDiff) {
        minDiff = diff;
        closest = j;
      }
    }
    return closest;
  })());

  // ===== CATATAN RAPOR STATE =====
  let catatanRaporMap = $state({});
  let loadingCatatan = $state(false);
  let errorCatatan = $state('');
  
  // Form input catatan
  let catatanFormSiswaId = $state('');
  let catatanFormText = $state('');
  let catatanFormLoading = $state(false);
  let catatanFormError = $state('');

  $effect(() => {
    if (sekolahId) {
      loadIndikator();
      loadJenisPelanggaran();
      loadJenisIqob();
    }
  });

  $effect(() => {
    if (sekolahId && kelasId && selectedIndikator) {
      fetchRekap();
    }
  });

  // ===== KESEHARIAN FUNCTIONS =====
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
      if (guruState.activeTahunId) url.searchParams.append('tahun_ajaran_id', guruState.activeTahunId);
      if (guruState.activeSemesterId) url.searchParams.append('semester_id', guruState.activeSemesterId);
      if (rombel && rombel !== '__kosong__') url.searchParams.append('rombel', rombel);
      else if (rombel === '__kosong__') url.searchParams.append('rombel', '__kosong__');

      const res = await fetch(url.toString(), { credentials: 'include' });
      if (!res.ok) throw new Error('Gagal memuat rekap keseharian.');

      const body = await res.json();
      siswaData = body.data?.siswa || [];
      const rekaps = body.data?.rekap || [];

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

  async function handleSavePoin(siswaId, newPoin) {
    const poinVal = Number(newPoin);
    if (isNaN(poinVal)) return;

    const r = rekapData[siswaId];
    if (r) r.saving = true;

    try {
      const res = await fetch(`${apiUrl}/api/guru/sekolah/${sekolahId}/keseharian/poin`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          siswa_id: siswaId,
          indikator_id: Number(selectedIndikator),
          poin_akhir: poinVal
        })
      });
      if (!res.ok) throw new Error('Gagal simpan poin');

      const body = await res.json();
      const updated = body.data;
      rekapData[siswaId] = {
        ...(rekapData[siswaId] || {}),
        id: updated.id,
        poin_murni: updated.poin_murni,
        poin_akhir: updated.poin_akhir,
        poin_penyesuaian: updated.poin_penyesuaian,
        saving: false
      };

      saveStatus = 'Tersimpan!';
      setTimeout(() => (saveStatus = ''), 2000);
    } catch (err) {
      toastStore.show(err.message, 'info');
    } finally {
      if (rekapData[siswaId]) rekapData[siswaId].saving = false;
    }
  }

  // ===== PELANGGARAN FUNCTIONS =====
  async function loadJenisPelanggaran() {
    try {
      const res = await fetch(`${apiUrl}/api/guru/sekolah/${sekolahId}/pelanggaran/jenis`, {
        credentials: 'include'
      });
      if (res.ok) {
        const body = await res.json();
        jenisPelanggaranList = (body.data || []).filter(j => j.aktif);
      }
    } catch (err) {
      console.error(err);
    }
  }

  async function loadPelanggaranKelas() {
    if (!kelasId) return;
    loadingPelanggaran = true;
    errorPelanggaran = '';

    try {
      const url = new URL(`${apiUrl}/api/guru/kelas/${kelasId}/pelanggaran`);
      url.searchParams.append('sekolah_id', sekolahId);
      if (guruState.activeTahunId) url.searchParams.append('tahun_ajaran_id', guruState.activeTahunId);
      if (guruState.activeSemesterId) url.searchParams.append('semester_id', guruState.activeSemesterId);
      if (rombel && rombel !== '__kosong__') url.searchParams.append('rombel', rombel);
      else if (rombel === '__kosong__') url.searchParams.append('rombel', '__kosong__');

      const res = await fetch(url.toString(), { credentials: 'include' });
      if (!res.ok) throw new Error('Gagal memuat pelanggaran kelas');

      const body = await res.json();
      pelanggaranSiswaList = body.data?.siswa || [];
      pelanggaranCatatanMap = body.data?.catatan || {};
    } catch (err) {
      errorPelanggaran = err.message;
    } finally {
      loadingPelanggaran = false;
    }
  }

  $effect(() => {
    if (activeTab === 'pelanggaran_iqob' && sekolahId && kelasId) {
      loadPelanggaranKelas();
      loadIqobKelas();
          }
  });

  
  function formatDate(dateStr) {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    return date.toLocaleDateString('id-ID', { day: '2-digit', month: '2-digit', year: 'numeric' });
  }

  function selectPelanggaran(j) {
    plgFormJenisId = j.id;
    plgSelectedNama = j.nama;
    plgSearchQuery = '';
    plgShowDropdown = false;
    plgFormPoin = j.poin;
  }

  function tambahPendingPelanggaran() {
    if (!plgFormJenisId || !plgFormTanggal) {
      plgFormError = 'Pilih jenis pelanggaran dan tanggal terlebih dahulu';
      return;
    }
    plgFormError = '';
    pendingPelanggaranList = [...pendingPelanggaranList, {
      jenis_pelanggaran_id: plgFormJenisId,
      nama: plgSelectedNama,
      tanggal: plgFormTanggal,
      poin: plgFormPoin,
      iqob: plgFormIqob
    }];
    plgFormJenisId = '';
    plgFormPoin = 0;
    plgSearchQuery = '';
    plgShowDropdown = true;
    plgSelectedNama = '';
  }

  function removePendingPelanggaran(index) {
    pendingPelanggaranList = pendingPelanggaranList.filter((_, i) => i !== index);
  }

  async function submitPelanggaran() {
    if (pendingPelanggaranList.length === 0) {
      plgFormError = 'Tambahkan setidaknya satu pelanggaran ke daftar';
      return;
    }
    plgFormLoading = true;
    plgFormError = '';

    try {
      const url = new URL(`${apiUrl}/api/guru/pelanggaran/catatan`);
      if (guruState.activeTahunId) url.searchParams.append('tahun_ajaran_id', guruState.activeTahunId);
      if (guruState.activeSemesterId) url.searchParams.append('semester_id', guruState.activeSemesterId);

      const promises = pendingPelanggaranList.map(item => {
        return fetch(url.toString(), {
          method: 'POST',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            siswa_id: Number(plgFormSiswaId),
            sekolah_id: Number(sekolahId),
            jenis_pelanggaran_id: Number(item.jenis_pelanggaran_id),
            tanggal: item.tanggal,
            poin: Number(item.poin),
            iqob: item.iqob || ''
          })
        }).then(async r => {
          if (!r.ok) {
            const b = await r.json().catch(() => ({}));
            throw new Error(b.message || `Gagal menambah pelanggaran: ${item.nama}`);
          }
        });
      });

      await Promise.all(promises);

      toastStore.show('Pelanggaran berhasil dicatat', 'success');
      pendingPelanggaranList = [];
      await loadPelanggaranKelas();
    } catch (err) {
      plgFormError = err.message;
    } finally {
      plgFormLoading = false;
    }
  }

  async function hapusCatatan(id) {
    if (!(await modalStore.show({ message: 'Hapus catatan pelanggaran ini?' }))) return;
    try {
      const res = await fetch(`${apiUrl}/api/guru/pelanggaran/catatan/${id}`, {
        method: 'DELETE',
        credentials: 'include'
      });
      if (!res.ok) throw new Error('Gagal menghapus');
      await loadPelanggaranKelas();
    } catch (err) {
      toastStore.show(err.message, 'info');
    }
  }


  // ===== IQOB FUNCTIONS =====
  async function loadJenisIqob() {
    try {
      const res = await fetch(`${apiUrl}/api/guru/sekolah/${sekolahId}/iqob/jenis`, {
        credentials: 'include'
      });
      if (res.ok) {
        const body = await res.json();
        jenisIqobList = (body.data || []).filter(j => j.aktif);
      }
    } catch (err) {
      console.error(err);
    }
  }

  async function loadIqobKelas() {
    if (!kelasId) return;
    loadingIqob = true;
    errorIqob = '';

    try {
      const url = new URL(`${apiUrl}/api/guru/kelas/${kelasId}/iqob`);
      url.searchParams.append('sekolah_id', sekolahId);
      if (guruState.activeTahunId) url.searchParams.append('tahun_ajaran_id', guruState.activeTahunId);
      if (guruState.activeSemesterId) url.searchParams.append('semester_id', guruState.activeSemesterId);
      if (rombel && rombel !== '__kosong__') url.searchParams.append('rombel', rombel);
      else if (rombel === '__kosong__') url.searchParams.append('rombel', '__kosong__');

      const res = await fetch(url.toString(), { credentials: 'include' });
      if (!res.ok) throw new Error('Gagal memuat iqob kelas');

      const body = await res.json();
      iqobSiswaList = body.data?.siswa || [];
      iqobCatatanMap = body.data?.catatan || {};
    } catch (err) {
      errorIqob = err.message;
    } finally {
      loadingIqob = false;
    }
  }



  function selectIqob(j) {
    iqobFormJenisId = j.id;
    iqobFormPoin = j.poin;
    iqobSelectedNama = j.nama;
    iqobShowDropdown = false;
    iqobSearchQuery = '';
  }

  function applyRecommendedIqob() {
    if (recommendedIqob) {
      iqobFormJenisId = recommendedIqob.id;
      iqobFormPoin = recommendedIqob.poin;
      iqobSelectedNama = recommendedIqob.nama;
      iqobShowDropdown = false;
      iqobSearchQuery = '';
    }
  }

  function tambahPendingIqob() {
    if (!iqobFormJenisId || !iqobFormTanggal) {
      iqobFormError = 'Pilih jenis iqob dan tanggal terlebih dahulu';
      return;
    }
    iqobFormError = '';
    pendingIqobList = [...pendingIqobList, {
      jenis_iqob_id: iqobFormJenisId,
      nama: iqobSelectedNama,
      tanggal: iqobFormTanggal,
      poin: iqobFormPoin
    }];
    iqobFormJenisId = '';
    iqobFormPoin = 0;
    iqobSearchQuery = '';
    iqobShowDropdown = true;
    iqobSelectedNama = '';
  }

  function removePendingIqob(index) {
    pendingIqobList = pendingIqobList.filter((_, i) => i !== index);
  }

  async function submitIqob() {
    if (pendingIqobList.length === 0) {
      iqobFormError = 'Tambahkan setidaknya satu iqob ke daftar';
      return;
    }
    iqobFormLoading = true;
    iqobFormError = '';

    try {
      const url = new URL(`${apiUrl}/api/guru/iqob/catatan`);
      if (guruState.activeTahunId) url.searchParams.append('tahun_ajaran_id', guruState.activeTahunId);
      if (guruState.activeSemesterId) url.searchParams.append('semester_id', guruState.activeSemesterId);

      const promises = pendingIqobList.map(item => {
        return fetch(url.toString(), {
          method: 'POST',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            siswa_id: Number(iqobFormSiswaId),
            sekolah_id: Number(sekolahId),
            jenis_iqob_id: Number(item.jenis_iqob_id),
            tanggal: item.tanggal,
            poin: Number(item.poin)
          })
        }).then(async r => {
          if (!r.ok) {
            const b = await r.json().catch(() => ({}));
            throw new Error(b.message || `Gagal menambah iqob: ${item.nama}`);
          }
        });
      });

      await Promise.all(promises);

      toastStore.show('Iqob berhasil dicatat', 'success');
      pendingIqobList = [];
      await loadIqobKelas();
    } catch (err) {
      iqobFormError = err.message;
    } finally {
      iqobFormLoading = false;
    }
  }

  async function hapusCatatanIqob(id) {
    if (!(await modalStore.show({ message: 'Hapus catatan iqob ini?' }))) return;
    try {
      const res = await fetch(`${apiUrl}/api/guru/iqob/catatan/${id}`, {
        method: 'DELETE',
        credentials: 'include'
      });
      if (!res.ok) throw new Error('Gagal menghapus');
      await loadIqobKelas();
      toastStore.show('Catatan iqob dihapus', 'info');
    } catch (err) {
      alert(err.message);
    }
  }

  function getTotalPoinPelanggaran(siswaId) {
    const list = pelanggaranCatatanMap[siswaId] || [];
    return list.reduce((acc, c) => acc + (c.poin || 0), 0);
  }

  function getTotalPoinIqob(siswaId) {
    const list = iqobCatatanMap[siswaId] || [];
    return list.reduce((acc, c) => acc + (c.poin || 0), 0);
  }

  function getPoinAkhir(siswaId) {
    return getTotalPoinPelanggaran(siswaId) - getTotalPoinIqob(siswaId);
  }
</script>

<svelte:head>
  <title>Input Keseharian - SIAKAD</title>
</svelte:head>

<div class="page-container">
  <button class="btn-back" onclick={() => goto('/guru/input-nilai/keseharian')}>
    <Icon name="chevronLeft" size={18} /> Kembali ke Daftar Kelas
  </button>

  <div class="header-action" style="margin-top: 1.5rem; display: flex; justify-content: space-between; align-items: flex-end; flex-wrap: wrap; gap: 1rem;">
    <div>
      <h2>Input Keseharian: {displayKelasName}</h2>
      <p class="subtitle">
        Input poin keseharian dan catat pelanggaran siswa.
      </p>
    </div>
    <div class="search-bar">
      <Icon name="search" size={16} class="search-icon" />
      <input type="text" bind:value={searchQuery} placeholder="Cari nama siswa..." class="search-input" />
    </div>
  </div>

  <!-- Tab Navigation -->
  <div class="tabs">
    <button
      class="tab-btn"
      class:active={activeTab === 'keseharian'}
      onclick={() => (activeTab = 'keseharian')}
    >
      Poin Keseharian
    </button>
    <button
      class="tab-btn"
      class:active={activeTab === 'pelanggaran_iqob'}
      onclick={() => (activeTab = 'pelanggaran_iqob')}
    >
      Pelanggaran & Iqob
    </button>
  </div>

  <!-- ===== TAB: POIN KESEHARIAN ===== -->
  {#if activeTab === 'keseharian'}
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
              <th width="35%">Nama Siswa</th>
              <th width="20%">Poin (Input Langsung)</th>
              <th width="15%">Nilai Akhir (Rapor)</th>
            </tr>
          </thead>
          <tbody>
            {#each filteredSiswaData as s, i}
              {@const r = rekapData[s.siswa_id]}
              <tr>
                <td>{i + 1}</td>
                <td class="font-medium"
                  >{s.nama_lengkap}
                  <div class="text-sm text-gray-500">{s.nisn}</div>
                </td>
                <td>
                  <input
                    type="number"
                    class="input input-sm"
                    min="0"
                    max="100"
                    value={r ? r.poin_akhir : ''}
                    onchange={(e) => handleSavePoin(s.siswa_id, e.target.value)}
                    disabled={r?.saving}
                    placeholder="0 - 100"
                  />
                </td>
                <td>
                  {#if r}
                    <div class="final-score" class:saving={r.saving}>
                      {Number(r.poin_akhir).toFixed(0)}
                    </div>
                  {:else}
                    <span class="badge-gray">-</span>
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
  {/if}

  {#if activeTab === 'pelanggaran_iqob'}
    <!-- TAB PELANGGARAN & IQOB (INDEX) -->
    <div class="card slide-up" style="margin-top: 1rem;">
      <div class="flex" style="justify-content: space-between; align-items: center; margin-bottom: 1rem;">
        <h3 style="margin: 0;">Rekap Pelanggaran & Iqob</h3>
        <p class="text-muted" style="margin: 0; font-size: 0.9em;">Klik pada siswa untuk melihat detail & menambah catatan.</p>
      </div>

      {#if errorPelanggaran || errorIqob}
        <div class="alert alert-error">{errorPelanggaran || errorIqob}</div>
      {/if}

      {#if loadingPelanggaran || loadingIqob}
        <div class="loading-state">Memuat data kelas...</div>
      {:else if pelanggaranSiswaList.length > 0}
        <div class="table-container">
          <table class="data-table">
            <thead>
              <tr>
                <th width="5%">No</th>
                <th width="35%">Nama Siswa</th>
                <th width="15%" style="text-align: center;">Poin Pelanggaran</th>
                <th width="15%" style="text-align: center;">Poin Iqob</th>
                <th width="15%" style="text-align: center;">Total Poin Akhir</th>
                <th width="15%" style="text-align: center;">Aksi</th>
              </tr>
            </thead>
            <tbody>
              {#each filteredPelanggaranSiswaList as s, i}
                {@const plg = getTotalPoinPelanggaran(s.siswa_id)}
                {@const iqob = getTotalPoinIqob(s.siswa_id)}
                {@const akhir = getPoinAkhir(s.siswa_id)}
                <tr>
                  <td>{i + 1}</td>
                  <td class="font-medium">{s.nama_lengkap}<br /><span style="font-size: 0.8em; color: #64748b;">{s.nisn}</span></td>
                  <td style="text-align: center;">
                    {#if plg > 0}
                      <span class="badge-red" style="font-size: 1rem; padding: 4px 10px;">{plg}</span>
                    {:else}
                      <span class="text-muted">0</span>
                    {/if}
                  </td>
                  <td style="text-align: center;">
                    {#if iqob > 0}
                      <span class="badge-green" style="font-size: 1rem; padding: 4px 10px;">-{iqob}</span>
                    {:else}
                      <span class="text-muted">0</span>
                    {/if}
                  </td>
                  <td style="text-align: center; font-weight: bold;">
                    {#if akhir > 0}
                      <span style="color: #dc2626;">{akhir}</span>
                    {:else if akhir < 0}
                      <span style="color: #16a34a;">{akhir}</span>
                    {:else}
                      0
                    {/if}
                  </td>
                  <td style="text-align: center;">
                    <button class="btn-primary" style="padding: 0.4rem 0.8rem; font-size: 0.85rem;" onclick={() => openStudentModal(s)}>
                      Detail & Input
                    </button>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {:else}
        <div class="empty-state">Belum ada data siswa.</div>
      {/if}
    </div>
  {/if}
</div>

<!-- ================= MODAL KUSTOM INTERNAL ================= -->
{#if showStudentModal && selectedStudent}
  <div class="modal-backdrop fade-in" onclick={closeStudentModal}>
    <div class="modal-content slide-up" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <div style="display:flex; align-items: center; gap: 12px;">
          <div class="avatar-placeholder">
            <Icon name="user" size={24} color="#64748b" />
          </div>
          <div>
            <h3 style="margin: 0; font-size: 1.2rem; color: #1e293b;">{selectedStudent.nama_lengkap}</h3>
            <span style="font-size: 0.85rem; color: #64748b;">NISN: {selectedStudent.nisn}</span>
          </div>
        </div>
        <button class="btn-icon-mini" style="background: none; border: none; font-size: 1.5rem; cursor: pointer; color: #94a3b8;" onclick={closeStudentModal}>&times;</button>
      </div>

      <div class="modal-body">
        <!-- Skor Header -->
        <div class="score-cards" style="display: flex; gap: 1rem; margin-bottom: 1.5rem;">
          <div class="score-card danger">
            <div class="score-label">Poin Pelanggaran</div>
            <div class="score-value">{getTotalPoinPelanggaran(selectedStudent.siswa_id)}</div>
          </div>
          <div class="score-card success">
            <div class="score-label">Poin Iqob</div>
            <div class="score-value">-{getTotalPoinIqob(selectedStudent.siswa_id)}</div>
          </div>
          <div class="score-card primary">
            <div class="score-label">Total Poin Akhir</div>
            <div class="score-value">{getPoinAkhir(selectedStudent.siswa_id)}</div>
          </div>
        </div>

        <div class="tabs" style="margin-bottom: 1rem;">
          <button class="tab-btn" class:active={modalActiveTab === 'pelanggaran'} onclick={() => (modalActiveTab = 'pelanggaran')}>
            Riwayat Pelanggaran
          </button>
          <button class="tab-btn" class:active={modalActiveTab === 'iqob'} onclick={() => (modalActiveTab = 'iqob')}>
            Riwayat Iqob
          </button>

        </div>

        <!-- ISI TAB MODAL: PELANGGARAN -->
        {#if modalActiveTab === 'pelanggaran'}
          <div class="modal-form-section" style="background: #f8fafc; padding: 1rem; border-radius: 8px; border: 1px solid #e2e8f0; margin-bottom: 1.5rem;">
            <h4 style="margin: 0 0 1rem 0; font-size: 0.95rem; display:flex; align-items:center; gap:6px;"><Icon name="plus" size={16}/> Tambah Pelanggaran</h4>
            
            <div style="display: flex; gap: 1rem; flex-wrap: wrap;">
              <div class="form-group" style="position: relative; flex: 2; min-width: 200px;">
                <label>Jenis Pelanggaran</label>
                {#if !plgShowDropdown && plgSelectedNama}
                  <div style="display: flex; gap: 0.5rem; align-items: center;">
                    <div class="input" style="flex: 1; background: #e2e8f0; color: #334155; font-weight: 500; font-size: 0.9rem; line-height: 1.4;">
                      {plgSelectedNama} <span style="color: #dc2626;">(+{plgFormPoin})</span>
                    </div>
                    <button class="btn-secondary" style="padding: 0.5rem; height: 100%;" onclick={() => { plgShowDropdown = true; plgSelectedNama = ''; plgFormJenisId = ''; plgFormPoin = 0; }}>
                      Ubah
                    </button>
                  </div>
                {:else}
                  <input type="text" class="input" placeholder="Ketik jenis pelanggaran..." bind:value={plgSearchQuery} />
                  {#if plgSearchQuery.length > 0}
                    <div style="position: absolute; top: 100%; left: 0; right: 0; background: white; border: 1px solid #cbd5e1; border-radius: 4px; max-height: 250px; overflow-y: auto; z-index: 50; box-shadow: 0 4px 12px rgba(0,0,0,0.15); margin-top: 4px;">
                      {#if plgFilteredOptions.length === 0}
                        <div style="padding: 0.75rem; color: #64748b; font-size: 0.9rem; text-align: center;">Tidak ada hasil ditemukan.</div>
                      {:else}
                        {#each plgFilteredOptions as j}
                          <!-- svelte-ignore a11y_click_events_have_key_events -->
                          <!-- svelte-ignore a11y_no_static_element_interactions -->
                          <div 
                            style="padding: 0.75rem; border-bottom: 1px solid #f1f5f9; cursor: pointer; font-size: 0.9rem; transition: background 0.1s;"
                            onmouseover={(e) => e.currentTarget.style.background = '#f1f5f9'}
                            onmouseout={(e) => e.currentTarget.style.background = 'transparent'}
                            onclick={() => selectPelanggaran(j)}
                          >
                            <div style="font-weight: 500; color: #0f172a; margin-bottom: 0.25rem;">{j.nama}</div>
                            <div style="color: #dc2626; font-size: 0.8rem; font-weight: 700;">+{j.poin} Poin</div>
                          </div>
                        {/each}
                      {/if}
                    </div>
                  {/if}
                {/if}
              </div>
              <div class="form-group" style="flex: 1; min-width: 120px;">
                <label for="plg-tgl-modal">Tanggal</label>
                <input id="plg-tgl-modal" type="date" class="input" bind:value={plgFormTanggal} />
              </div>
              <div class="form-group" style="flex: 1; min-width: 100px;">
                <label for="plg-poin-modal">Poin</label>
                <input id="plg-poin-modal" type="number" step="0.1" class="input" bind:value={plgFormPoin} />
              </div>
            </div>
            <div style="display: flex; justify-content: flex-end; margin-top: 1rem;">
              <button class="btn-secondary" onclick={tambahPendingPelanggaran} disabled={plgFormLoading}>
                <Icon name="plus" size={14} /> Tambah ke Daftar
              </button>
            </div>
            {#if plgFormError}
              <div class="alert-error mt-2">{plgFormError}</div>
            {/if}

            {#if pendingPelanggaranList.length > 0}
              <div style="margin-top: 1.5rem; background: #ffffff; border: 1px dashed #cbd5e1; border-radius: 8px; padding: 1rem;">
                <h5 style="margin: 0 0 0.75rem 0; font-size: 0.9rem; color: #475569;">Daftar Pelanggaran Baru (Belum Disimpan)</h5>
                {#each pendingPelanggaranList as p, i}
                  <div style="display: flex; justify-content: space-between; align-items: center; padding: 0.5rem; border-bottom: 1px solid #f1f5f9;">
                    <div>
                      <span style="font-weight: 500; font-size: 0.9rem;">{p.nama}</span>
                      <span style="color: #64748b; font-size: 0.8rem; margin-left: 8px;">{formatDate(p.tanggal)}</span>
                      <span style="color: #dc2626; font-size: 0.8rem; font-weight: 600; margin-left: 8px;">(+{p.poin})</span>
                    </div>
                    <button class="btn-icon-mini text-red" title="Hapus" onclick={() => removePendingPelanggaran(i)}>
                      <Icon name="x" size={14} />
                    </button>
                  </div>
                {/each}
                <div style="display: flex; justify-content: flex-end; margin-top: 1rem;">
                  <button class="btn-primary" onclick={submitPelanggaran} disabled={plgFormLoading}>
                    {plgFormLoading ? 'Menyimpan...' : `Simpan Semua (${pendingPelanggaranList.length})`}
                  </button>
                </div>
              </div>
            {/if}
          </div>

          {@const plgHistory = pelanggaranCatatanMap[selectedStudent.siswa_id] || []}
          {#if plgHistory.length > 0}
            <div class="history-list">
              {#each plgHistory as c}
                <div class="history-item">
                  <div class="history-info">
                    <span class="history-date">{formatDate(c.tanggal)}</span>
                    <span class="history-desc">{c.nama_pelanggaran} <span style="color: #dc2626; font-weight: 600;">(+{c.poin})</span></span>
                  </div>
                  <button class="btn-icon-mini text-red" title="Hapus" onclick={() => hapusCatatan(c.id)}>
                    <Icon name="x" size={14} />
                  </button>
                </div>
              {/each}
            </div>
          {:else}
            <div class="text-muted" style="text-align: center; padding: 2rem;">Tidak ada catatan pelanggaran.</div>
          {/if}

        <!-- ISI TAB MODAL: IQOB -->
        {:else if modalActiveTab === 'iqob'}
          <div class="modal-form-section" style="background: #f8fafc; padding: 1rem; border-radius: 8px; border: 1px solid #e2e8f0; margin-bottom: 1.5rem;">
            <h4 style="margin: 0 0 1rem 0; font-size: 0.95rem; display:flex; align-items:center; gap:6px;"><Icon name="plus" size={16}/> Tambah Iqob (Pengurang Poin)</h4>
            
            <div style="display: flex; gap: 1rem; flex-wrap: wrap;">
              <div class="form-group" style="position: relative; flex: 2; min-width: 200px;">
                <label>Jenis Iqob</label>
                {#if !iqobShowDropdown && iqobSelectedNama}
                  <div style="display: flex; gap: 0.5rem; align-items: center;">
                    <div class="input" style="flex: 1; background: #e2e8f0; color: #334155; font-weight: 500; font-size: 0.9rem; line-height: 1.4;">
                      {iqobSelectedNama} <span style="color: #16a34a;">(-{iqobFormPoin})</span>
                    </div>
                    <button class="btn-secondary" style="padding: 0.5rem; height: 100%;" onclick={() => { iqobShowDropdown = true; iqobSelectedNama = ''; iqobFormJenisId = ''; iqobFormPoin = 0; }}>
                      Ubah
                    </button>
                  </div>
                {:else}
                  {#if recommendedIqob}
                    <div style="margin-bottom: 0.5rem; background: #e0f2fe; border: 1px solid #bae6fd; padding: 0.5rem; border-radius: 4px; display: flex; align-items: center; justify-content: space-between;">
                      <div>
                        <span style="font-size: 0.8rem; font-weight: 600; color: #0284c7; display: block; margin-bottom: 2px;">💡 Rekomendasi Iqob (Poin Pelanggaran Tersisa: {getPoinAkhir(selectedStudent.siswa_id) - pendingIqobPoints}):</span>
                        <span style="font-size: 0.85rem; color: #0f172a; font-weight: 500;">{recommendedIqob.nama}</span>
                        <span style="color: #16a34a; font-size: 0.85rem; font-weight: 600; margin-left: 4px;">(-{recommendedIqob.poin})</span>
                      </div>
                      <button class="btn-secondary" style="padding: 0.25rem 0.5rem; font-size: 0.75rem;" onclick={applyRecommendedIqob}>
                        Gunakan
                      </button>
                    </div>
                  {/if}
                  <input type="text" class="input" placeholder="Ketik jenis iqob..." bind:value={iqobSearchQuery} />
                  {#if iqobSearchQuery.length > 0}
                    <div style="position: absolute; top: 100%; left: 0; right: 0; background: white; border: 1px solid #cbd5e1; border-radius: 4px; max-height: 250px; overflow-y: auto; z-index: 50; box-shadow: 0 4px 12px rgba(0,0,0,0.15); margin-top: 4px;">
                      {#if iqobFilteredOptions.length === 0}
                        <div style="padding: 0.75rem; color: #64748b; font-size: 0.9rem; text-align: center;">Tidak ada hasil ditemukan.</div>
                      {:else}
                        {#each iqobFilteredOptions as j}
                          <!-- svelte-ignore a11y_click_events_have_key_events -->
                          <!-- svelte-ignore a11y_no_static_element_interactions -->
                          <div
                            style="padding: 0.75rem; border-bottom: 1px solid #f1f5f9; cursor: pointer; font-size: 0.9rem; transition: background 0.1s;"
                            onmouseover={(e) => e.currentTarget.style.background = '#f1f5f9'}
                            onmouseout={(e) => e.currentTarget.style.background = 'transparent'}
                            onclick={() => selectIqob(j)}
                          >
                            <div style="font-weight: 500; color: #0f172a; margin-bottom: 0.25rem;">{j.nama}</div>
                            <div style="color: #16a34a; font-size: 0.8rem; font-weight: 700;">-{j.poin} Poin</div>
                          </div>
                        {/each}
                      {/if}
                    </div>
                  {/if}
                {/if}
              </div>
              <div class="form-group" style="flex: 1; min-width: 120px;">
                <label for="iqob-tgl-modal">Tanggal</label>
                <input id="iqob-tgl-modal" type="date" class="input" bind:value={iqobFormTanggal} />
              </div>
              <div class="form-group" style="flex: 1; min-width: 100px;">
                <label for="iqob-poin-modal">Poin (-)</label>
                <input id="iqob-poin-modal" type="number" step="0.1" class="input" bind:value={iqobFormPoin} />
              </div>
            </div>

            <div style="display: flex; justify-content: flex-end; margin-top: 1rem;">
              <button class="btn-secondary" onclick={tambahPendingIqob} disabled={iqobFormLoading}>
                <Icon name="plus" size={14} /> Tambah ke Daftar
              </button>
            </div>
            {#if iqobFormError}
              <div class="alert-error mt-2">{iqobFormError}</div>
            {/if}

            {#if pendingIqobList.length > 0}
              <div style="margin-top: 1.5rem; background: #ffffff; border: 1px dashed #cbd5e1; border-radius: 8px; padding: 1rem;">
                <h5 style="margin: 0 0 0.75rem 0; font-size: 0.9rem; color: #475569;">Daftar Iqob Baru (Belum Disimpan)</h5>
                {#each pendingIqobList as p, i}
                  <div style="display: flex; justify-content: space-between; align-items: center; padding: 0.5rem; border-bottom: 1px solid #f1f5f9;">
                    <div>
                      <span style="font-weight: 500; font-size: 0.9rem;">{p.nama}</span>
                      <span style="color: #64748b; font-size: 0.8rem; margin-left: 8px;">{formatDate(p.tanggal)}</span>
                      <span style="color: #16a34a; font-size: 0.8rem; font-weight: 600; margin-left: 8px;">(-{p.poin})</span>
                    </div>
                    <button class="btn-icon-mini text-red" title="Hapus" onclick={() => removePendingIqob(i)}>
                      <Icon name="x" size={14} />
                    </button>
                  </div>
                {/each}
                <div style="display: flex; justify-content: flex-end; margin-top: 1rem;">
                  <button class="btn-primary" onclick={submitIqob} disabled={iqobFormLoading}>
                    {iqobFormLoading ? 'Menyimpan...' : `Simpan Semua (${pendingIqobList.length})`}
                  </button>
                </div>
              </div>
            {/if}
          </div>

          {@const iqobHistory = iqobCatatanMap[selectedStudent.siswa_id] || []}
          {#if iqobHistory.length > 0}
            <div class="history-list">
              {#each iqobHistory as c}
                <div class="history-item">
                  <div class="history-info">
                    <span class="history-date">{formatDate(c.tanggal)}</span>
                    <span class="history-desc">{c.nama_iqob} <span style="color: #16a34a; font-weight: 600;">(-{c.poin})</span></span>
                    {#if c.keterangan}
                      <span style="font-size: 0.85em; color: #64748b; font-style: italic;">Ket: {c.keterangan}</span>
                    {/if}
                  </div>
                  <button class="btn-icon-mini text-red" title="Hapus" onclick={() => hapusCatatanIqob(c.id)}>
                    <Icon name="x" size={14} />
                  </button>
                </div>
              {/each}
            </div>
          {:else}
            <div class="text-muted" style="text-align: center; padding: 2rem;">Tidak ada catatan iqob.</div>
          {/if}
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>

  .search-bar {
    display: inline-flex;
    align-items: center;
    background: #fff;
    border: 1px solid #cbd5e1;
    border-radius: 8px;
    padding: 0.4rem 0.75rem;
    gap: 0.5rem;
  }
  .search-icon {
    color: #94a3b8;
  }
  .search-input {
    border: none;
    background: transparent;
    font-size: 0.9rem;
    font-family: inherit;
    color: #1e293b;
    width: 200px;
    outline: none;
  }

  /* MODAL KUSTOM INTERNAL */
  .modal-backdrop {
    position: fixed;
    top: 0; left: 0; right: 0; bottom: 0;
    background: rgba(15, 23, 42, 0.6);
    backdrop-filter: blur(4px);
    z-index: 9999;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem;
  }
  .modal-content {
    background: white;
    width: 100%;
    max-width: 800px;
    max-height: 90vh;
    border-radius: 16px;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  .modal-header {
    padding: 1.5rem;
    border-bottom: 1px solid #e2e8f0;
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    background: #f8fafc;
  }
  .avatar-placeholder {
    width: 48px; height: 48px;
    border-radius: 50%;
    background: #e2e8f0;
    display: flex; align-items: center; justify-content: center;
  }
  .modal-body {
    padding: 1.5rem;
    overflow-y: auto;
    flex: 1;
  }
  
  .score-card {
    flex: 1;
    padding: 1rem;
    border-radius: 12px;
    border: 1px solid #e2e8f0;
    text-align: center;
  }
  .score-card.danger { background: #fef2f2; border-color: #fecaca; }
  .score-card.danger .score-label { color: #dc2626; }
  .score-card.success { background: #f0fdf4; border-color: #bbf7d0; }
  .score-card.success .score-label { color: #16a34a; }
  .score-card.primary { background: #eff6ff; border-color: #bfdbfe; }
  .score-card.primary .score-label { color: #2563eb; }
  
  .score-label { font-size: 0.85rem; font-weight: 600; margin-bottom: 0.25rem; }
  .score-value { font-size: 1.5rem; font-weight: 700; color: #1e293b; }

  .history-list {
    display: flex; flex-direction: column; gap: 0.75rem;
  }
  .history-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem 1rem;
    background: #ffffff;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    box-shadow: 0 1px 2px rgba(0,0,0,0.05);
  }
  .history-info {
    display: flex; align-items: center; gap: 1rem; flex-wrap: wrap;
  }
  .history-date {
    font-size: 0.85rem; color: #64748b; min-width: 80px;
  }
  .text-red { color: #ef4444 !important; }

  .page-container {
    padding: 1rem 0;
  }
  .header-action {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1.5rem;
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

  .tabs {
    display: flex;
    gap: 0;
    margin-bottom: 1.5rem;
    border-bottom: 2px solid #e2e8f0;
  }
  .tab-btn {
    padding: 0.75rem 1.5rem;
    font-size: 0.9rem;
    font-weight: 600;
    border: none;
    background: transparent;
    color: #64748b;
    cursor: pointer;
    border-bottom: 2px solid transparent;
    margin-bottom: -2px;
    transition: all 0.2s;
  }
  .tab-btn.active {
    color: #4f46e5;
    border-bottom-color: #4f46e5;
  }
  .tab-btn:hover:not(.active) {
    color: #334155;
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
  .text-sm {
    font-size: 0.8rem;
  }
  .text-gray-500 {
    color: #64748b;
  }
  .text-muted {
    color: #94a3b8;
    font-style: italic;
    font-size: 0.85rem;
  }

  .badge-gray {
    background: #f1f5f9;
    color: #94a3b8;
    padding: 0.2rem 0.5rem;
    border-radius: 4px;
    font-size: 0.8rem;
  }
  .badge-red {
    background: #fef2f2;
    color: #dc2626;
    padding: 0.2rem 0.6rem;
    border-radius: 4px;
    font-size: 0.85rem;
    font-weight: 700;
  }
  .badge-green {
    background: #f0fdf4;
    color: #16a34a;
    padding: 0.2rem 0.6rem;
    border-radius: 4px;
    font-size: 0.85rem;
    font-weight: 700;
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

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
    flex-wrap: wrap;
  }
  .materi-count {
    font-size: 0.9rem;
    color: #64748b;
  }
  .btn-primary {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.6rem 1.2rem;
    border: none;
    border-radius: 8px;
    background: #4f46e5;
    color: white;
    font-weight: 600;
    font-size: 0.875rem;
    cursor: pointer;
    transition: background 0.2s;
  }
  .btn-primary:hover:not(:disabled) {
    background: #4338ca;
  }
  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .btn-secondary {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.6rem 1.2rem;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    background: white;
    color: #475569;
    font-weight: 600;
    font-size: 0.875rem;
    cursor: pointer;
    transition: all 0.2s;
  }
  .btn-secondary:hover {
    background: #f8fafc;
  }

  .form-card {
    background: white;
    padding: 1.5rem;
    border-radius: 12px;
    border: 1px solid #e2e8f0;
    margin-bottom: 1.5rem;
  }
  .form-grid-plg {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
  }
  .form-grid-plg .form-group {
    max-width: none;
  }
  .form-actions {
    display: flex;
    gap: 0.75rem;
    justify-content: flex-end;
  }

  .alert-error {
    background: #fef2f2;
    color: #dc2626;
    padding: 0.75rem 1rem;
    border-radius: 8px;
    border: 1px solid #fecaca;
    font-size: 0.9rem;
  }

  .catatan-list {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }
  .catatan-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.8rem;
    background: #fef2f2;
    padding: 0.3rem 0.5rem;
    border-radius: 4px;
  }
  .catatan-date {
    color: #64748b;
    font-size: 0.75rem;
    min-width: 70px;
  }
  .catatan-poin {
    color: #dc2626;
    font-weight: 700;
    min-width: 30px;
  }
  .catatan-iqob {
    color: #475569;
    flex: 1;
  }

  .btn-icon-mini {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    height: 20px;
    border: none;
    background: transparent;
    color: #dc2626;
    cursor: pointer;
    border-radius: 4px;
    opacity: 0.6;
    transition: all 0.15s;
  }
  .btn-icon-mini:hover {
    opacity: 1;
    background: #fecaca;
  }

  .loading-state,
  .empty-state {
    padding: 3rem;
    text-align: center;
    color: #64748b;
    border: 1px dashed #cbd5e1;
    border-radius: 12px;
    background: #f8fafc;
  }
</style>
