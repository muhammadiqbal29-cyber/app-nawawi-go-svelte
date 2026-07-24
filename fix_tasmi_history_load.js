const fs = require('fs');

const paths = ['tasmi', 'tasmi-v2'];

for (const p of paths) {
  const filePath = `frontend/src/routes/guru/rapor/${p}/[siswaId]/+page@.svelte`;
  let content = fs.readFileSync(filePath, 'utf-8');
  const funcName = p === 'tasmi' ? 'loadNilaiTasmi' : 'loadNilaiTasmiV2';

  // Find loadNilaiTasmi function
  const searchStr = `const params = new URLSearchParams({
        jenis_penilaian: 'semester',
        mata_pelajaran: 'Ujian Tasmi',
        tahap: String(tahap)
      });`;
  
  const replaceStr = `const params = new URLSearchParams({
        jenis_penilaian: 'semester',
        mata_pelajaran: 'Ujian Tasmi',
        tahap: String(tahap)
      });
      // Cari riwayat untuk tahap yang dipilih
      const riwayatMatched = riwayatList.find(r => r.tahap === parseInt(tahap));
      if (riwayatMatched) {
        params.append('tahun_ajaran_id', riwayatMatched.tahun_ajaran_id);
        params.append('semester_id', riwayatMatched.semester_id);
      }`;

  if (content.includes(searchStr)) {
    content = content.replace(searchStr, replaceStr);
  }

  fs.writeFileSync(filePath, content, 'utf-8');
}
