const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'frontend/src/routes/(dashboard)/admin/bank-catatan/+page.svelte');
let content = fs.readFileSync(file, 'utf8');

if (!content.includes("import KeseharianTab from './KeseharianTab.svelte';")) {
    content = content.replace("import DiniyahTab from './DiniyahTab.svelte';", "import DiniyahTab from './DiniyahTab.svelte';\n  import KeseharianTab from './KeseharianTab.svelte';");
}

if (!content.includes("<KeseharianTab />")) {
    content = content.replace("{:else if activeTab === 'diniyah'}\n      <DiniyahTab />\n    {/if}", "{:else if activeTab === 'diniyah'}\n      <DiniyahTab />\n    {:else if activeTab === 'keseharian'}\n      <KeseharianTab />\n    {/if}");
}

fs.writeFileSync(file, content);
console.log('Fixed admin bank-catatan');
