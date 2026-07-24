const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'frontend/src/routes/guru/rapor/keseharian/kelas/[kelas_id]/catatan-masal/+page.svelte');
let content = fs.readFileSync(file, 'utf8');

const calcAvg = `      siswaData = (raporData || []).map(item => {
        let totalNilai = 0;
        let count = 0;
        if (item.rekaps && item.rekaps.length > 0) {
          item.rekaps.forEach(r => {
            totalNilai += r.nilai_akhir;
            count++;
          });
        }
        let avg = count > 0 ? totalNilai / count : 0;
        return {
          siswa_id: item.siswa.id,
          nama: item.siswa.nama_lengkap,
          nisn: item.siswa.nisn,
          rata_rata: avg,
          catatan: map[item.siswa.id] || ''
        };
      }).sort((a, b) => a.nama.localeCompare(b.nama));`;

if (content.includes('rata_rata: 0,')) {
    content = content.replace(/      siswaData = \(raporData \|\| \[\]\)\.map\(item => \{[\s\S]*?\}\)\.sort\(\(a, b\) => a\.nama\.localeCompare\(b\.nama\)\);/, calcAvg);
}

fs.writeFileSync(file, content);
console.log('Fixed avg calculation in catatan masal');
