const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'frontend/src/routes/guru/konfigurasi/+page.svelte');
let content = fs.readFileSync(file, 'utf8');

// The file currently has:
//       {/if}
//       {#if activeSubTab === 'pelanggaran'}
// Let's replace it to:
//       {:else if activeSubTab === 'pelanggaran'}
content = content.replace("      {/if}\n      {#if activeSubTab === 'pelanggaran'}", "      {:else if activeSubTab === 'pelanggaran'}");

//       {/if}
//       {#if activeSubTab === 'iqob'}
// Replace it to:
//       {:else if activeSubTab === 'iqob'}
content = content.replace("      {/if}\n      {#if activeSubTab === 'iqob'}", "      {:else if activeSubTab === 'iqob'}");

// Then we have:
//       {/if}
//       {#if activeSubTab === 'kriteria'}
//         <KeseharianKriteriaTab />
//       {/if}
//       {:else if activeTab === 'kop_logo'}
// Let's replace it to:
//       {:else if activeSubTab === 'kriteria'}
//         <KeseharianKriteriaTab />
//       {/if}
//       {:else if activeTab === 'kop_logo'}
content = content.replace("      {/if}\n      {#if activeSubTab === 'kriteria'}\n        <KeseharianKriteriaTab />\n      {/if}\n      {:else if activeTab === 'kop_logo'}", "      {:else if activeSubTab === 'kriteria'}\n        <KeseharianKriteriaTab />\n      {/if}\n      {:else if activeTab === 'kop_logo'}");

fs.writeFileSync(file, content);
console.log('Fixed tabs block.');
