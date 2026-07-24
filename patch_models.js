const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'backend/internal/models/models.go');
let content = fs.readFileSync(file, 'utf8');

// Remove BatasBawah and BatasAtas
content = content.replace("    BatasBawah  int       `gorm:\"not null;default:0\" json:\"batas_bawah\"`\n", "");
content = content.replace("    BatasAtas   int       `gorm:\"not null;default:0\" json:\"batas_atas\"`\n", "");

// Add KonfigurasiNilaiKeseharian if not exists
const newModel = `
type KonfigurasiNilaiKeseharian struct {
    ID                int       \`gorm:"primaryKey;autoIncrement" json:"id"\`
    SekolahID         int       \`gorm:"type:int(11);not null;index;uniqueIndex:idx_konfig_nilai_keseharian" json:"sekolah_id"\`
    KriteriaPenilaian string    \`gorm:"type:text;not null" json:"-"\`
    Aktif             bool      \`gorm:"type:tinyint(1);not null;default:1" json:"aktif"\`
    CreatedAt         time.Time \`gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"\`
    UpdatedAt         time.Time \`gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"\`
}
`;

if (!content.includes('KonfigurasiNilaiKeseharian')) {
    content = content.replace("type BankCatatanKeseharian struct {\n", newModel + "\ntype BankCatatanKeseharian struct {\n");
}

fs.writeFileSync(file, content);
console.log('Fixed models.go');
