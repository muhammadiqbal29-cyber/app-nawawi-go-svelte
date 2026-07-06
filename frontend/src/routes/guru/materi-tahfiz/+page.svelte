<script>
  import { toastStore } from '$lib/toast.svelte.js';
  import { modalStore } from '$lib/modal.svelte.js';
  import { SvelteMap } from 'svelte/reactivity';
  import { guruState } from '$lib/guruState.svelte.js';
  import Icon from '$lib/components/ui/Icon.svelte';

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

  // --- Derived State: Sekolah Terpilih ---
  let selectedSekolah = $derived(
    guruState.sekolahList.find((s) => s.id == guruState.activeSekolahId) || null
  );

  // --- Derived State: Fitur ---
  let selectedFeatures = $derived(
    (() => {
      if (!selectedSekolah || !selectedSekolah.fitur) return [];
      try {
        return JSON.parse(selectedSekolah.fitur);
      } catch {
        return [];
      }
    })()
  );
  let hasTahfiz = $derived(selectedFeatures.includes('input_tahfiz'));
  let hasDiniyah = $derived(selectedFeatures.includes('input_diniyah'));
  let hasKeseharian = $derived(selectedFeatures.includes('rapor_keseharian'));

  // --- State: Materi ---
  let materiList = $state([]);
  let loadingMateri = $state(false);
  let errorMateri = $state('');

  // --- State: Form tambah ---
  let showForm = $state(false);
  let formNama = $state('');
  let formJuz = $state('');
  let formUrutan = $state('');
  let formLoading = $state(false);
  let formError = $state('');

  // --- State: Edit inline ---
  let editingId = $state(null);
  let editNama = $state('');
  let editJuz = $state('');
  let editUrutan = $state('');
  let editLoading = $state(false);

  // --- State: Import Tahfiz ---
  let fileInputTahfiz = $state(null);
  let loadingImportTahfiz = $state(false);

  // --- State: Hapus ---
  let deletingId = $state(null);

  // --- State: Konfigurasi nilai kelulusan ---
  let loadingKonfigurasi = $state(false);
  let konfigurasiPembagi = $state('1');
  let konfigurasiError = $state('');
  let konfigurasiSuccess = $state('');
  let konfigurasiSaving = $state(false);
  let kelasKelulusanOptions = $state([]);
  let selectedTingkatKelulusan = $state([]);
  let kriteriaPenilaian = $state([]);

  // --- State: Tab ---
  let activeTab = $state('tahfiz');

  // --- State: Mapel Diniyah ---
  let diniyahList = $state([]);
  let loadingDiniyah = $state(false);
  let errorDiniyah = $state('');
  let showDiniyahForm = $state(false);
  let diniyahFormNama = $state('');
  let diniyahFormNamaArab = $state('');
  let diniyahFormKategori = $state('tes_tulis');
  let diniyahFormUrutan = $state('');
  let diniyahFormLoading = $state(false);
  let diniyahFormError = $state('');
  let diniyahEditingId = $state(null);
  let diniyahEditNama = $state('');
  let diniyahEditNamaArab = $state('');
  let diniyahEditKategori = $state('tes_tulis');
  let diniyahEditUrutan = $state('');
  let diniyahEditLoading = $state(false);
  let diniyahDeletingId = $state(null);

  // --- State: Keseharian ---
  let indikators = $state([]);
  let loadingKeseharian = $state(false);
  let errorKeseharian = $state('');

  // Import Keseharian
  let fileInputKeseharian = $state(null);
  let loadingImportKeseharian = $state(false);

  // Form Tambah Indikator
  let showKeseharianForm = $state(false);
  let keseharianFormNama = $state('');
  let keseharianFormNilaiPengurang = $state(1);
  let keseharianFormLoading = $state(false);
  let keseharianFormError = $state('');

  // Inline Edit Indikator
  let keseharianEditingId = $state(null);
  let keseharianEditNama = $state('');
  let keseharianEditNilaiPengurang = $state(1);
  let keseharianEditLoading = $state(false);
  let keseharianDeletingId = $state(null);

  $effect(() => {
    if (selectedSekolah) {
      loadMateri();
      loadKelasKelulusanOptions();
      loadKonfigurasiNilaiTahfiz();
      loadDiniyah();
      loadIndikators();
    }
  });

  async function loadMateri() {
    if (!selectedSekolah) return;
    loadingMateri = true;
    errorMateri = '';
    try {
      const res = await fetch(
        `${apiUrl}/api/guru/sekolah/${selectedSekolah.id}/materi-tahfiz/all`,
        {
          credentials: 'include'
        }
      );
      if (!res.ok) throw new Error('Gagal memuat materi');
      materiList = await res.json();
    } catch (err) {
      errorMateri = err.message;
    } finally {
      loadingMateri = false;
    }
  }

  async function loadKonfigurasiNilaiTahfiz() {
    if (!selectedSekolah) return;
    loadingKonfigurasi = true;
    konfigurasiError = '';
    konfigurasiSuccess = '';
    try {
      const res = await fetch(
        `${apiUrl}/api/guru/sekolah/${selectedSekolah.id}/konfigurasi-nilai-tahfiz`,
        {
          credentials: 'include'
        }
      );
      if (!res.ok) throw new Error('Gagal memuat pengaturan nilai kelulusan');
      const payload = await res.json();
      konfigurasiPembagi = String(payload.pembagi || 1);
      selectedTingkatKelulusan = payload.tingkat_level_kelulusan || [];
      kriteriaPenilaian = payload.kriteria_penilaian || [];
    } catch (err) {
      konfigurasiError = err.message;
    } finally {
      loadingKonfigurasi = false;
    }
  }

  async function loadKelasKelulusanOptions() {
    if (!selectedSekolah) return;
    try {
      const res = await fetch(`${apiUrl}/api/guru/sekolah/${selectedSekolah.id}/kelas`, {
        credentials: 'include'
      });
      if (!res.ok) throw new Error('Gagal memuat daftar kelas');
      const kelasData = await res.json();
      const levelMap = new SvelteMap();
      for (const kelas of kelasData || []) {
        const level = Number(kelas.tingkat_level || 0);
        if (level <= 0 || levelMap.has(level)) continue;
        levelMap.set(level, {
          level,
          label: kelas.tingkat ? `Kelas ${kelas.tingkat}` : `Tingkat ${level}`
        });
      }
      kelasKelulusanOptions = Array.from(levelMap.values()).sort((a, b) => a.level - b.level);
    } catch (err) {
      konfigurasiError = err.message;
    }
  }

  async function simpanKonfigurasiNilaiTahfiz() {
    const pembagi = Number(konfigurasiPembagi);
    if (!Number.isInteger(pembagi) || pembagi < 1) {
      konfigurasiError = 'Pembagi harus berupa angka positif minimal 1.';
      konfigurasiSuccess = '';
      return;
    }
    if (selectedTingkatKelulusan.length === 0) {
      konfigurasiError = 'Pilih minimal satu tingkat kelas untuk ujian kelulusan.';
      konfigurasiSuccess = '';
      return;
    }

    konfigurasiSaving = true;
    konfigurasiError = '';
    konfigurasiSuccess = '';
    try {
      const res = await fetch(
        `${apiUrl}/api/guru/sekolah/${selectedSekolah.id}/konfigurasi-nilai-tahfiz`,
        {
          method: 'POST',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            jenis_penilaian: 'kelulusan',
            pembagi,
            tingkat_level_kelulusan: selectedTingkatKelulusan,
            kriteria_penilaian: kriteriaPenilaian
          })
        }
      );
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.message || 'Gagal menyimpan pengaturan nilai kelulusan');
      }
      const payload = await res.json();
      konfigurasiPembagi = String(payload.pembagi || pembagi);
      selectedTingkatKelulusan = payload.tingkat_level_kelulusan || selectedTingkatKelulusan;
      kriteriaPenilaian = payload.kriteria_penilaian || kriteriaPenilaian;
      konfigurasiSuccess = 'Pengaturan nilai ujian kelulusan berhasil disimpan.';
    } catch (err) {
      konfigurasiError = err.message;
    } finally {
      konfigurasiSaving = false;
    }
  }

  async function tambahMateri() {
    if (!formNama.trim()) {
      formError = 'Nama materi wajib diisi.';
      return;
    }
    formLoading = true;
    formError = '';
    try {
      const res = await fetch(`${apiUrl}/api/guru/sekolah/${selectedSekolah.id}/materi-tahfiz`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          nama_materi: formNama.trim(),
          juz: formJuz ? Number(formJuz) : 0,
          urutan: formUrutan ? Number(formUrutan) : 0
        })
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.message || 'Gagal menambahkan materi');
      }
      formNama = '';
      formJuz = '';
      formUrutan = '';
      showForm = false;
      await loadMateri();
    } catch (err) {
      formError = err.message;
    } finally {
      formLoading = false;
    }
  }

  function handleDownloadTemplateTahfiz() {
    if (!selectedSekolah) return;
    window.location.href = `${apiUrl}/api/guru/sekolah/${selectedSekolah.id}/materi-tahfiz-template`;
  }

  async function handleImportTahfiz(event) {
    if (!selectedSekolah) return;
    const file = event.target.files[0];
    if (!file) return;

    loadingImportTahfiz = true;
    errorMateri = '';
    const formData = new FormData();
    formData.append('file', file);

    try {
      const res = await fetch(
        `${apiUrl}/api/guru/sekolah/${selectedSekolah.id}/materi-tahfiz-import`,
        {
          method: 'POST',
          credentials: 'include',
          body: formData
        }
      );
      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.message || 'Gagal melakukan import');
      }
      toastStore.show(data.message || 'Import berhasil', 'info');
      await loadMateri();
    } catch (err) {
      errorMateri = err.message;
    } finally {
      loadingImportTahfiz = false;
      if (fileInputTahfiz) {
        fileInputTahfiz.value = '';
      }
    }
  }

  function startEdit(materi) {
    editingId = materi.id;
    editNama = materi.nama_materi;
    editJuz = materi.juz > 0 ? String(materi.juz) : '';
    editUrutan = materi.urutan > 0 ? String(materi.urutan) : '';
  }

  function cancelEdit() {
    editingId = null;
    editNama = '';
    editJuz = '';
    editUrutan = '';
  }

  async function simpanEdit(materi) {
    if (!editNama.trim()) return;
    editLoading = true;
    try {
      const res = await fetch(`${apiUrl}/api/guru/materi-tahfiz/${materi.id}`, {
        method: 'PUT',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          nama_materi: editNama.trim(),
          juz: editJuz ? Number(editJuz) : 0,
          urutan: editUrutan ? Number(editUrutan) : 0
        })
      });
      if (!res.ok) throw new Error('Gagal menyimpan');
      editingId = null;
      await loadMateri();
    } catch (err) {
      errorMateri = err.message;
    } finally {
      editLoading = false;
    }
  }

  async function toggleAktif(materi) {
    try {
      const aktifBaru = !materi.aktif;
      const res = await fetch(`${apiUrl}/api/guru/materi-tahfiz/${materi.id}`, {
        method: 'PUT',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ aktif: aktifBaru })
      });
      if (!res.ok) throw new Error('Gagal mengubah status');
      await loadMateri();
    } catch (err) {
      errorMateri = err.message;
    }
  }

  async function hapusMateri(id) {
    if (!(await modalStore.show({ message: 'Yakin ingin menghapus materi ini?' }))) return;
    deletingId = id;
    try {
      const res = await fetch(`${apiUrl}/api/guru/materi-tahfiz/${id}`, {
        method: 'DELETE',
        credentials: 'include'
      });
      if (!res.ok) throw new Error('Gagal menghapus');
      await loadMateri();
    } catch (err) {
      errorMateri = err.message;
    } finally {
      deletingId = null;
    }
  }

  function batalForm() {
    showForm = false;
    formNama = '';
    formJuz = '';
    formUrutan = '';
    formError = '';
  }

  function toggleTingkatKelulusan(level) {
    if (selectedTingkatKelulusan.includes(level)) {
      selectedTingkatKelulusan = selectedTingkatKelulusan.filter((item) => item !== level);
    } else {
      selectedTingkatKelulusan = [...selectedTingkatKelulusan, level].sort((a, b) => a - b);
    }
  }

  function tambahKriteria() {
    kriteriaPenilaian = [...kriteriaPenilaian, { min: 0, max: 0, arab: '', indo: '' }];
  }

  function hapusKriteria(index) {
    kriteriaPenilaian = kriteriaPenilaian.filter((_, i) => i !== index);
  }

  // ===== DINIYAH FUNCTIONS =====
  async function loadDiniyah() {
    if (!selectedSekolah) return;
    loadingDiniyah = true;
    errorDiniyah = '';
    try {
      const res = await fetch(
        `${apiUrl}/api/guru/sekolah/${selectedSekolah.id}/mapel-diniyah/all`,
        {
          credentials: 'include'
        }
      );
      if (!res.ok) throw new Error('Gagal memuat mapel diniyah');
      diniyahList = await res.json();
    } catch (err) {
      errorDiniyah = err.message;
    } finally {
      loadingDiniyah = false;
    }
  }

  async function tambahDiniyah() {
    if (!diniyahFormNama.trim()) {
      diniyahFormError = 'Nama mapel wajib diisi.';
      return;
    }
    diniyahFormLoading = true;
    diniyahFormError = '';
    try {
      const res = await fetch(`${apiUrl}/api/guru/sekolah/${selectedSekolah.id}/mapel-diniyah`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          nama_mapel: diniyahFormNama.trim(),
          nama_mapel_arab: diniyahFormNamaArab.trim(),
          kategori: diniyahFormKategori,
          urutan: diniyahFormUrutan ? Number(diniyahFormUrutan) : 0
        })
      });
      if (!res.ok) {
        const b = await res.json().catch(() => ({}));
        throw new Error(b.message || 'Gagal menambahkan mapel');
      }
      diniyahFormNama = '';
      diniyahFormNamaArab = '';
      diniyahFormKategori = 'tes_tulis';
      diniyahFormUrutan = '';
      showDiniyahForm = false;
      await loadDiniyah();
    } catch (err) {
      diniyahFormError = err.message;
    } finally {
      diniyahFormLoading = false;
    }
  }

  function startDiniyahEdit(m) {
    diniyahEditingId = m.id;
    diniyahEditNama = m.nama_mapel;
    diniyahEditNamaArab = m.nama_mapel_arab || '';
    diniyahEditKategori = m.kategori || 'tes_tulis';
    diniyahEditUrutan = m.urutan > 0 ? String(m.urutan) : '';
  }

  function cancelDiniyahEdit() {
    diniyahEditingId = null;
  }

  async function simpanDiniyahEdit(m) {
    if (!diniyahEditNama.trim()) return;
    diniyahEditLoading = true;
    try {
      const res = await fetch(`${apiUrl}/api/guru/mapel-diniyah/${m.id}`, {
        method: 'PUT',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          nama_mapel: diniyahEditNama.trim(),
          nama_mapel_arab: diniyahEditNamaArab.trim(),
          kategori: diniyahEditKategori,
          urutan: diniyahEditUrutan ? Number(diniyahEditUrutan) : 0
        })
      });
      if (!res.ok) throw new Error('Gagal menyimpan');
      diniyahEditingId = null;
      await loadDiniyah();
    } catch (err) {
      errorDiniyah = err.message;
    } finally {
      diniyahEditLoading = false;
    }
  }

  async function toggleDiniyahAktif(m) {
    try {
      const res = await fetch(`${apiUrl}/api/guru/mapel-diniyah/${m.id}`, {
        method: 'PUT',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ aktif: !m.aktif })
      });
      if (!res.ok) throw new Error('Gagal mengubah status');
      await loadDiniyah();
    } catch (err) {
      errorDiniyah = err.message;
    }
  }

  async function hapusDiniyah(id) {
    if (!(await modalStore.show({ message: 'Yakin ingin menghapus mapel ini?' }))) return;
    diniyahDeletingId = id;
    try {
      const res = await fetch(`${apiUrl}/api/guru/mapel-diniyah/${id}`, {
        method: 'DELETE',
        credentials: 'include'
      });
      if (!res.ok) throw new Error('Gagal menghapus');
      await loadDiniyah();
    } catch (err) {
      errorDiniyah = err.message;
    } finally {
      diniyahDeletingId = null;
    }
  }

  function batalDiniyahForm() {
    showDiniyahForm = false;
    diniyahFormNama = '';
    diniyahFormNamaArab = '';
    diniyahFormKategori = 'tes_tulis';
    diniyahFormUrutan = '';
    diniyahFormError = '';
  }

  // ==== INDIKATOR KESEHARIAN ACTIONS ====
  async function loadIndikators() {
    if (!selectedSekolah) return;
    loadingKeseharian = true;
    errorKeseharian = '';
    try {
      const res = await fetch(
        `${apiUrl}/api/guru/sekolah/${selectedSekolah.id}/keseharian/indikator`,
        {
          credentials: 'include'
        }
      );
      if (!res.ok) throw new Error('Gagal memuat indikator keseharian.');

      const body = await res.json();
      indikators = body.data || [];
    } catch (err) {
      errorKeseharian = err.message;
    } finally {
      loadingKeseharian = false;
    }
  }

  function handleDownloadTemplateKeseharian() {
    if (!selectedSekolah) return;
    window.location.href = `${apiUrl}/api/guru/sekolah/${selectedSekolah.id}/keseharian/indikator/template`;
  }

  async function handleImportKeseharian(event) {
    if (!selectedSekolah) return;
    const file = event.target.files[0];
    if (!file) return;

    loadingImportKeseharian = true;
    errorKeseharian = '';
    const formData = new FormData();
    formData.append('file', file);

    try {
      const res = await fetch(
        `${apiUrl}/api/guru/sekolah/${selectedSekolah.id}/keseharian/indikator/import`,
        {
          method: 'POST',
          credentials: 'include',
          body: formData
        }
      );
      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.message || 'Gagal melakukan import indikator');
      }
      toastStore.show(data.message || 'Import indikator keseharian berhasil', 'info');
      await loadIndikators();
    } catch (err) {
      errorKeseharian = err.message;
      toastStore.show(errorKeseharian, 'info');
    } finally {
      loadingImportKeseharian = false;
      if (fileInputKeseharian) {
        fileInputKeseharian.value = '';
      }
    }
  }

  function startKeseharianEdit(ind) {
    keseharianEditingId = ind.id;
    keseharianEditNama = ind.nama;
    keseharianEditNilaiPengurang = ind.nilai_pengurang || 1;
  }

  function cancelKeseharianEdit() {
    keseharianEditingId = null;
    keseharianEditNama = '';
    keseharianEditNilaiPengurang = 1;
  }

  async function simpanKeseharianEdit(ind) {
    if (!keseharianEditNama.trim()) return;
    keseharianEditLoading = true;
    try {
      const res = await fetch(`${apiUrl}/api/guru/keseharian/indikator/${ind.id}`, {
        method: 'PUT',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          nama: keseharianEditNama,
          nilai_pengurang: keseharianEditNilaiPengurang,
          aktif: ind.aktif
        })
      });
      if (!res.ok) throw new Error('Gagal menyimpan perubahan');
      await loadIndikators();
      cancelKeseharianEdit();
    } catch (err) {
      toastStore.show(err.message, 'info');
    } finally {
      keseharianEditLoading = false;
    }
  }

  async function toggleKeseharianAktif(ind) {
    try {
      const res = await fetch(`${apiUrl}/api/guru/keseharian/indikator/${ind.id}`, {
        method: 'PUT',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ nama: ind.nama, aktif: !ind.aktif })
      });
      if (!res.ok) throw new Error('Gagal mengubah status');
      await loadIndikators();
    } catch (err) {
      toastStore.show(err.message, 'info');
    }
  }

  async function hapusKeseharian(id) {
    if (!(await modalStore.show({ message: 'Hapus indikator ini?' }))) return;
    keseharianDeletingId = id;
    try {
      const res = await fetch(`${apiUrl}/api/guru/keseharian/indikator/${id}`, {
        method: 'DELETE',
        credentials: 'include'
      });
      if (!res.ok) throw new Error('Gagal menghapus');
      await loadIndikators();
    } catch (err) {
      toastStore.show(err.message, 'info');
    } finally {
      keseharianDeletingId = null;
    }
  }

  function batalKeseharianForm() {
    showKeseharianForm = false;
    keseharianFormNama = '';
    keseharianFormNilaiPengurang = 1;
    keseharianFormError = '';
  }

  async function tambahKeseharian() {
    if (!keseharianFormNama.trim()) {
      keseharianFormError = 'Nama indikator harus diisi';
      return;
    }
    keseharianFormLoading = true;
    keseharianFormError = '';
    try {
      const res = await fetch(
        `${apiUrl}/api/guru/sekolah/${selectedSekolah.id}/keseharian/indikator`,
        {
          method: 'POST',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            nama: keseharianFormNama,
            nilai_pengurang: keseharianFormNilaiPengurang,
            aktif: true
          })
        }
      );
      if (!res.ok) throw new Error('Gagal menyimpan indikator');
      await loadIndikators();
      batalKeseharianForm();
    } catch (err) {
      keseharianFormError = err.message;
    } finally {
      keseharianFormLoading = false;
    }
  }
</script>

<svelte:head>
  <title>Konfigurasi Materi - SIAKAD</title>
</svelte:head>

<div class="page-container">
  <div class="page-header">
    <div>
      <h1 class="page-title">Konfigurasi Materi</h1>
      <p class="page-subtitle">Kelola daftar materi tahfiz dan mapel diniyah per sekolah.</p>
    </div>
  </div>

  <!-- Pilih Sekolah via Layout -->
  {#if !selectedSekolah}
    <div class="empty-state">
      <Icon name="grid" size={48} />
      <h3>Pilih Jenjang Sekolah</h3>
      <p>Silakan pilih jenjang sekolah pada menu di pojok kanan atas.</p>
    </div>
  {:else}
    <!-- Tabs -->
    <div class="tabs">
      {#if hasTahfiz}
        <button
          class="tab-btn"
          class:active={activeTab === 'tahfiz'}
          onclick={() => (activeTab = 'tahfiz')}
        >
          Materi Tahfiz
        </button>
      {/if}
      {#if hasDiniyah}
        <button
          class="tab-btn"
          class:active={activeTab === 'diniyah'}
          onclick={() => (activeTab = 'diniyah')}
        >
          Mapel Diniyah
        </button>
      {/if}
      {#if hasKeseharian}
        <button
          class="tab-btn"
          class:active={activeTab === 'keseharian'}
          onclick={() => (activeTab = 'keseharian')}
        >
          Rapor Keseharian
        </button>
      {/if}
    </div>

    {#if !hasTahfiz && !hasDiniyah && !hasKeseharian}
      <div class="empty-state">
        <Icon name="slash" size={48} />
        <h3>Tidak Ada Program Khusus</h3>
        <p>
          Sekolah ini tidak memiliki pengaturan program Tahfiz, Diniyah, ataupun Keseharian. Silakan
          aktifkan melalui Pengaturan Institusi jika diperlukan.
        </p>
      </div>
    {:else}
      {#if activeTab === 'tahfiz' && hasTahfiz}
        <!-- ======= TAB TAHFIZ (konten lama) ======= -->

        <!-- Toolbar -->
        <div class="toolbar">
          <span class="materi-count">
            {materiList.length} materi terdaftar ({materiList.filter((m) => m.aktif).length} aktif)
          </span>
          <div class="toolbar-actions" style="display: flex; gap: 8px;">
            <button class="btn-secondary" onclick={handleDownloadTemplateTahfiz}>
              <Icon name="download" size={16} />
              Template
            </button>
            <button
              class="btn-secondary"
              onclick={() => fileInputTahfiz.click()}
              disabled={loadingImportTahfiz}
            >
              <Icon name="import" size={16} />
              {loadingImportTahfiz ? 'Mengimport...' : 'Import'}
            </button>
            <input
              type="file"
              accept=".xlsx"
              bind:this={fileInputTahfiz}
              onchange={handleImportTahfiz}
              style="display: none;"
            />
            <button class="btn-primary" onclick={() => (showForm = !showForm)}>
              <Icon name="plus" size={16} />
              Tambah Materi
            </button>
          </div>
        </div>

        <!-- Form tambah -->
        {#if showForm}
          <div class="form-card">
            <h3 class="form-title">Tambah Materi Baru</h3>
            {#if formError}
              <div class="alert-error">{formError}</div>
            {/if}
            <div class="form-row">
              <div class="form-group form-group-wide">
                <label for="form-nama">Nama Materi <span class="required">*</span></label>
                <input
                  id="form-nama"
                  type="text"
                  placeholder="contoh: An-Naba, Halaman 201-205"
                  bind:value={formNama}
                  class="input"
                />
              </div>
              <div class="form-group">
                <label for="form-juz">Juz</label>
                <input
                  id="form-juz"
                  type="number"
                  min="0"
                  max="30"
                  placeholder="0"
                  bind:value={formJuz}
                  class="input"
                />
              </div>
              <div class="form-group">
                <label for="form-urutan">Urutan</label>
                <input
                  id="form-urutan"
                  type="number"
                  min="1"
                  placeholder="Otomatis"
                  bind:value={formUrutan}
                  class="input"
                />
              </div>
            </div>
            <div class="form-actions">
              <button class="btn-secondary" onclick={batalForm} disabled={formLoading}>Batal</button
              >
              <button class="btn-primary" onclick={tambahMateri} disabled={formLoading}>
                {formLoading ? 'Menyimpan...' : 'Simpan Materi'}
              </button>
            </div>
          </div>
        {/if}

        {#if errorMateri}
          <div class="alert-error">{errorMateri}</div>
        {/if}

        <!-- Tabel materi -->
        {#if loadingMateri}
          <div class="card-loading">Memuat daftar materi...</div>
        {:else if materiList.length === 0}
          <div class="empty-state">
            <Icon name="list" size={48} />
            <h3>Belum ada materi</h3>
            <p>Tambahkan materi tahfiz agar guru dapat mulai menginput nilai.</p>
            <button class="btn-primary" onclick={() => (showForm = true)}>
              <Icon name="plus" size={16} /> Tambah Materi Pertama
            </button>
          </div>
        {:else}
          <div class="table-wrapper">
            <table class="materi-table">
              <thead>
                <tr>
                  <th class="col-urutan">Urutan</th>
                  <th class="col-nama">Nama Materi</th>
                  <th class="col-juz">Juz</th>
                  <th class="col-status">Status</th>
                  <th class="col-aksi">Aksi</th>
                </tr>
              </thead>
              <tbody>
                {#each materiList as materi (materi.id)}
                  <tr class:row-nonaktif={!materi.aktif}>
                    {#if editingId === materi.id}
                      <!-- Edit mode -->
                      <td>
                        <input
                          type="number"
                          min="1"
                          bind:value={editUrutan}
                          class="input-inline"
                          placeholder="—"
                        />
                      </td>
                      <td>
                        <input
                          type="text"
                          bind:value={editNama}
                          class="input-inline input-wide"
                          placeholder="Nama materi"
                        />
                      </td>
                      <td>
                        <input
                          type="number"
                          min="0"
                          max="30"
                          bind:value={editJuz}
                          class="input-inline"
                          placeholder="0"
                        />
                      </td>
                      <td>—</td>
                      <td class="aksi-cell">
                        <button
                          class="btn-save"
                          onclick={() => simpanEdit(materi)}
                          disabled={editLoading}
                        >
                          {editLoading ? '...' : 'Simpan'}
                        </button>
                        <button class="btn-cancel" onclick={cancelEdit}>Batal</button>
                      </td>
                    {:else}
                      <!-- Tampil normal -->
                      <td class="text-center text-muted">{materi.urutan || '—'}</td>
                      <td class="nama-cell">
                        <span class:text-muted={!materi.aktif}>{materi.nama_materi}</span>
                      </td>
                      <td class="text-center">
                        {materi.juz > 0 ? `Juz ${materi.juz}` : '—'}
                      </td>
                      <td class="text-center">
                        <button
                          class="toggle-btn"
                          class:aktif={materi.aktif}
                          class:nonaktif={!materi.aktif}
                          onclick={() => toggleAktif(materi)}
                          title={materi.aktif ? 'Klik untuk nonaktifkan' : 'Klik untuk aktifkan'}
                        >
                          {materi.aktif ? 'Aktif' : 'Nonaktif'}
                        </button>
                      </td>
                      <td class="aksi-cell">
                        <button class="btn-icon" title="Edit" onclick={() => startEdit(materi)}>
                          <Icon name="edit2" size={15} />
                        </button>
                        <button
                          class="btn-icon btn-icon-danger"
                          title="Hapus"
                          onclick={() => hapusMateri(materi.id)}
                          disabled={deletingId === materi.id}
                        >
                          <Icon name="trash" size={15} />
                        </button>
                      </td>
                    {/if}
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}

        <section class="kelulusan-settings">
          <div class="settings-header">
            <div>
              <h2 class="settings-title">Pengaturan Nilai Ujian Kelulusan</h2>
              <p class="settings-subtitle">Rumus: (Total Poin x 4) / Pembagi</p>
            </div>
          </div>

          {#if konfigurasiError}
            <div class="alert-error">{konfigurasiError}</div>
          {/if}
          {#if konfigurasiSuccess}
            <div class="alert-success">{konfigurasiSuccess}</div>
          {/if}

          <div class="settings-row">
            <div class="form-group settings-input">
              <label for="konfigurasi-pembagi">Pembagi nilai ujian kelulusan</label>
              <input
                id="konfigurasi-pembagi"
                type="number"
                min="1"
                step="1"
                bind:value={konfigurasiPembagi}
                class="input"
                disabled={loadingKonfigurasi || konfigurasiSaving}
              />
            </div>
            <button
              class="btn-primary"
              onclick={simpanKonfigurasiNilaiTahfiz}
              disabled={loadingKonfigurasi || konfigurasiSaving}
            >
              {konfigurasiSaving ? 'Menyimpan...' : 'Simpan'}
            </button>
          </div>

          <div class="settings-class-block">
            <div class="settings-class-label">Tingkat kelas yang mengikuti ujian kelulusan</div>
            {#if kelasKelulusanOptions.length === 0}
              <div class="settings-note">Belum ada kelas yang bisa dipilih untuk sekolah ini.</div>
            {:else}
              <div class="level-options">
                {#each kelasKelulusanOptions as option (option.level)}
                  <label class="level-option">
                    <input
                      type="checkbox"
                      checked={selectedTingkatKelulusan.includes(option.level)}
                      onchange={() => toggleTingkatKelulusan(option.level)}
                      disabled={loadingKonfigurasi || konfigurasiSaving}
                    />
                    <span>{option.label}</span>
                  </label>
                {/each}
              </div>
            {/if}
          </div>

          <div class="settings-kriteria-block">
            <div class="settings-kriteria-header">
              <div class="settings-class-label">Kriteria Penilaian Rapor (Skala 100)</div>
              <button
                class="btn-secondary btn-sm"
                onclick={tambahKriteria}
                disabled={loadingKonfigurasi || konfigurasiSaving}
              >
                <Icon name="plus" size={14} /> Tambah
              </button>
            </div>

            <div class="kriteria-list">
              {#each kriteriaPenilaian as kriteria, index (kriteria.id || index)}
                <div class="kriteria-row">
                  <div class="form-group">
                    <label>Batas Bawah</label>
                    <input
                      type="number"
                      class="input"
                      bind:value={kriteria.min}
                      min="0"
                      max="100"
                    />
                  </div>
                  <div class="form-group">
                    <label>Batas Atas</label>
                    <input
                      type="number"
                      class="input"
                      bind:value={kriteria.max}
                      min="0"
                      max="100"
                    />
                  </div>
                  <div class="form-group kriteria-arab">
                    <label>Predikat Arab</label>
                    <input
                      type="text"
                      class="input"
                      bind:value={kriteria.arab}
                      placeholder="مُمْتَاز"
                      dir="rtl"
                    />
                  </div>
                  <div class="form-group kriteria-indo">
                    <label>Predikat Indonesia</label>
                    <input
                      type="text"
                      class="input"
                      bind:value={kriteria.indo}
                      placeholder="Istimewa"
                    />
                  </div>
                  <button
                    class="btn-icon btn-icon-danger kriteria-hapus"
                    onclick={() => hapusKriteria(index)}
                    title="Hapus Kriteria"
                  >
                    <Icon name="trash" size={16} />
                  </button>
                </div>
              {/each}
              {#if kriteriaPenilaian.length === 0}
                <div class="settings-note">
                  Belum ada kriteria penilaian. Klik Tambah untuk membuat.
                </div>
              {/if}
            </div>
          </div>

          <p class="settings-note">
            Hanya berlaku untuk nilai yang belum tersimpan. Nilai yang sudah ada tidak akan berubah.
          </p>
        </section>
        <!-- ======= END TAB TAHFIZ ======= -->
      {:else if activeTab === 'diniyah' && hasDiniyah}
        <!-- ======= TAB DINIYAH ======= -->
        <div class="toolbar">
          <span class="materi-count">
            {diniyahList.length} mapel terdaftar ({diniyahList.filter((m) => m.aktif).length} aktif)
          </span>
          <button class="btn-primary" onclick={() => (showDiniyahForm = !showDiniyahForm)}>
            <Icon name="plus" size={16} />
            Tambah Mapel
          </button>
        </div>

        {#if showDiniyahForm}
          <div class="form-card">
            <h3 class="form-title">Tambah Mapel Diniyah Baru</h3>
            {#if diniyahFormError}
              <div class="alert-error">{diniyahFormError}</div>
            {/if}
            <div class="form-row">
              <div class="form-group form-group-wide">
                <label for="diniyah-nama">Nama Mapel (Latin) <span class="required">*</span></label>
                <input
                  id="diniyah-nama"
                  type="text"
                  placeholder="contoh: KHUDZ AQIDATAKA"
                  bind:value={diniyahFormNama}
                  class="input"
                />
              </div>
              <div class="form-group form-group-wide">
                <label for="diniyah-arab">Nama Mapel (Arab)</label>
                <input
                  id="diniyah-arab"
                  type="text"
                  placeholder="contoh: خذ عقيدتك"
                  bind:value={diniyahFormNamaArab}
                  class="input input-arab"
                  dir="rtl"
                />
              </div>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label for="diniyah-kategori">Kategori</label>
                <select id="diniyah-kategori" bind:value={diniyahFormKategori} class="input">
                  <option value="tes_tulis">Tes Tulis</option>
                  <option value="tes_lisan">Tes Lisan</option>
                </select>
              </div>
              <div class="form-group">
                <label for="diniyah-urutan">Urutan</label>
                <input
                  id="diniyah-urutan"
                  type="number"
                  min="1"
                  placeholder="Otomatis"
                  bind:value={diniyahFormUrutan}
                  class="input"
                />
              </div>
            </div>
            <div class="form-actions">
              <button class="btn-secondary" onclick={batalDiniyahForm} disabled={diniyahFormLoading}
                >Batal</button
              >
              <button class="btn-primary" onclick={tambahDiniyah} disabled={diniyahFormLoading}>
                {diniyahFormLoading ? 'Menyimpan...' : 'Simpan Mapel'}
              </button>
            </div>
          </div>
        {/if}

        {#if errorDiniyah}
          <div class="alert-error">{errorDiniyah}</div>
        {/if}

        {#if loadingDiniyah}
          <div class="card-loading">Memuat daftar mapel diniyah...</div>
        {:else if diniyahList.length === 0}
          <div class="empty-state">
            <Icon name="book" size={48} />
            <h3>Belum ada mapel diniyah</h3>
            <p>Tambahkan mapel diniyah agar guru dapat mulai menginput nilai.</p>
            <button class="btn-primary" onclick={() => (showDiniyahForm = true)}>
              <Icon name="plus" size={16} /> Tambah Mapel Pertama
            </button>
          </div>
        {:else}
          <div class="table-wrapper">
            <table class="materi-table">
              <thead>
                <tr>
                  <th class="col-urutan">Urutan</th>
                  <th class="col-nama">Nama Mapel</th>
                  <th class="col-nama">Nama Arab</th>
                  <th class="col-juz">Kategori</th>
                  <th class="col-status">Status</th>
                  <th class="col-aksi">Aksi</th>
                </tr>
              </thead>
              <tbody>
                {#each diniyahList as mapel (mapel.id)}
                  <tr class:row-nonaktif={!mapel.aktif}>
                    {#if diniyahEditingId === mapel.id}
                      <td
                        ><input
                          type="number"
                          min="1"
                          bind:value={diniyahEditUrutan}
                          class="input-inline"
                          placeholder="—"
                        /></td
                      >
                      <td
                        ><input
                          type="text"
                          bind:value={diniyahEditNama}
                          class="input-inline input-wide"
                          placeholder="Nama mapel"
                        /></td
                      >
                      <td
                        ><input
                          type="text"
                          bind:value={diniyahEditNamaArab}
                          class="input-inline input-wide input-arab"
                          dir="rtl"
                          placeholder="Nama arab"
                        /></td
                      >
                      <td>
                        <select bind:value={diniyahEditKategori} class="input-inline">
                          <option value="tes_tulis">Tes Tulis</option>
                          <option value="tes_lisan">Tes Lisan</option>
                        </select>
                      </td>
                      <td>—</td>
                      <td class="aksi-cell">
                        <button
                          class="btn-save"
                          onclick={() => simpanDiniyahEdit(mapel)}
                          disabled={diniyahEditLoading}
                          >{diniyahEditLoading ? '...' : 'Simpan'}</button
                        >
                        <button class="btn-cancel" onclick={cancelDiniyahEdit}>Batal</button>
                      </td>
                    {:else}
                      <td class="text-center text-muted">{mapel.urutan || '—'}</td>
                      <td class="nama-cell"
                        ><span class:text-muted={!mapel.aktif}>{mapel.nama_mapel}</span></td
                      >
                      <td class="nama-cell arab-cell" dir="rtl"
                        ><span class:text-muted={!mapel.aktif}>{mapel.nama_mapel_arab || '—'}</span
                        ></td
                      >
                      <td class="text-center"
                        >{mapel.kategori === 'tes_lisan' ? 'Tes Lisan' : 'Tes Tulis'}</td
                      >
                      <td class="text-center">
                        <button
                          class="toggle-btn"
                          class:aktif={mapel.aktif}
                          class:nonaktif={!mapel.aktif}
                          onclick={() => toggleDiniyahAktif(mapel)}
                          title={mapel.aktif ? 'Klik untuk nonaktifkan' : 'Klik untuk aktifkan'}
                        >
                          {mapel.aktif ? 'Aktif' : 'Nonaktif'}
                        </button>
                      </td>
                      <td class="aksi-cell">
                        <button
                          class="btn-icon"
                          title="Edit"
                          onclick={() => startDiniyahEdit(mapel)}
                          ><Icon name="edit2" size={15} /></button
                        >
                        <button
                          class="btn-icon btn-icon-danger"
                          title="Hapus"
                          onclick={() => hapusDiniyah(mapel.id)}
                          disabled={diniyahDeletingId === mapel.id}
                          ><Icon name="trash" size={15} /></button
                        >
                      </td>
                    {/if}
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
        <!-- ======= END TAB DINIYAH ======= -->
      {:else if activeTab === 'keseharian' && hasKeseharian}
        <!-- ======= TAB KESEHARIAN ======= -->
        <div class="toolbar">
          <span class="materi-count">
            {indikators.length} indikator ({indikators.filter((m) => m.aktif).length} aktif)
          </span>
          <div class="toolbar-actions" style="display: flex; gap: 8px;">
            <button class="btn-secondary" onclick={handleDownloadTemplateKeseharian}>
              <Icon name="download" size={16} />
              Template
            </button>
            <button
              class="btn-secondary"
              onclick={() => fileInputKeseharian.click()}
              disabled={loadingImportKeseharian}
            >
              <Icon name="import" size={16} />
              {loadingImportKeseharian ? 'Mengimport...' : 'Import'}
            </button>
            <input
              type="file"
              accept=".xlsx"
              bind:this={fileInputKeseharian}
              onchange={handleImportKeseharian}
              style="display: none;"
            />
            <button class="btn-primary" onclick={() => (showKeseharianForm = !showKeseharianForm)}>
              <Icon name="plus" size={16} /> Tambah Indikator
            </button>
          </div>
        </div>

        {#if showKeseharianForm}
          <div class="materi-form">
            <div class="form-grid" style="grid-template-columns: 2fr 1fr;">
              <div class="form-group">
                <label for="kes-nama">Nama Indikator Keseharian</label>
                <input
                  id="kes-nama"
                  type="text"
                  placeholder="Contoh: Shalat Dhuha, Piket Asrama, dsb"
                  bind:value={keseharianFormNama}
                  class="input"
                />
              </div>
              <div class="form-group">
                <label for="kes-minus">Poin Pengurang (Pelanggaran)</label>
                <input
                  id="kes-minus"
                  type="number"
                  min="0"
                  step="1"
                  bind:value={keseharianFormNilaiPengurang}
                  class="input"
                />
              </div>
            </div>
            <div class="form-actions">
              <button
                class="btn-secondary"
                onclick={batalKeseharianForm}
                disabled={keseharianFormLoading}>Batal</button
              >
              <button
                class="btn-primary"
                onclick={tambahKeseharian}
                disabled={keseharianFormLoading || !keseharianFormNama.trim()}
              >
                {keseharianFormLoading ? 'Menyimpan...' : 'Simpan Indikator'}
              </button>
            </div>
          </div>
        {/if}

        {#if keseharianFormError}
          <div class="alert-error">{keseharianFormError}</div>
        {/if}

        {#if errorKeseharian}
          <div class="alert-error">{errorKeseharian}</div>
        {/if}

        {#if loadingKeseharian}
          <div class="card-loading">Memuat daftar indikator keseharian...</div>
        {:else if indikators.length === 0}
          <div class="empty-state">
            <Icon name="checkCircle" size={48} />
            <h3>Belum ada indikator keseharian</h3>
            <p>
              Tambahkan indikator (misal: Shalat Dhuha, Izin) agar musyrif dapat mengisi rapor
              keseharian siswa.
            </p>
            <button class="btn-primary" onclick={() => (showKeseharianForm = true)}>
              <Icon name="plus" size={16} /> Tambah Indikator Pertama
            </button>
          </div>
        {:else}
          <div class="table-wrapper">
            <table class="materi-table">
              <thead>
                <tr>
                  <th style="width: 50px;">No</th>
                  <th class="col-nama">Nama Indikator</th>
                  <th>Poin Minus</th>
                  <th class="col-status">Status</th>
                  <th class="col-aksi">Aksi</th>
                </tr>
              </thead>
              <tbody>
                {#each indikators as ind, i (ind.id)}
                  <tr class:row-nonaktif={!ind.aktif}>
                    {#if keseharianEditingId === ind.id}
                      <td class="text-center">{i + 1}</td>
                      <td>
                        <input
                          type="text"
                          bind:value={keseharianEditNama}
                          class="input-inline input-wide"
                          placeholder="Nama indikator"
                        />
                      </td>
                      <td>
                        <input
                          type="number"
                          bind:value={keseharianEditNilaiPengurang}
                          class="input-inline"
                          style="width: 70px"
                        />
                      </td>
                      <td>—</td>
                      <td class="aksi-cell">
                        <button
                          class="btn-save"
                          onclick={() => simpanKeseharianEdit(ind)}
                          disabled={keseharianEditLoading}
                        >
                          {keseharianEditLoading ? '...' : 'Simpan'}
                        </button>
                        <button class="btn-cancel" onclick={cancelKeseharianEdit}>Batal</button>
                      </td>
                    {:else}
                      <td class="text-center text-muted">{i + 1}</td>
                      <td class="nama-cell">
                        <span class:text-muted={!ind.aktif}>{ind.nama}</span>
                      </td>
                      <td class="text-center">
                        <span class:text-muted={!ind.aktif}>-{ind.nilai_pengurang || 1}</span>
                      </td>
                      <td class="text-center">
                        <button
                          class="toggle-btn"
                          class:aktif={ind.aktif}
                          class:nonaktif={!ind.aktif}
                          onclick={() => toggleKeseharianAktif(ind)}
                          title={ind.aktif ? 'Klik untuk nonaktifkan' : 'Klik untuk aktifkan'}
                        >
                          {ind.aktif ? 'Aktif' : 'Nonaktif'}
                        </button>
                      </td>
                      <td class="aksi-cell">
                        <button
                          class="btn-icon"
                          title="Edit"
                          onclick={() => startKeseharianEdit(ind)}
                        >
                          <Icon name="edit2" size={15} />
                        </button>
                        <button
                          class="btn-icon btn-icon-danger"
                          title="Hapus"
                          onclick={() => hapusKeseharian(ind.id)}
                          disabled={keseharianDeletingId === ind.id}
                        >
                          <Icon name="trash" size={15} />
                        </button>
                      </td>
                    {/if}
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
        <!-- ======= END TAB KESEHARIAN ======= -->
      {/if}
    {/if}
    <!-- Akhir else (!hasTahfiz && !hasDiniyah) -->
  {/if}
</div>

<style>
  @import url('https://fonts.googleapis.com/css2?family=Amiri:wght@400;700&display=swap');

  /* Tabs */
  .tabs-header {
    display: inline-flex;
    background-color: var(--color-surface, #fff);
    border: 1px solid var(--color-border, #e2e8f0);
    border-radius: 12px;
    padding: 0.35rem;
    box-shadow: var(--shadow-sm, 0 1px 2px rgba(0, 0, 0, 0.05));
    margin-bottom: 1.5rem;
  }
  .tab-btn {
    padding: 0.6rem 1.5rem;
    font-size: 0.9rem;
    font-weight: 600;
    border: none;
    background: transparent;
    color: var(--color-text-muted, #64748b);
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s;
    font-family: inherit;
  }
  .tab-btn.active {
    background-color: var(--color-primary, #4f46e5);
    color: white;
  }

  /* Arabic */
  .input-arab,
  .arab-cell {
    font-family: 'Amiri', 'Traditional Arabic', serif;
    font-size: 1.2rem;
  }
  .arab-cell {
    text-align: right;
  }
  .page-container {
    max-width: 900px;
    margin: 0 auto;
    padding-bottom: 3rem;
  }

  .page-header {
    margin-bottom: 1.75rem;
  }

  .page-title {
    margin: 0;
    font-size: 1.5rem;
    font-weight: 800;
    color: #0f172a;
  }

  .page-subtitle {
    margin: 0.3rem 0 0;
    color: #64748b;
    font-size: 0.9rem;
  }

  /* Sekolah picker */
  .section-label {
    font-size: 0.9rem;
    font-weight: 600;
    color: #475569;
    margin-bottom: 0.75rem;
  }

  .level-option span {
    font-size: 0.95rem;
    font-weight: 500;
  }

  /* Kriteria */
  .settings-kriteria-block {
    margin-top: 2rem;
    background: #f8fafc;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    padding: 1.25rem;
  }

  .settings-kriteria-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
    border-bottom: 1px solid #cbd5e1;
    padding-bottom: 0.75rem;
  }

  .settings-kriteria-header .settings-class-label {
    margin-bottom: 0;
  }

  .kriteria-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .kriteria-row {
    display: flex;
    gap: 0.75rem;
    align-items: flex-end;
    background: #fff;
    padding: 1rem;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    flex-wrap: wrap;
  }

  .kriteria-row .form-group {
    flex: 1 1 100px;
  }

  .kriteria-row .kriteria-arab,
  .kriteria-row .kriteria-indo {
    flex: 2 1 150px;
  }

  .kriteria-arab input {
    font-family: 'Aref Ruqaa', serif;
    font-size: 1.2rem;
  }

  .kriteria-hapus {
    height: 38px;
    width: 38px;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 1px;
    background: #fee2e2;
    border-radius: 8px;
  }

  .kriteria-hapus:hover {
    background: #fecaca;
  }

  .sekolah-grid {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
  }

  .sekolah-card {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1rem 1.25rem;
    background: #fff;
    border: 1px solid #e2e8f0;
    border-radius: 10px;
    cursor: pointer;
    font-family: inherit;
    font-size: 1rem;
    text-align: left;
    transition: all 0.2s;
    color: #1e293b;
  }

  .sekolah-card:hover {
    border-color: #22c55e;
    box-shadow: 0 4px 12px rgba(34, 197, 94, 0.1);
  }

  .sekolah-icon {
    width: 40px;
    height: 40px;
    border-radius: 8px;
    background: #f0fdf4;
    color: #16a34a;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .sekolah-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.1rem;
  }

  .sekolah-nama {
    font-weight: 700;
    font-size: 0.95rem;
    color: #0f172a;
  }

  .sekolah-jenjang {
    font-size: 0.82rem;
    color: #64748b;
  }

  /* Selected header */
  .selected-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    margin-bottom: 1.25rem;
    padding: 0.75rem 1rem;
    background: #f8fafc;
    border: 1px solid #e2e8f0;
    border-radius: 10px;
  }

  .selected-sekolah {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.9rem;
    color: #1e293b;
  }

  .badge-jenjang {
    background: #dcfce7;
    color: #166534;
    font-size: 0.75rem;
    font-weight: 700;
    padding: 0.15rem 0.5rem;
    border-radius: 999px;
  }

  .btn-back {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    padding: 0.45rem 0.85rem;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    background: #fff;
    color: #475569;
    font-size: 0.85rem;
    font-weight: 600;
    cursor: pointer;
    font-family: inherit;
    transition: all 0.2s;
  }

  .btn-back:hover {
    border-color: #94a3b8;
    color: #1e293b;
  }

  /* Toolbar */
  .toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  .materi-count {
    font-size: 0.875rem;
    color: #64748b;
    font-weight: 600;
  }

  /* Form tambah */
  .form-card {
    background: #fff;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    padding: 1.25rem 1.5rem;
    margin-bottom: 1.25rem;
  }

  .form-title {
    margin: 0 0 1rem;
    font-size: 1rem;
    font-weight: 700;
    color: #0f172a;
  }

  .form-row {
    display: flex;
    gap: 0.75rem;
    align-items: flex-end;
    flex-wrap: wrap;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    flex: 0 0 110px;
  }

  .form-group-wide {
    flex: 1 1 260px;
  }

  .form-group label {
    font-size: 0.82rem;
    font-weight: 600;
    color: #475569;
  }

  .required {
    color: #ef4444;
  }

  .input {
    height: 38px;
    padding: 0 0.75rem;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    font-size: 0.9rem;
    font-family: inherit;
    color: #0f172a;
    outline: none;
    transition: border-color 0.2s;
    background: #f8fafc;
  }

  .input:focus {
    border-color: #22c55e;
    background: #fff;
  }

  .form-actions {
    display: flex;
    gap: 0.5rem;
    justify-content: flex-end;
    margin-top: 1rem;
  }

  .kelulusan-settings {
    margin-top: 1.25rem;
    padding: 1.25rem;
    background: #fff;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
  }

  .settings-header {
    display: flex;
    justify-content: space-between;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  .settings-title {
    margin: 0;
    color: #0f172a;
    font-size: 1rem;
    font-weight: 800;
  }

  .settings-subtitle {
    margin: 0.25rem 0 0;
    color: #64748b;
    font-size: 0.86rem;
  }

  .settings-row {
    display: flex;
    align-items: flex-end;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .settings-input {
    flex: 0 1 260px;
  }

  .settings-note {
    margin: 0.9rem 0 0;
    color: #64748b;
    font-size: 0.84rem;
  }

  .settings-class-block {
    margin-top: 1rem;
  }

  .settings-class-label {
    margin-bottom: 0.5rem;
    font-size: 0.82rem;
    font-weight: 700;
    color: #475569;
  }

  .level-options {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .level-option {
    display: inline-flex;
    align-items: center;
    gap: 0.45rem;
    min-height: 36px;
    padding: 0.45rem 0.75rem;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    background: #f8fafc;
    color: #1e293b;
    font-size: 0.86rem;
    font-weight: 700;
    cursor: pointer;
  }

  .level-option input {
    width: 15px;
    height: 15px;
    accent-color: #22c55e;
  }

  /* Buttons */
  .btn-primary {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.55rem 1.1rem;
    background: #22c55e;
    color: #fff;
    border: none;
    border-radius: 8px;
    font-size: 0.875rem;
    font-weight: 700;
    cursor: pointer;
    font-family: inherit;
    transition: background 0.2s;
  }

  .btn-primary:hover:not(:disabled) {
    background: #16a34a;
  }

  .btn-primary:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .btn-secondary {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.55rem 1.1rem;
    background: #fff;
    color: #475569;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    font-size: 0.875rem;
    font-weight: 600;
    cursor: pointer;
    font-family: inherit;
    transition: all 0.2s;
  }

  .btn-secondary:hover:not(:disabled) {
    border-color: #94a3b8;
    color: #1e293b;
  }

  /* Table */
  .table-wrapper {
    background: #fff;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    overflow: hidden;
  }

  .materi-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.875rem;
  }

  .materi-table th {
    background: #f8fafc;
    padding: 0.75rem 1rem;
    text-align: left;
    font-weight: 700;
    font-size: 0.8rem;
    color: #64748b;
    text-transform: uppercase;
    letter-spacing: 0.03em;
    border-bottom: 1px solid #e2e8f0;
  }

  .materi-table td {
    padding: 0.75rem 1rem;
    border-bottom: 1px solid #f1f5f9;
    vertical-align: middle;
  }

  .materi-table tbody tr:last-child td {
    border-bottom: none;
  }

  .materi-table tbody tr:hover {
    background: #f8fafc;
  }

  .row-nonaktif {
    opacity: 0.55;
  }

  .col-urutan {
    width: 80px;
  }
  .col-juz {
    width: 100px;
  }
  .col-status {
    width: 110px;
  }
  .col-aksi {
    width: 100px;
  }

  .nama-cell {
    font-weight: 600;
    color: #1e293b;
  }
  .text-center {
    text-align: center;
  }
  .text-muted {
    color: #94a3b8;
  }

  /* Toggle aktif */
  .toggle-btn {
    display: inline-block;
    padding: 0.2rem 0.65rem;
    border-radius: 999px;
    font-size: 0.78rem;
    font-weight: 700;
    cursor: pointer;
    border: none;
    font-family: inherit;
    transition: all 0.2s;
  }

  .toggle-btn.aktif {
    background: #dcfce7;
    color: #166534;
  }

  .toggle-btn.aktif:hover {
    background: #bbf7d0;
  }

  .toggle-btn.nonaktif {
    background: #f1f5f9;
    color: #64748b;
  }

  .toggle-btn.nonaktif:hover {
    background: #e2e8f0;
  }

  /* Inline edit */
  .input-inline {
    height: 32px;
    width: 70px;
    padding: 0 0.5rem;
    border: 1px solid #22c55e;
    border-radius: 6px;
    font-size: 0.875rem;
    font-family: inherit;
    outline: none;
    background: #f0fdf4;
    color: #0f172a;
  }

  .input-wide {
    width: 200px;
  }

  /* Aksi */
  .aksi-cell {
    display: flex;
    gap: 0.35rem;
    align-items: center;
  }

  .btn-icon {
    width: 30px;
    height: 30px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: 1px solid #e2e8f0;
    border-radius: 6px;
    background: #fff;
    color: #64748b;
    cursor: pointer;
    transition: all 0.2s;
    padding: 0;
  }

  .btn-icon:hover {
    border-color: #22c55e;
    color: #16a34a;
    background: #f0fdf4;
  }

  .btn-icon-danger:hover {
    border-color: #fca5a5;
    color: #dc2626;
    background: #fef2f2;
  }

  .btn-save {
    padding: 0.3rem 0.7rem;
    background: #22c55e;
    color: #fff;
    border: none;
    border-radius: 6px;
    font-size: 0.8rem;
    font-weight: 700;
    cursor: pointer;
    font-family: inherit;
  }

  .btn-cancel {
    padding: 0.3rem 0.7rem;
    background: #fff;
    color: #64748b;
    border: 1px solid #e2e8f0;
    border-radius: 6px;
    font-size: 0.8rem;
    font-weight: 600;
    cursor: pointer;
    font-family: inherit;
  }

  /* Alert & Empty */
  .alert-error {
    padding: 0.75rem 1rem;
    background: #fef2f2;
    color: #dc2626;
    border: 1px solid #fecaca;
    border-radius: 8px;
    font-size: 0.875rem;
    font-weight: 600;
    margin-bottom: 1rem;
  }

  .alert-success {
    padding: 0.75rem 1rem;
    background: #f0fdf4;
    color: #166534;
    border: 1px solid #bbf7d0;
    border-radius: 8px;
    font-size: 0.875rem;
    font-weight: 600;
    margin-bottom: 1rem;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.75rem;
    min-height: 260px;
    background: #fff;
    border: 1px dashed #cbd5e1;
    border-radius: 12px;
    color: #64748b;
    text-align: center;
    padding: 2rem;
  }

  .empty-state h3 {
    margin: 0;
    font-size: 1rem;
    color: #1e293b;
  }

  .empty-state p {
    margin: 0;
    font-size: 0.875rem;
  }

  .card-loading {
    padding: 2rem;
    text-align: center;
    color: #64748b;
    font-size: 0.9rem;
    background: #fff;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
  }

  @media (max-width: 640px) {
    .form-row {
      flex-direction: column;
    }
    .form-group {
      flex: 1 1 auto;
    }
    .toolbar {
      flex-direction: column;
      align-items: flex-start;
    }
    .materi-table {
      font-size: 0.8rem;
    }
    .col-urutan,
    .col-juz {
      display: none;
    }
  }

  .text-danger:hover {
    color: #dc2626 !important;
    background: #fee2e2 !important;
  }

  /* Modals for Keseharian */
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background: rgba(0, 0, 0, 0.5);
    z-index: 50;
    backdrop-filter: blur(2px);
  }
  .modal {
    position: fixed;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    background: white;
    width: 100%;
    max-width: 400px;
    border-radius: 12px;
    z-index: 51;
    box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
  }
  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 1.25rem;
    border-bottom: 1px solid #e2e8f0;
  }
  .modal-header h3 {
    margin: 0;
    font-size: 1.1rem;
    color: #1e293b;
  }
  .btn-close {
    background: none;
    border: none;
    font-size: 1.2rem;
    cursor: pointer;
    color: #64748b;
  }
  .modal-body {
    padding: 1.25rem;
  }
  .modal-footer {
    padding: 1rem 1.25rem;
    border-top: 1px solid #e2e8f0;
    display: flex;
    justify-content: flex-end;
    gap: 0.75rem;
  }
  .form-group-check {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }
  .form-group-check label {
    font-size: 0.9rem;
    color: #334155;
    cursor: pointer;
  }
</style>
