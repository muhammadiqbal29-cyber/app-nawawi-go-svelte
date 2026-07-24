const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'backend/internal/database/db.go');
let content = fs.readFileSync(file, 'utf8');

if (!content.includes('models.KonfigurasiNilaiKeseharian{}')) {
    content = content.replace("models.BankCatatanKeseharian{},", "models.KonfigurasiNilaiKeseharian{},\n\t\tmodels.BankCatatanKeseharian{},");
}

fs.writeFileSync(file, content);
console.log('Fixed db.go');
