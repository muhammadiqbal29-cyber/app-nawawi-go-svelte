const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'frontend/src/routes/guru/konfigurasi/KeseharianKriteriaTab.svelte');
let content = fs.readFileSync(file, 'utf8');

// Replace standard sekolahList fetch with guruState.activeSekolahId if we want, but it's simpler to just let it use it.
content = content.replace("import { invalidateAll } from '$app/navigation';", "import { invalidateAll } from '$app/navigation';\n  import { guruState } from '$lib/guruState.svelte.js';");

// Make it reactive to guruState.activeSekolahId
content = content.replace("  let selectedSekolahId = '';", "  let selectedSekolahId = $derived(guruState.activeSekolahId);");
content = content.replace("  onMount(async () => {\n    await loadSekolah();\n    if (selectedSekolahId) {\n      await loadData();\n    }\n  });", "  $effect(() => {\n    if (selectedSekolahId) {\n      loadData();\n    }\n  });");
content = content.replace("  async function handleSekolahChange() {\n    if (selectedSekolahId) {\n      await loadData();\n    }\n  }", "");
// Remove loadSekolah since we don't need it
content = content.replace(/  async function loadSekolah\(\) \{[\s\S]*?\} catch \(e\) \{\}\n  \}/, "");

// Remove the page-container padding and header for Subtab usage
content = content.replace(/  \.page-container \{\n    padding: 2rem;\n    max-width: 1200px;\n    margin: 0 auto;\n  \}/, "  .page-container { padding: 1rem 0; width: 100%; }");

fs.writeFileSync(file, content);
console.log('Kriteria tab safely updated for Guru.');
