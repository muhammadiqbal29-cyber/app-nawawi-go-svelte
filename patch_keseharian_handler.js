const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'backend/internal/konfigurasiKeseharian/handler.go');
let content = fs.readFileSync(file, 'utf8');

content = content.replace(/package konfigurasiDiniyah/g, "package konfigurasiKeseharian");
content = content.replace(/konfigurasiNilaiDiniyah/g, "konfigurasiNilaiKeseharian");
content = content.replace(/KonfigurasiNilaiDiniyah/g, "KonfigurasiNilaiKeseharian");
content = content.replace(/diniyah/g, "keseharian");
content = content.replace(/Diniyah/g, "Keseharian");

fs.writeFileSync(file, content);
console.log('Fixed keseharian handler');
