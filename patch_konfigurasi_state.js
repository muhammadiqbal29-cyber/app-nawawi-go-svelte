const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'frontend/src/routes/guru/konfigurasi/+page.svelte');
let content = fs.readFileSync(file, 'utf8');

if (!content.includes("let activeSubTab = 'keseharian';")) {
    content = content.replace("let activeTab = 'tahfiz';", "let activeTab = 'tahfiz';\n  let activeSubTab = 'keseharian';");
    fs.writeFileSync(file, content);
    console.log('State updated.');
} else {
    console.log('State already updated.');
}
