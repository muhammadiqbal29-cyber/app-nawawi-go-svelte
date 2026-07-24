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

// Remove the "Tambah Catatan" button's header-actions style if it's conflicting, but it should be fine.
// Wait, we need to remove the top header padding (page-container) because it's a subtab now.
content = content.replace(/<style>[\s\S]*?\.page-container \{[\s\S]*?\}[\s\S]*?<\/style>/, "<style>\n  /* Scoped Styles for Component */\n</style>"); // I will just remove the page-container style if there is one. 

fs.writeFileSync(file, content);
console.log('Kriteria tab updated for Guru.');
