const fs = require('fs');
let content = fs.readFileSync('./frontend/src/routes/guru/konfigurasi/+page.svelte', 'utf8');

// Tahfiz sub-tabs
const tahfizRegex = /<!-- Sub-tabs for Tahfiz -->[\s\S]*?<\/div>\s*{#if tahfizActiveTab === 'materi'}/;
const tahfizNew = `<!-- Sub-tabs for Tahfiz -->
        <div class="card-tabs">
          <button
            class="card-tab-btn"
            class:active={tahfizActiveTab === 'materi'}
            onclick={() => (tahfizActiveTab = 'materi')}
          >
            Daftar Materi Hafalan
          </button>
          <button
            class="card-tab-btn"
            class:active={tahfizActiveTab === 'kelulusan'}
            onclick={() => (tahfizActiveTab = 'kelulusan')}
          >
            Konfigurasi Nilai Kelulusan
          </button>
          <button
            class="card-tab-btn"
            class:active={tahfizActiveTab === 'kriteria'}
            onclick={() => (tahfizActiveTab = 'kriteria')}
          >
            Kriteria Penilaian
          </button>
        </div>

        {#if tahfizActiveTab === 'materi'}`;

content = content.replace(tahfizRegex, tahfizNew);

// Diniyah sub-tabs
const diniyahRegex = /<!-- Sub-tabs for Diniyah -->[\s\S]*?<\/div>\s*{#if diniyahActiveTab === 'mapel'}/;
const diniyahNew = `<!-- Sub-tabs for Diniyah -->
        <div class="card-tabs">
          <button
            class="card-tab-btn"
            class:active={diniyahActiveTab === 'mapel'}
            onclick={() => (diniyahActiveTab = 'mapel')}
          >
            Daftar Mapel
          </button>
          <button
            class="card-tab-btn"
            class:active={diniyahActiveTab === 'kriteria'}
            onclick={() => (diniyahActiveTab = 'kriteria')}
          >
            Kriteria Penilaian
          </button>
        </div>

        {#if diniyahActiveTab === 'mapel'}`;

content = content.replace(diniyahRegex, diniyahNew);

// Keseharian sub-tabs
const keseharianRegex = /<div\s*class="tabs sub-tabs"[\s\S]*?<\/div>\s*{#if activeSubTab === 'keseharian'}/;
const keseharianNew = `<div class="card-tabs">
          <button
            class="card-tab-btn"
            class:active={activeSubTab === 'keseharian'}
            onclick={() => (activeSubTab = 'keseharian')}
          >
            Indikator Keseharian
          </button>
          <button
            class="card-tab-btn"
            class:active={activeSubTab === 'pelanggaran'}
            onclick={() => (activeSubTab = 'pelanggaran')}
          >
            Pelanggaran
          </button>
          <button
            class="card-tab-btn"
            class:active={activeSubTab === 'iqob'}
            onclick={() => (activeSubTab = 'iqob')}
          >
            Iqob
          </button>
          <button
            class="card-tab-btn"
            class:active={activeSubTab === 'kriteria'}
            onclick={() => (activeSubTab = 'kriteria')}
          >
            Kriteria Penilaian
          </button>
        </div>

        {#if activeSubTab === 'keseharian'}`;

content = content.replace(keseharianRegex, keseharianNew);

fs.writeFileSync('./frontend/src/routes/guru/konfigurasi/+page.svelte', content, 'utf8');
console.log("Replaced sub tabs");
