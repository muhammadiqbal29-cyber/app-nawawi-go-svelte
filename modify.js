const fs = require('fs');
const paths = [
  'tasmi', 'tasmi-v2', 'kelulusan', 'kelulusan-v2'
];

for (const type of paths) {
  const filePath = `frontend/src/routes/guru/rapor/${type}/kelas/[kelas_id]/catatan-masal/+page.svelte`;
  let content = fs.readFileSync(filePath, 'utf-8');

  // Replace fetching tahfiz-masal with fetching siswa
  content = content.replace(
    /const params = new URLSearchParams\(\{ rombel \}\);\s+const resRapor = await fetch\(\`\$\{apiUrl\}\/api\/guru\/kelas\/\$\{kelasId\}\/tahfiz-masal\?\$\{params\.toString\(\)\}\`, \{\s+credentials: 'include'\s+\}\);\s+if \(!resRapor\.ok\) throw new Error\('Gagal memuat data rapor tahfiz'\);\s+const raporData = await resRapor\.json\(\);/,
    `const params = new URLSearchParams({ rombel });
      const resSiswa = await fetch(\`\${apiUrl}/api/guru/kelas/\${kelasId}/siswa?\${params.toString()}\`, {
        credentials: 'include'
      });
      if (!resSiswa.ok) throw new Error('Gagal memuat data siswa');
      const dataSiswa = await resSiswa.json();
      const siswaList = dataSiswa.data || dataSiswa || [];`
  );

  // Replace jenis_rapor: 'tahfiz' with 'tasmi' or 'kelulusan'
  const jenisRapor = type.startsWith('tasmi') ? 'tasmi' : 'kelulusan';
  content = content.replace(/jenis_rapor: 'tahfiz',/, `jenis_rapor: '${jenisRapor}',`);

  // Replace siswaData mapping
  content = content.replace(
    /siswaData = \(raporData \|\| \[\]\)\.map\(item => \{\s+return \{\s+siswa_id: item\.siswa\.id,\s+nama: item\.siswa\.nama_lengkap,\s+nisn: item\.siswa\.nisn,\s+rata_rata: 0,\s+catatan: map\[item\.siswa\.id\] \|\| ''\s+\};\s+\}\)\.sort\(\(a, b\) => a\.nama\.localeCompare\(b\.nama\)\);/,
    `siswaData = (siswaList || []).map(item => {
        return {
          siswa_id: item.id,
          nama: item.nama_lengkap,
          nisn: item.nisn,
          rata_rata: 0,
          catatan: map[item.id] || ''
        };
      }).sort((a, b) => a.nama.localeCompare(b.nama));`
  );

  // Replace generateAutoCatatan
  const autoCatatanText = type.startsWith('tasmi') 
    ? "Pencapaian yang baik dalam ujian Tasmi'. Terus murajaah dan pertahankan hafalan." 
    : "Selamat atas kelulusannya. Semoga hafalan yang telah dicapai menjadi berkah.";
  content = content.replace(
    /cat = 'Pencapaian yang baik dalam hafalan Al-Qur\\'an\. Terus semangat dan jaga murajaahnya\.';/,
    `cat = "${autoCatatanText}";`
  );

  // Replace simpanSemua logic
  content = content.replace(
    /const promises = siswaData\.map\(s => \{\s+return fetch\(\`\$\{apiUrl\}\/api\/guru\/nilai-catatan\`, \{\s+method: 'POST',\s+credentials: 'include',\s+headers: \{ 'Content-Type': 'application\/json' \},\s+body: JSON\.stringify\(\{\s+sekolah_id: Number\(sekolahId\),\s+siswa_id: s\.siswa_id,\s+catatan: s\.catatan \|\| ''\s+\}\)\s+\}\)\.then\(res => \{\s+if \(!res\.ok\) throw new Error\(\`Gagal menyimpan untuk siswa ID \$\{s\.siswa_id\}\`\);\s+\}\);\s+\}\);\s+await Promise\.all\(promises\);/g,
    `const payload = {
        sekolah_id: Number(sekolahId),
        jenis_rapor: '${jenisRapor}',
        data: siswaData.map(s => ({
          siswa_id: s.siswa_id,
          catatan: s.catatan || ''
        }))
      };

      const res = await fetch(\`\${apiUrl}/api/guru/catatan-rapor\`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      if (!res.ok) throw new Error('Gagal menyimpan catatan masal');`
  );

  // Replace goto
  content = content.replace(/goto\('\/guru\/rapor\/ujian-tahfiz'\)/g, `goto('/guru/rapor/${type}')`);
  
  // Replace titles
  const titleText = type.startsWith('tasmi') ? "Tasmi'" : 'Kelulusan';
  content = content.replace(/Catatan Masal Tahfiz/g, `Catatan Masal ${titleText}`);
  content = content.replace(/Catatan Masal Ujian Tahfiz/g, `Catatan Masal ${type.startsWith('tasmi') ? "Ujian Tasmi'" : 'Kelulusan'}`);

  fs.writeFileSync(filePath, content, 'utf-8');
}
