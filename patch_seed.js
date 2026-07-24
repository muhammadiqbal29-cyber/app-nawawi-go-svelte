const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'backend/internal/database/seed_bank.go');
let content = fs.readFileSync(file, 'utf8');

content = content.replace(/BatasBawah:\s*\d+,\s*BatasAtas:\s*\d+,\s*/g, "");

fs.writeFileSync(file, content);
console.log('Fixed seed_bank.go');
