const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'frontend/src/routes/guru/konfigurasi/+page.svelte');
let content = fs.readFileSync(file, 'utf8');

// Add states
const keseharianStates = `
  let keseharianKriteriaPenilaian = $state([]);
  let keseharianKonfigurasiLoading = $state(false);
  let keseharianKonfigurasiSaving = $state(false);
  let keseharianKonfigurasiError = $state('');
  let keseharianKonfigurasiSuccess = $state('');
`;

if (!content.includes('let keseharianKriteriaPenilaian = $state([]);')) {
    content = content.replace("let indikators = $state([]);", keseharianStates + "\n  let indikators = $state([]);");
}

// Add functions
const keseharianFns = `
  async function loadKonfigurasiKeseharian() {
    if (!selectedSekolah) return;
    keseharianKonfigurasiLoading = true;
    keseharianKonfigurasiError = '';
    try {
      const res = await fetch(\`\${apiUrl}/api/guru/sekolah/\${selectedSekolah.id}/konfigurasi-nilai-keseharian\`, {
        credentials: 'include'
      });
      if (res.ok) {
        const body = await res.json();
        keseharianKriteriaPenilaian = body.kriteria_penilaian || [];
      } else {
        keseharianKriteriaPenilaian = [];
      }
    } catch (err) {
      console.error(err);
      keseharianKriteriaPenilaian = [];
    } finally {
      keseharianKonfigurasiLoading = false;
    }
  }

  async function simpanKonfigurasiNilaiKeseharian() {
    keseharianKonfigurasiSaving = true;
    keseharianKonfigurasiError = '';
    keseharianKonfigurasiSuccess = '';
    try {
      const res = await fetch(
        \`\${apiUrl}/api/guru/sekolah/\${selectedSekolah.id}/konfigurasi-nilai-keseharian\`,
        {
          method: 'POST',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            kriteria_penilaian: keseharianKriteriaPenilaian
          })
        }
      );
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.message || 'Gagal menyimpan pengaturan kriteria keseharian');
      }
      const payload = await res.json();
      keseharianKriteriaPenilaian = payload.kriteria_penilaian || [];
      keseharianKonfigurasiSuccess = 'Pengaturan berhasil disimpan';
      toastStore.show('Pengaturan Keseharian berhasil disimpan', 'success');
      setTimeout(() => (keseharianKonfigurasiSuccess = ''), 3000);
    } catch (err) {
      keseharianKonfigurasiError = err.message;
      toastStore.show(err.message, 'error');
    } finally {
      keseharianKonfigurasiSaving = false;
    }
  }

  function tambahKriteriaKeseharian() {
    keseharianKriteriaPenilaian = [
      ...keseharianKriteriaPenilaian,
      { min: 0, max: 0, indo: '', arab: '' }
    ];
  }

  function hapusKriteriaKeseharian(index) {
    if (confirm('Hapus kriteria ini?')) {
      keseharianKriteriaPenilaian = keseharianKriteriaPenilaian.filter((_, i) => i !== index);
    }
  }
`;

if (!content.includes('loadKonfigurasiKeseharian()')) {
    content = content.replace("async function loadData() {", keseharianFns + "\n\n  async function loadData() {");
}

// Add loadKonfigurasiKeseharian to loadData()
if (!content.includes('await loadKonfigurasiKeseharian();')) {
    content = content.replace("await loadIqob();", "await loadIqob();\n      await loadKonfigurasiKeseharian();");
}

// Replace the UI block for activeSubTab === 'kriteria'
const oldUi = "{:else if activeSubTab === 'kriteria'}\n      {/if}";
const newUi = `{:else if activeSubTab === 'kriteria'}
        <!-- ======= TAB KRITERIA KESEHARIAN ======= -->
        <section class="kelulusan-settings" style="margin-top: 1rem;">
          <div class="settings-header">
            <div>
              <h2 class="settings-title">Kriteria Penilaian Keseharian</h2>
              <p class="settings-subtitle">Atur batas predikat (Misal: 90 - 100 = Sangat Baik). Pastikan predikat sama persis dengan yang ada di Bank Catatan Admin.</p>
            </div>
          </div>

          {#if keseharianKonfigurasiError}
            <div class="alert-error">{keseharianKonfigurasiError}</div>
          {/if}
          {#if keseharianKonfigurasiSuccess}
            <div class="alert-success">{keseharianKonfigurasiSuccess}</div>
          {/if}

          <div class="settings-kriteria-block" style="margin-top: 1rem;">
            <div class="settings-kriteria-header">
              <div class="settings-class-label">Kriteria Penilaian Keseharian</div>
              <button
                class="btn-secondary btn-sm"
                onclick={tambahKriteriaKeseharian}
                disabled={keseharianKonfigurasiLoading || keseharianKonfigurasiSaving}
              >
                <Icon name="plus" size={14} /> Tambah
              </button>
            </div>

            {#if keseharianKonfigurasiLoading}
              <div style="padding: 1rem; color: #64748b; font-weight: 500;">Memuat kriteria...</div>
            {:else if keseharianKriteriaPenilaian.length === 0}
              <div style="padding: 1rem; color: #64748b; font-weight: 500;">Belum ada kriteria penilaian. Silakan klik Tambah.</div>
            {:else}
              <div class="kriteria-list">
                <div class="kriteria-row kriteria-header" style="font-weight: 600; color: #475569; background: #f8fafc; padding: 0.5rem; border-radius: 6px;">
                  <div>Batas Bawah</div>
                  <div>Batas Atas</div>
                  <div>Predikat (Arab)</div>
                  <div>Predikat (Indo)</div>
                  <div style="width: 40px;"></div>
                </div>
                {#each keseharianKriteriaPenilaian as k, i}
                  <div class="kriteria-row">
                    <input type="number" min="0" max="100" class="form-input" bind:value={k.min} disabled={keseharianKonfigurasiLoading || keseharianKonfigurasiSaving} />
                    <input type="number" min="0" max="100" class="form-input" bind:value={k.max} disabled={keseharianKonfigurasiLoading || keseharianKonfigurasiSaving} />
                    <input type="text" class="form-input arabic-text" dir="rtl" placeholder="مُمْتَاز" bind:value={k.arab} disabled={keseharianKonfigurasiLoading || keseharianKonfigurasiSaving} />
                    <input type="text" class="form-input" placeholder="Sangat Baik" bind:value={k.indo} disabled={keseharianKonfigurasiLoading || keseharianKonfigurasiSaving} />
                    <button class="btn-icon btn-icon-danger" onclick={() => hapusKriteriaKeseharian(i)} title="Hapus kriteria ini" disabled={keseharianKonfigurasiLoading || keseharianKonfigurasiSaving}>
                      <Icon name="trash" size={15} />
                    </button>
                  </div>
                {/each}
              </div>
            {/if}

            <div style="margin-top: 1rem; text-align: right;">
              <button
                class="btn-primary"
                onclick={simpanKonfigurasiNilaiKeseharian}
                disabled={keseharianKonfigurasiLoading || keseharianKonfigurasiSaving}
              >
                {keseharianKonfigurasiSaving ? 'Menyimpan...' : 'Simpan Kriteria'}
              </button>
            </div>
          </div>
        </section>
      {/if}`;

content = content.replace(oldUi, newUi);
fs.writeFileSync(file, content);
console.log('Fixed guru konfigurasi keseharian');
