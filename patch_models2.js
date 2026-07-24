const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'backend/internal/models/models.go');
let content = fs.readFileSync(file, 'utf8');

content = content.replace('Predikat    string    `gorm:"type:varchar(20);not null" json:"predikat"`', 'Predikat    string    `gorm:"type:varchar(20);not null" json:"predikat"`\n    BatasBawah  int       `gorm:"not null;default:0" json:"batas_bawah"`\n    BatasAtas   int       `gorm:"not null;default:0" json:"batas_atas"`');

fs.writeFileSync(file, content);
console.log('Model updated.');
