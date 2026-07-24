const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'frontend/src/routes/guru/konfigurasi/KeseharianKriteriaTab.svelte');
let content = fs.readFileSync(file, 'utf8');

// Replace the closing brace of formData with })
content = content.replace("    teks_catatan: ''\n  };", "    teks_catatan: ''\n  });");

fs.writeFileSync(file, content);
console.log('Fixed syntax error.');
