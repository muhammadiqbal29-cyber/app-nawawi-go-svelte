const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'frontend/src/routes/guru/konfigurasi/KeseharianKriteriaTab.svelte');
let content = fs.readFileSync(file, 'utf8');

content = content.replace(/\/api\/bank-catatan-keseharian/g, "/api/guru/bank-catatan-keseharian");

fs.writeFileSync(file, content);
console.log('Fixed API paths.');
