const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'frontend/src/routes/guru/konfigurasi/+page.svelte');
let content = fs.readFileSync(file, 'utf8');

const oldTabs = `      {#if hasKeseharian}
        <button
          class="tab-btn"
          class:active={activeTab === 'keseharian'}
          onclick={() => (activeTab = 'keseharian')}
        >
          Indikator Keseharian
        </button>
        <button
          class="tab-btn"
          class:active={activeTab === 'pelanggaran'}
          onclick={() => (activeTab = 'pelanggaran')}
        >
          Pelanggaran
        </button>
        <button
          class="tab-btn"
          class:active={activeTab === 'iqob'}
          onclick={() => (activeTab = 'iqob')}
        >
          Iqob
        </button>
      {/if}`;

const newTabs = `      {#if hasKeseharian}
        <button
          class="tab-btn"
          class:active={activeTab === 'keseharian'}
          onclick={() => (activeTab = 'keseharian')}
        >
          Keseharian
        </button>
      {/if}`;

content = content.replace(oldTabs, newTabs);
fs.writeFileSync(file, content);
console.log('Tabs updated.');
