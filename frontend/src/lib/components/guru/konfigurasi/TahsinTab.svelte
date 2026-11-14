<script>
  import { onMount } from 'svelte';
  import Icon from '$lib/components/ui/Icon.svelte';
  import { toastStore } from '$lib/toast.svelte.js';

  let { sekolahId } = $props();

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

  // State
  let loading = $state(false);
  let saving = $state(false);
  let error = $state('');

  let selectedJilid = $state(5); // Default ke Jilid 5 yang sering diedit
  let searchQuery = $state('');

  let jilidList = $state([]);
  let indikatorList = $state([]);

  // Modal State: Indikator
  let showModalIndikator = $state(false);
  let isEditingIndikator = $state(false);
  let formIndikator = $state({
    id: 0,
    jilid: 5,
    kode: '',
    indikator: '',
    urutan: 1,
    aktif: true
  });

  // Modal State: Info Jilid
  let showModalJilid = $state(false);
  let formJilid = $state({
    jilid: 5,
    nama_jilid: '',
    fokus: '',
    kriteria_kelulusan: ''
  });

  // Modal State: Confirm Reset
  let showConfirmReset = $state(false);

  // Derived
  let currentJilidInfo = $derived(
    jilidList.find((j) => j.jilid === selectedJilid) || {
      jilid: selectedJilid,
      nama_jilid: `Jilid ${selectedJilid}`,
      fokus: '-',
      kriteria_kelulusan: '-'
    }
  );

  let filteredIndikator = $derived(
    indikatorList
      .filter((item) => {
        if (selectedJilid !== 0 && item.jilid !== selectedJilid) return false;
        if (!searchQuery.trim()) return true;
        const q = searchQuery.trim().toLowerCase();
        return (
          (item.kode || '').toLowerCase().includes(q) ||
          (item.indikator || '').toLowerCase().includes(q)
        );
      })
      .sort((a, b) => {
        if (a.jilid !== b.jilid) return a.jilid - b.jilid;
        return (a.urutan || 0) - (b.urutan || 0);
      })
  );

  function countPerJilid(j) {
    return indikatorList.filter((item) => item.jilid === j).length;
  }

  // Load Data
  $effect(() => {
    if (sekolahId) {
      loadData();
    }
  });

  async function loadData() {
    if (!sekolahId) return;
    loading = true;
    error = '';
    try {
      const [resJilid, resIndikator] = await Promise.all([
        fetch(`${apiUrl}/api/guru/sekolah/${sekolahId}/tahsin/jilid`, { credentials: 'include' }),
        fetch(`${apiUrl}/api/guru/sekolah/${sekolahId}/tahsin/indikator`, { credentials: 'include' })
      ]);

      if (!resJilid.ok) {
        const body = await resJilid.json().catch(() => ({}));
        throw new Error(body.message || 'Gagal memuat data jilid tahsin');
      }
      if (!resIndikator.ok) {
        const body = await resIndikator.json().catch(() => ({}));
        throw new Error(body.message || 'Gagal memuat data indikator tahsin');
      }

      jilidList = await resJilid.json();
      indikatorList = await resIndikator.json();
    } catch (err) {
      error = err.message || 'Terjadi kesalahan saat memuat konfigurasi tahsin';
    } finally {
      loading = false;
    }
  }

  // Action: Open Modal Tambah Indikator
  function openTambahIndikator() {
    isEditingIndikator = false;
    const targetJilid = selectedJilid === 0 ? 1 : selectedJilid;
    const count = countPerJilid(targetJilid);
    formIndikator = {
      id: 0,
      jilid: targetJilid,
      kode: `J${targetJilid}.${count + 1}`,
      indikator: '',
      urutan: count + 1,
      aktif: true
    };
    showModalIndikator = true;
  }

  // Action: Open Modal Edit Indikator
  function openEditIndikator(item) {
    isEditingIndikator = true;
    formIndikator = {
      id: item.id,
      jilid: item.jilid,
      kode: item.kode,
      indikator: item.indikator,
      urutan: item.urutan || 1,
      aktif: item.aktif
    };
    showModalIndikator = true;
  }

  // Action: Save Indikator (Create / Update)
  async function saveIndikator() {
    const trimmedKode = formIndikator.kode.trim();
    const trimmedIndikator = formIndikator.indikator.trim();
    if (!trimmedKode || !trimmedIndikator) {
      toastStore.show('Kode dan teks indikator wajib diisi', 'error');
      return;
    }

    saving = true;
    try {
      const url = isEditingIndikator
        ? `${apiUrl}/api/guru/tahsin/indikator/${formIndikator.id}`
        : `${apiUrl}/api/guru/sekolah/${sekolahId}/tahsin/indikator`;
      const method = isEditingIndikator ? 'PUT' : 'POST';

      const payload = {
        jilid: Number(formIndikator.jilid) || 1,
        kode: trimmedKode,
        indikator: trimmedIndikator,
        urutan: Number(formIndikator.urutan) || 1,
        aktif: Boolean(formIndikator.aktif)
      };

      const res = await fetch(url, {
        method,
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });

      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.message || 'Gagal menyimpan indikator');
      }

      toastStore.show(
        isEditingIndikator ? 'Indikator berhasil diperbarui' : 'Indikator baru berhasil ditambahkan',
        'success'
      );
      showModalIndikator = false;
      await loadData();
    } catch (err) {
      toastStore.show(err.message, 'error');
    } finally {
      saving = false;
    }
  }

  // Action: Toggle Aktif Indikator
  async function toggleAktif(item) {
    try {
      const res = await fetch(`${apiUrl}/api/guru/tahsin/indikator/${item.id}`, {
        method: 'PUT',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ aktif: !item.aktif })
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.message || 'Gagal mengubah status');
      }
      toastStore.show(
        `Indikator ${item.kode} ${!item.aktif ? 'diaktifkan' : 'dinonaktifkan'}`,
        'success'
      );
      await loadData();
    } catch (err) {
      toastStore.show(err.message, 'error');
    }
  }

  // Action: Delete Indikator
  async function deleteIndikator(item) {
    if (!confirm(`Yakin ingin menghapus indikator ${item.kode}: "${item.indikator}"?`)) {
      return;
    }

    try {
      const res = await fetch(`${apiUrl}/api/guru/tahsin/indikator/${item.id}`, {
        method: 'DELETE',
        credentials: 'include'
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.message || 'Gagal menghapus indikator');
      }
      toastStore.show(`Indikator ${item.kode} berhasil dihapus`, 'success');
      await loadData();
    } catch (err) {
      toastStore.show(err.message, 'error');
    }
  }

  // Action: Open Modal Edit Jilid
  function openEditJilid() {
    formJilid = {
      jilid: currentJilidInfo.jilid,
      nama_jilid: currentJilidInfo.nama_jilid,
      fokus: currentJilidInfo.fokus || '',
      kriteria_kelulusan: currentJilidInfo.kriteria_kelulusan || ''
    };
    showModalJilid = true;
  }

  // Action: Save Info Jilid
  async function saveJilid() {
    saving = true;
    try {
      const payload = {
        nama_jilid: (formJilid.nama_jilid || '').trim(),
        fokus: (formJilid.fokus || '').trim(),
        kriteria_kelulusan: (formJilid.kriteria_kelulusan || '').trim()
      };

      const res = await fetch(
        `${apiUrl}/api/guru/sekolah/${sekolahId}/tahsin/jilid/${formJilid.jilid}`,
        {
          method: 'PUT',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        }
      );
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.message || 'Gagal menyimpan info jilid');
      }
      toastStore.show(`Info Jilid ${formJilid.jilid} berhasil diperbarui`, 'success');
      showModalJilid = false;
      await loadData();
    } catch (err) {
      toastStore.show(err.message, 'error');
    } finally {
      saving = false;
    }
  }

  // Action: Reset Template Standar
  async function resetTemplate() {
    saving = true;
    try {
      const res = await fetch(`${apiUrl}/api/guru/sekolah/${sekolahId}/tahsin/reset-template`, {
        method: 'POST',
        credentials: 'include'
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.message || 'Gagal mereset template');
      }
      toastStore.show('Konfigurasi berhasil dikembalikan ke template standar Iqro 1-6', 'success');
      showConfirmReset = false;
      await loadData();
    } catch (err) {
      toastStore.show(err.message, 'error');
    } finally {
      saving = false;
    }
  }
</script>

<div class="tahsin-config-container">
  <!-- HEADER & SUBTITLE -->
  <div class="config-header-row">
    <div>
      <h2 class="config-title">Konfigurasi Indikator Penilaian Tahsin</h2>
      <p class="config-subtitle">
        Kelola standar indikator penilaian, fokus pembelajaran, dan kriteria kelulusan per jilid (Iqro 1 s.d 6).
      </p>
    </div>
    <div class="config-actions">
      <button
        type="button"
        class="btn-secondary"
        onclick={() => (showConfirmReset = true)}
        title="Kembalikan indikator ke susunan template bawaan"
      >
        <Icon name="clock" size={15} /> Reset ke Template Standar
      </button>
      <button type="button" class="btn-primary" onclick={openTambahIndikator}>
        <Icon name="plus" size={16} /> Tambah Indikator
      </button>
    </div>
  </div>

  {#if error}
    <div class="alert-error">{error}</div>
  {/if}

  <!-- JILID SELECTOR TABS -->
  <div class="jilid-selector-tabs">
    {#each [1, 2, 3, 4, 5, 6] as j}
      <button
        type="button"
        class="jilid-tab-btn"
        class:active={selectedJilid === j}
        onclick={() => (selectedJilid = j)}
      >
        <span class="jilid-num">Jilid {j}</span>
        <span class="jilid-pill-badge">{countPerJilid(j)}</span>
      </button>
    {/each}
    <button
      type="button"
      class="jilid-tab-btn"
      class:active={selectedJilid === 0}
      onclick={() => (selectedJilid = 0)}
    >
      <span class="jilid-num">Semua Jilid</span>
      <span class="jilid-pill-badge">{indikatorList.length}</span>
    </button>
  </div>

  <!-- JILID INFO CARD (Hanya tampil saat jilid 1-6 dipilih) -->
  {#if selectedJilid > 0}
    <div class="jilid-hero-card">
      <div class="jilid-hero-header">
        <div class="jilid-badge-col">
          <div class="jilid-circle-badge">J{selectedJilid}</div>
          <div>
            <h3 class="jilid-hero-title">{currentJilidInfo.nama_jilid || `Jilid ${selectedJilid}`}</h3>
            <span class="jilid-hero-fokus">
              <strong>Fokus:</strong> {currentJilidInfo.fokus || 'Belum diatur'}
            </span>
          </div>
        </div>
        <button type="button" class="btn-edit-jilid" onclick={openEditJilid}>
          <Icon name="edit" size={14} /> Edit Info Jilid
        </button>
      </div>

      {#if currentJilidInfo.kriteria_kelulusan}
        <div class="jilid-hero-kriteria">
          <span class="kriteria-label">Kriteria Kelulusan:</span>
          <p class="kriteria-text">{currentJilidInfo.kriteria_kelulusan}</p>
        </div>
      {/if}
    </div>
  {/if}

  <!-- TOOLBAR: SEARCH & COUNT -->
  <div class="table-toolbar">
    <div class="search-box">
      <Icon name="search" size={16} class="search-icon" />
      <input
        type="text"
        placeholder="Cari kode atau teks indikator..."
        class="form-input search-input"
        bind:value={searchQuery}
      />
      {#if searchQuery}
        <button type="button" class="search-clear" onclick={() => (searchQuery = '')}>×</button>
      {/if}
    </div>

    <div class="table-count-info">
      Menampilkan <strong>{filteredIndikator.length}</strong> indikator
      {#if selectedJilid > 0}
        di Jilid {selectedJilid}
      {/if}
    </div>
  </div>

  <!-- TABLE INDIKATOR -->
  <div class="table-card">
    {#if loading}
      <div class="loading-state">
        <div class="spinner"></div>
        <span>Memuat indikator tahsin...</span>
      </div>
    {:else if filteredIndikator.length > 0}
      <div class="table-responsive">
        <table class="data-table">
          <thead>
            <tr>
              <th style="width: 60px; text-align: center;">No</th>
              {#if selectedJilid === 0}
                <th style="width: 90px; text-align: center;">Jilid</th>
              {/if}
              <th style="width: 100px;">Kode</th>
              <th>Indikator Penilaian</th>
              <th style="width: 100px; text-align: center;">Status</th>
              <th style="width: 130px; text-align: right;">Aksi</th>
            </tr>
          </thead>
          <tbody>
            {#each filteredIndikator as item, idx (item.id)}
              <tr class:row-inactive={!item.aktif}>
                <td style="text-align: center; color: #94a3b8; font-weight: 600;">
                  {item.urutan || idx + 1}
                </td>
                {#if selectedJilid === 0}
                  <td style="text-align: center;">
                    <span class="jilid-chip">Jilid {item.jilid}</span>
                  </td>
                {/if}
                <td>
                  <span class="kode-badge">{item.kode}</span>
                </td>
                <td>
                  <span class="indikator-text">{item.indikator}</span>
                </td>
                <td style="text-align: center;">
                  <button
                    type="button"
                    class="status-pill-btn"
                    class:status-active={item.aktif}
                    class:status-inactive={!item.aktif}
                    onclick={() => toggleAktif(item)}
                    title="Klik untuk ubah status"
                  >
                    {item.aktif ? 'Aktif' : 'Nonaktif'}
                  </button>
                </td>
                <td style="text-align: right;">
                  <div class="action-btns">
                    <button
                      type="button"
                      class="btn-icon btn-edit"
                      onclick={() => openEditIndikator(item)}
                      title="Edit Indikator"
                    >
                      <Icon name="edit" size={14} />
                    </button>
                    <button
                      type="button"
                      class="btn-icon btn-delete"
                      onclick={() => deleteIndikator(item)}
                      title="Hapus Indikator"
                    >
                      <Icon name="trash" size={14} />
                    </button>
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      <div class="empty-state">
        <Icon name="bookmark" size={44} />
        <h4>Belum Ada Indikator</h4>
        <p>
          {#if searchQuery}
            Tidak ada indikator yang cocok dengan "{searchQuery}".
          {:else if selectedJilid > 0}
            Belum ada indikator untuk Jilid {selectedJilid}.
          {:else}
            Belum ada data indikator penilaian tahsin.
          {/if}
        </p>
        <button type="button" class="btn-primary" onclick={openTambahIndikator}>
          <Icon name="plus" size={15} /> Tambah Indikator Sekarang
        </button>
      </div>
    {/if}
  </div>
</div>

<!-- MODAL TAMBAH / EDIT INDIKATOR -->
{#if showModalIndikator}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="modal-overlay" onclick={() => (showModalIndikator = false)}>
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="modal-dialog" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <h3>{isEditingIndikator ? 'Edit Indikator Penilaian' : 'Tambah Indikator Baru'}</h3>
        <button type="button" class="btn-close" onclick={() => (showModalIndikator = false)}>×</button>
      </div>

      <div class="modal-body">
        <div class="form-group-row">
          <div class="form-field">
            <label for="modal-jilid">Jilid (Iqro) *</label>
            <select
              id="modal-jilid"
              class="form-select"
              bind:value={formIndikator.jilid}
              onchange={() => {
                if (!isEditingIndikator) {
                  const count = countPerJilid(formIndikator.jilid);
                  formIndikator.kode = `J${formIndikator.jilid}.${count + 1}`;
                  formIndikator.urutan = count + 1;
                }
              }}
            >
              {#each [1, 2, 3, 4, 5, 6] as j}
                <option value={j}>Jilid {j}</option>
              {/each}
            </select>
          </div>

          <div class="form-field">
            <label for="modal-kode">Kode Indikator *</label>
            <input
              id="modal-kode"
              type="text"
              class="form-input"
              placeholder="Contoh: J5.1"
              bind:value={formIndikator.kode}
            />
          </div>

          <div class="form-field" style="max-width: 90px;">
            <label for="modal-urutan">Urutan</label>
            <input
              id="modal-urutan"
              type="number"
              class="form-input"
              min="1"
              bind:value={formIndikator.urutan}
            />
          </div>
        </div>

        <div class="form-field" style="margin-top: 1rem;">
          <label for="modal-indikator">Deskripsi Indikator Penilaian *</label>
          <textarea
            id="modal-indikator"
            rows="3"
            class="form-textarea"
            placeholder="Contoh: Membaca alif lam qamariyah dengan benar."
            bind:value={formIndikator.indikator}
          ></textarea>
        </div>

        <div class="form-field-checkbox" style="margin-top: 0.85rem;">
          <label class="checkbox-label">
            <input type="checkbox" bind:checked={formIndikator.aktif} />
            <span>Aktif (digunakan dalam lembar penilaian)</span>
          </label>
        </div>
      </div>

      <div class="modal-footer">
        <button
          type="button"
          class="btn-cancel"
          onclick={() => (showModalIndikator = false)}
          disabled={saving}
        >
          Batal
        </button>
        <button
          type="button"
          class="btn-primary"
          onclick={saveIndikator}
          disabled={saving}
        >
          {saving ? 'Menyimpan...' : isEditingIndikator ? 'Simpan Perubahan' : 'Tambahkan'}
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- MODAL EDIT INFO JILID -->
{#if showModalJilid}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="modal-overlay" onclick={() => (showModalJilid = false)}>
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="modal-dialog" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <h3>Edit Informasi Jilid {formJilid.jilid}</h3>
        <button type="button" class="btn-close" onclick={() => (showModalJilid = false)}>×</button>
      </div>

      <div class="modal-body">
        <div class="form-field">
          <label for="modal-nama-jilid">Nama Jilid</label>
          <input
            id="modal-nama-jilid"
            type="text"
            class="form-input"
            placeholder="Contoh: Jilid 5"
            bind:value={formJilid.nama_jilid}
          />
        </div>

        <div class="form-field" style="margin-top: 1rem;">
          <label for="modal-fokus">Fokus Pembelajaran</label>
          <input
            id="modal-fokus"
            type="text"
            class="form-input"
            placeholder="Contoh: Penerapan hukum bacaan"
            bind:value={formJilid.fokus}
          />
        </div>

        <div class="form-field" style="margin-top: 1rem;">
          <label for="modal-kriteria">Kriteria Kelulusan</label>
          <textarea
            id="modal-kriteria"
            rows="3"
            class="form-textarea"
            placeholder="Contoh: Siswa mampu menerapkan pola bacaan tajwid yang dipelajari tanpa harus selalu diingatkan guru..."
            bind:value={formJilid.kriteria_kelulusan}
          ></textarea>
        </div>
      </div>

      <div class="modal-footer">
        <button
          type="button"
          class="btn-cancel"
          onclick={() => (showModalJilid = false)}
          disabled={saving}
        >
          Batal
        </button>
        <button
          type="button"
          class="btn-primary"
          onclick={saveJilid}
          disabled={saving}
        >
          {saving ? 'Menyimpan...' : 'Simpan Info Jilid'}
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- MODAL CONFIRM RESET -->
{#if showConfirmReset}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="modal-overlay" onclick={() => (showConfirmReset = false)}>
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="modal-dialog" style="max-width: 480px;" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <h3 style="color: #b91c1c;">Reset ke Template Standar?</h3>
        <button type="button" class="btn-close" onclick={() => (showConfirmReset = false)}>×</button>
      </div>

      <div class="modal-body">
        <p style="font-size: 0.9rem; color: #475569; line-height: 1.5; margin: 0 0 1rem 0;">
          Tindakan ini akan <strong>menghapus seluruh penyesuaian indikator</strong> pada sekolah ini dan mengembalikannya ke template standar kurikulum Iqro 1 s.d 6 (termasuk indikator J5.1 - J5.8 dan J6.1 - J6.10).
        </p>
        <div style="background: #fffbeb; border: 1px solid #fde68a; border-radius: 8px; padding: 0.75rem 1rem; font-size: 0.82rem; color: #92400e;">
          ⚠️ <strong>Peringatan:</strong> Pastikan Anda telah mempertimbangkan sebelum melanjutkan.
        </div>
      </div>

      <div class="modal-footer">
        <button
          type="button"
          class="btn-cancel"
          onclick={() => (showConfirmReset = false)}
          disabled={saving}
        >
          Batal
        </button>
        <button
          type="button"
          class="btn-danger"
          onclick={resetTemplate}
          disabled={saving}
        >
          {saving ? 'Mereset...' : 'Ya, Reset ke Template'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .tahsin-config-container {
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
  }

  .config-header-row {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .config-title {
    margin: 0;
    font-size: 1.25rem;
    font-weight: 800;
    color: #0f172a;
  }

  .config-subtitle {
    margin: 0.25rem 0 0;
    font-size: 0.88rem;
    color: #64748b;
  }

  .config-actions {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  /* BUTTONS */
  .btn-primary {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.55rem 1rem;
    border-radius: 8px;
    background: #0b7c3d;
    color: #ffffff;
    border: none;
    font-weight: 700;
    font-size: 0.85rem;
    cursor: pointer;
    transition: background 0.2s ease;
  }

  .btn-primary:hover {
    background: #096833;
  }

  .btn-secondary {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.55rem 0.9rem;
    border-radius: 8px;
    background: #ffffff;
    color: #475569;
    border: 1px solid #cbd5e1;
    font-weight: 700;
    font-size: 0.82rem;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .btn-secondary:hover {
    background: #f8fafc;
    border-color: #94a3b8;
    color: #0f172a;
  }

  .btn-danger {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.55rem 1rem;
    border-radius: 8px;
    background: #dc2626;
    color: #ffffff;
    border: none;
    font-weight: 700;
    font-size: 0.85rem;
    cursor: pointer;
    transition: background 0.2s ease;
  }

  .btn-danger:hover {
    background: #b91c1c;
  }

  .btn-cancel {
    padding: 0.55rem 1rem;
    border-radius: 8px;
    background: #f1f5f9;
    color: #475569;
    border: 1px solid #cbd5e1;
    font-weight: 700;
    font-size: 0.85rem;
    cursor: pointer;
  }

  .btn-cancel:hover {
    background: #e2e8f0;
  }

  /* JILID SELECTOR TABS */
  .jilid-selector-tabs {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-wrap: wrap;
    background: #f8fafc;
    padding: 0.4rem;
    border-radius: 12px;
    border: 1px solid #e2e8f0;
  }

  .jilid-tab-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.45rem 0.9rem;
    border-radius: 8px;
    font-size: 0.85rem;
    font-weight: 700;
    color: #64748b;
    border: none;
    background: transparent;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .jilid-tab-btn:hover {
    color: #0f172a;
    background: rgba(255, 255, 255, 0.6);
  }

  .jilid-tab-btn.active {
    background: #ffffff;
    color: #0b7c3d;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
  }

  .jilid-pill-badge {
    font-size: 0.72rem;
    padding: 1px 6px;
    border-radius: 9999px;
    background: #e2e8f0;
    color: #475569;
    font-weight: 800;
  }

  .jilid-tab-btn.active .jilid-pill-badge {
    background: #dcfce7;
    color: #15803d;
  }

  /* JILID HERO CARD */
  .jilid-hero-card {
    background: #ffffff;
    border: 1px solid #e2e8f0;
    border-left: 4px solid #0b7c3d;
    border-radius: 12px;
    padding: 1.15rem 1.4rem;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.02);
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .jilid-hero-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .jilid-badge-col {
    display: flex;
    align-items: center;
    gap: 0.85rem;
  }

  .jilid-circle-badge {
    width: 42px;
    height: 42px;
    border-radius: 10px;
    background: #e6f7ec;
    color: #0b7c3d;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 800;
    font-size: 1rem;
    border: 1px solid #bbf7d0;
  }

  .jilid-hero-title {
    margin: 0;
    font-size: 1.15rem;
    font-weight: 800;
    color: #0f172a;
  }

  .jilid-hero-fokus {
    font-size: 0.85rem;
    color: #475569;
  }

  .btn-edit-jilid {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    background: #f8fafc;
    border: 1px solid #cbd5e1;
    color: #334155;
    padding: 0.35rem 0.75rem;
    border-radius: 6px;
    font-size: 0.8rem;
    font-weight: 700;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .btn-edit-jilid:hover {
    background: #e2e8f0;
    color: #0f172a;
  }

  .jilid-hero-kriteria {
    background: #f8fafc;
    border-radius: 8px;
    padding: 0.65rem 0.9rem;
    border: 1px dashed #cbd5e1;
  }

  .kriteria-label {
    display: block;
    font-size: 0.75rem;
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: #0369a1;
    margin-bottom: 0.2rem;
  }

  .kriteria-text {
    margin: 0;
    font-size: 0.85rem;
    color: #334155;
    line-height: 1.45;
  }

  /* TOOLBAR */
  .table-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .search-box {
    position: relative;
    display: flex;
    align-items: center;
    min-width: 260px;
  }

  .search-box :global(.search-icon) {
    position: absolute;
    left: 0.75rem;
    color: #94a3b8;
    pointer-events: none;
  }

  .search-input {
    width: 100%;
    padding: 0.5rem 1.8rem 0.5rem 2.2rem;
    border-radius: 8px;
    border: 1px solid #cbd5e1;
    background: #ffffff;
    font-size: 0.85rem;
    outline: none;
  }

  .search-input:focus {
    border-color: #0b7c3d;
    box-shadow: 0 0 0 3px rgba(11, 124, 61, 0.1);
  }

  .search-clear {
    position: absolute;
    right: 0.5rem;
    background: transparent;
    border: none;
    color: #94a3b8;
    font-size: 1rem;
    cursor: pointer;
  }

  .table-count-info {
    font-size: 0.85rem;
    color: #64748b;
  }

  /* TABLE CARD */
  .table-card {
    background: #ffffff;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    overflow: hidden;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.02);
  }

  .table-responsive {
    width: 100%;
    overflow-x: auto;
  }

  .data-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.85rem;
    text-align: left;
  }

  .data-table th {
    background: #f8fafc;
    color: #475569;
    font-weight: 700;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid #e2e8f0;
    font-size: 0.78rem;
    text-transform: uppercase;
    letter-spacing: 0.03em;
  }

  .data-table td {
    padding: 0.8rem 1rem;
    border-bottom: 1px solid #f1f5f9;
    color: #1e293b;
    vertical-align: middle;
  }

  .data-table tr:hover td {
    background: #fafafa;
  }

  .data-table tr.row-inactive td {
    opacity: 0.6;
    background: #fafafa;
  }

  .kode-badge {
    display: inline-block;
    padding: 0.2rem 0.55rem;
    border-radius: 6px;
    background: #f1f5f9;
    color: #0f172a;
    font-weight: 800;
    font-size: 0.82rem;
    border: 1px solid #e2e8f0;
    font-family: monospace;
  }

  .indikator-text {
    font-weight: 600;
    color: #1e293b;
    line-height: 1.4;
  }

  .jilid-chip {
    display: inline-block;
    padding: 0.15rem 0.5rem;
    border-radius: 9999px;
    background: #e0f2fe;
    color: #0369a1;
    font-size: 0.75rem;
    font-weight: 700;
  }

  .status-pill-btn {
    border: none;
    padding: 0.2rem 0.65rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 700;
    cursor: pointer;
    transition: opacity 0.15s ease;
  }

  .status-pill-btn:hover {
    opacity: 0.8;
  }

  .status-active {
    background: #dcfce7;
    color: #15803d;
  }

  .status-inactive {
    background: #f1f5f9;
    color: #94a3b8;
  }

  .action-btns {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 0.35rem;
  }

  .btn-icon {
    width: 30px;
    height: 30px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 6px;
    border: 1px solid #e2e8f0;
    background: #ffffff;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .btn-icon.btn-edit {
    color: #0284c7;
  }

  .btn-icon.btn-edit:hover {
    background: #e0f2fe;
    border-color: #7dd3fc;
  }

  .btn-icon.btn-delete {
    color: #dc2626;
  }

  .btn-icon.btn-delete:hover {
    background: #fee2e2;
    border-color: #fca5a5;
  }

  /* MODALS */
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background: rgba(15, 23, 42, 0.45);
    backdrop-filter: blur(2px);
    z-index: 9999;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem;
  }

  .modal-dialog {
    background: #ffffff;
    width: 100%;
    max-width: 560px;
    border-radius: 14px;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 8px 10px -6px rgba(0, 0, 0, 0.1);
    overflow: hidden;
    display: flex;
    flex-direction: column;
    animation: modalPop 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes modalPop {
    from { transform: scale(0.96); opacity: 0; }
    to { transform: scale(1); opacity: 1; }
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.15rem 1.4rem;
    border-bottom: 1px solid #e2e8f0;
  }

  .modal-header h3 {
    margin: 0;
    font-size: 1.1rem;
    font-weight: 800;
    color: #0f172a;
  }

  .btn-close {
    background: transparent;
    border: none;
    font-size: 1.4rem;
    color: #94a3b8;
    cursor: pointer;
    line-height: 1;
  }

  .btn-close:hover {
    color: #0f172a;
  }

  .modal-body {
    padding: 1.25rem 1.4rem;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    align-items: center;
    gap: 0.75rem;
    padding: 1rem 1.4rem;
    border-top: 1px solid #e2e8f0;
    background: #f8fafc;
  }

  .form-group-row {
    display: flex;
    gap: 0.75rem;
    align-items: flex-start;
  }

  .form-field {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }

  .form-field label {
    font-size: 0.8rem;
    font-weight: 700;
    color: #475569;
  }

  .form-input,
  .form-select,
  .form-textarea {
    width: 100%;
    padding: 0.55rem 0.85rem;
    border-radius: 8px;
    border: 1px solid #cbd5e1;
    background: #ffffff;
    font-size: 0.85rem;
    font-family: inherit;
    outline: none;
    color: #0f172a;
  }

  .form-input:focus,
  .form-select:focus,
  .form-textarea:focus {
    border-color: #0b7c3d;
    box-shadow: 0 0 0 3px rgba(11, 124, 61, 0.1);
  }

  .form-field-checkbox {
    display: flex;
    align-items: center;
  }

  .checkbox-label {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.85rem;
    font-weight: 600;
    color: #334155;
    cursor: pointer;
  }

  /* LOADING & EMPTY */
  .loading-state,
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem 1.5rem;
    text-align: center;
    color: #64748b;
  }

  .empty-state h4 {
    margin: 0.75rem 0 0.25rem;
    color: #0f172a;
    font-size: 1.1rem;
    font-weight: 800;
  }

  .empty-state p {
    margin: 0 0 1.25rem;
    font-size: 0.85rem;
  }

  .spinner {
    width: 32px;
    height: 32px;
    border: 3px solid #e2e8f0;
    border-top-color: #0b7c3d;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
    margin-bottom: 0.75rem;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .alert-error {
    background: #fef2f2;
    border: 1px solid #fecaca;
    color: #dc2626;
    padding: 0.75rem 1rem;
    border-radius: 8px;
    font-size: 0.85rem;
    font-weight: 600;
  }
</style>
