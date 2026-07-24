const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'frontend/src/routes/guru/konfigurasi/+page.svelte');
let content = fs.readFileSync(file, 'utf8');

const importStatement = "import Icon from '$lib/components/ui/Icon.svelte';\n  import KeseharianKriteriaTab from './KeseharianKriteriaTab.svelte';";
content = content.replace("import Icon from '$lib/components/ui/Icon.svelte';", importStatement);

const kriteriaBlock = `      {#if activeSubTab === 'kriteria'}
        <KeseharianKriteriaTab />
      {/if}`;

// Insert kriteriaBlock after iqob block
const oldIqobBlock = `      {/if}
      {#if activeSubTab === 'iqob'}
        <!-- ======= TAB IQOB ======= -->`;
// We will just find where the `kop_logo` starts and put it right before it
const oldKop = `      {/if}
      {:else if activeTab === 'kop_logo'}`;
const newKop = `      {/if}
      {#if activeSubTab === 'kriteria'}
        <KeseharianKriteriaTab />
      {/if}
      {:else if activeTab === 'kop_logo'}`;

content = content.replace(oldKop, newKop);

fs.writeFileSync(file, content);
console.log('Component imported and rendered.');
