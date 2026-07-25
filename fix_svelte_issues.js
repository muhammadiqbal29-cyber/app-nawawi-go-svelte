const fs = require('fs');

// 1. Fix {#else} in tasmi/+page@.svelte
const tasmiFile = 'frontend/src/routes/guru/rapor/tasmi/[siswaId]/+page@.svelte';
let tasmiContent = fs.readFileSync(tasmiFile, 'utf8');
tasmiContent = tasmiContent.replace(/{#else}/g, '{:else}');
fs.writeFileSync(tasmiFile, tasmiContent);

// 2. Fix KeseharianKriteriaTab in guru/konfigurasi/+page.svelte
const konfigFile = 'frontend/src/routes/guru/konfigurasi/+page.svelte';
let konfigContent = fs.readFileSync(konfigFile, 'utf8');

konfigContent = konfigContent.replace("import KeseharianKriteriaTab from './KeseharianKriteriaTab.svelte';\n", "");

const loadFunction = `
  async function loadKeseharianKonfigurasi() {
    keseharianKonfigurasiLoading = true;
    keseharianKonfigurasiError = '';
    keseharianKonfigurasiSuccess = '';
    try {
      const res = await fetch(\`\${apiUrl}/api/guru/konfigurasi-keseharian\`, { credentials: 'include' });
      if (res.ok) {
        const data = await res.json();
        if (data.kriteria_penilaian) {
          try {
            keseharianKriteriaPenilaian = JSON.parse(data.kriteria_penilaian);
          } catch (e) {
            keseharianKriteriaPenilaian = [];
          }
        } else {
          keseharianKriteriaPenilaian = [];
        }
      } else {
        keseharianKonfigurasiError = 'Gagal memuat konfigurasi keseharian.';
      }
    } catch (err) {
      keseharianKonfigurasiError = err.message;
    } finally {
      keseharianKonfigurasiLoading = false;
    }
  }

  async function simpanKeseharianKonfigurasi() {
    keseharianKonfigurasiSaving = true;
    keseharianKonfigurasiError = '';
    keseharianKonfigurasiSuccess = '';

    const payload = {
      kriteria_penilaian: JSON.stringify(keseharianKriteriaPenilaian)
    };

    try {
      const res = await fetch(\`\${apiUrl}/api/guru/konfigurasi-keseharian\`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
        credentials: 'include'
      });
      if (res.ok) {
        toastStore.addToast('Sukses', 'Konfigurasi Keseharian berhasil disimpan', 'success');
        keseharianKonfigurasiSuccess = 'Berhasil disimpan!';
      } else {
        const d = await res.json();
        keseharianKonfigurasiError = d.error || 'Gagal menyimpan';
        toastStore.addToast('Error', keseharianKonfigurasiError, 'error');
      }
    } catch (err) {
      keseharianKonfigurasiError = err.message;
      toastStore.addToast('Error', err.message, 'error');
    } finally {
      keseharianKonfigurasiSaving = false;
    }
  }

  function tambahKeseharianKriteria() {
    keseharianKriteriaPenilaian = [...keseharianKriteriaPenilaian, { min: 0, max: 0, arab: '', indo: '' }];
  }

  function hapusKeseharianKriteria(index) {
    keseharianKriteriaPenilaian = keseharianKriteriaPenilaian.filter((_, i) => i !== index);
  }
`;

if (!konfigContent.includes('loadKeseharianKonfigurasi()')) {
  konfigContent = konfigContent.replace(
    'async function loadIndikator() {',
    loadFunction + '\n  async function loadIndikator() {'
  );
  // Add to loadAll
  konfigContent = konfigContent.replace(
    'await Promise.all([\n      loadDiniyahMapel(),\n      loadKriteria(),',
    'await Promise.all([\n      loadDiniyahMapel(),\n      loadKriteria(),\n      loadKeseharianKonfigurasi(),'
  );
}

const keseharianUI = `
        <section class="kelulusan-settings">
          <div class="settings-header">
            <div>
              <h2 class="settings-title">Kriteria Penilaian Keseharian</h2>
              <p class="settings-subtitle">Atur batas predikat (Misal: 90 - 100 = Sangat Baik)</p>
            </div>
          </div>

          {#if keseharianKonfigurasiError}
            <div class="alert-error" style="margin-bottom: 1rem;">{keseharianKonfigurasiError}</div>
          {/if}
          {#if keseharianKonfigurasiSuccess}
            <div class="alert-success" style="margin-bottom: 1rem;">{keseharianKonfigurasiSuccess}</div>
          {/if}

          <div class="settings-kriteria-block" style="margin-top: 1rem;">
            <div class="settings-kriteria-header">
              <div class="settings-class-label">Kriteria Penilaian (Skala 100)</div>
              <button
                class="btn-secondary btn-sm"
                onclick={tambahKeseharianKriteria}
                disabled={keseharianKonfigurasiLoading || keseharianKonfigurasiSaving}
              >
                + Tambah Kriteria
              </button>
            </div>

            {#if keseharianKriteriaPenilaian.length === 0}
              <div class="settings-note">Belum ada kriteria penilaian yang dikonfigurasi.</div>
            {:else}
              <div class="kriteria-list">
                <div class="kriteria-header-row">
                  <div>Batas Minimal</div>
                  <div>Batas Maksimal</div>
                  <div>Predikat Arab</div>
                  <div>Predikat Indonesia</div>
                  <div>Aksi</div>
                </div>
                {#each keseharianKriteriaPenilaian as k, i}
                  <div class="kriteria-row">
                    <input type="number" min="0" max="100" class="form-input" bind:value={k.min} disabled={keseharianKonfigurasiLoading || keseharianKonfigurasiSaving} />
                    <input type="number" min="0" max="100" class="form-input" bind:value={k.max} disabled={keseharianKonfigurasiLoading || keseharianKonfigurasiSaving} />
                    <input type="text" class="form-input arabic-text" dir="rtl" placeholder="مُمْتَاز" bind:value={k.arab} disabled={keseharianKonfigurasiLoading || keseharianKonfigurasiSaving} />
                    <input type="text" class="form-input" placeholder="Sangat Baik" bind:value={k.indo} disabled={keseharianKonfigurasiLoading || keseharianKonfigurasiSaving} />
                    <button class="btn-icon btn-icon-danger" onclick={() => hapusKeseharianKriteria(i)} title="Hapus kriteria ini" disabled={keseharianKonfigurasiLoading || keseharianKonfigurasiSaving}>
                      <Icon name="trash" size={15} />
                    </button>
                  </div>
                {/each}
              </div>
            {/if}

            <div style="display: flex; justify-content: flex-end; margin-top: 1.5rem;">
              <button
                class="btn-primary"
                onclick={simpanKeseharianKonfigurasi}
                disabled={keseharianKonfigurasiLoading || keseharianKonfigurasiSaving}
              >
                {keseharianKonfigurasiSaving ? 'Menyimpan...' : 'Simpan Kriteria'}
              </button>
            </div>
          </div>
        </section>
`;

konfigContent = konfigContent.replace(
  '{:else if activeSubTab === \'kriteria\'}\n        <KeseharianKriteriaTab />',
  "{:else if activeSubTab === 'kriteria'}\n" + keseharianUI
);

fs.writeFileSync(konfigFile, konfigContent);

console.log('Fixed svelte issues');
