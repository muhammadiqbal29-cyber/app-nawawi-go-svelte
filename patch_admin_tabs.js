const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'frontend/src/routes/(dashboard)/admin/bank-catatan/+page.svelte');
let content = fs.readFileSync(file, 'utf8');

const oldTabs = `      <button
        class="tab-btn"
        class:active={activeTab === 'diniyah'}
        onclick={() => (activeTab = 'diniyah')}
      >
        Diniyah
      </button>
      <button
        class="tab-btn"
        class:active={activeTab === 'keseharian'}
        onclick={() => (activeTab = 'keseharian')}
      >
        Keseharian
      </button>
    </div>`;
const newTabs = `      <button
        class="tab-btn"
        class:active={activeTab === 'diniyah'}
        onclick={() => (activeTab = 'diniyah')}
      >
        Diniyah
      </button>
    </div>`;
content = content.replace(oldTabs, newTabs);

const oldImports = `import AlquranTab from './AlquranTab.svelte';
  import DiniyahTab from './DiniyahTab.svelte';
  import KeseharianTab from './KeseharianTab.svelte';`;
const newImports = `import AlquranTab from './AlquranTab.svelte';
  import DiniyahTab from './DiniyahTab.svelte';`;
content = content.replace(oldImports, newImports);

const oldRender = `    {:else if activeTab === 'diniyah'}
      <DiniyahTab />
    {:else if activeTab === 'keseharian'}
      <KeseharianTab />
    {/if}`;
const newRender = `    {:else if activeTab === 'diniyah'}
      <DiniyahTab />
    {/if}`;
content = content.replace(oldRender, newRender);

fs.writeFileSync(file, content);
console.log('Admin tabs updated.');
