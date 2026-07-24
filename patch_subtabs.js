const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'frontend/src/routes/guru/konfigurasi/+page.svelte');
let content = fs.readFileSync(file, 'utf8');

const oldKeseharian = `      {:else if activeTab === 'keseharian' && hasKeseharian}
        <!-- ======= TAB KESEHARIAN ======= -->
        <div class="toolbar">`;

const newKeseharian = `      {:else if activeTab === 'keseharian' && hasKeseharian}
        <div class="tabs sub-tabs" style="margin-bottom: 1rem; border-bottom: 1px solid var(--color-border); padding-bottom: 0;">
          <button class="tab-btn" class:active={activeSubTab === 'keseharian'} onclick={() => activeSubTab = 'keseharian'}>Indikator Keseharian</button>
          <button class="tab-btn" class:active={activeSubTab === 'pelanggaran'} onclick={() => activeSubTab = 'pelanggaran'}>Pelanggaran</button>
          <button class="tab-btn" class:active={activeSubTab === 'iqob'} onclick={() => activeSubTab = 'iqob'}>Iqob</button>
          <button class="tab-btn" class:active={activeSubTab === 'kriteria'} onclick={() => activeSubTab = 'kriteria'}>Kriteria Penilaian</button>
        </div>
        {#if activeSubTab === 'keseharian'}
        <!-- ======= TAB KESEHARIAN ======= -->
        <div class="toolbar">`;

content = content.replace(oldKeseharian, newKeseharian);

// Now change `{:else if activeTab === 'pelanggaran' && hasKeseharian}` to `{:else if activeSubTab === 'pelanggaran'}`
const oldPelanggaran = `      {:else if activeTab === 'pelanggaran' && hasKeseharian}`;
const newPelanggaran = `      {/if}\n      {#if activeSubTab === 'pelanggaran'}`;
content = content.replace(oldPelanggaran, newPelanggaran);

const oldIqob = `      {:else if activeTab === 'iqob' && hasKeseharian}`;
const newIqob = `      {/if}\n      {#if activeSubTab === 'iqob'}`;
content = content.replace(oldIqob, newIqob);

const oldKop = `      {:else if activeTab === 'kop_logo'}`;
const newKop = `      {/if}\n      {:else if activeTab === 'kop_logo'}`;
content = content.replace(oldKop, newKop);

fs.writeFileSync(file, content);
console.log('SubTabs updated.');
