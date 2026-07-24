const fs = require('fs');

const paths = ['tasmi', 'tasmi-v2'];

for (const p of paths) {
  const filePath = `frontend/src/routes/guru/rapor/${p}/[siswaId]/+page@.svelte`;
  let content = fs.readFileSync(filePath, 'utf-8');
  
  // Replace the old flawed auto-increment logic
  const oldLogic = `// Auto-increment Tahap Logic
      if (riwayatList.length > 0) {
        let maxTahap = 0;
        for (const r of riwayatList) {
          if (r.tahap > maxTahap) maxTahap = r.tahap;
        }

        // Try to load the maxTahap first to see if it belongs to the current semester
        tahap = maxTahap;
        await loadNilaiTasmi();

        // If it's empty (meaning the current semester has no record for maxTahap),
        // it means maxTahap was achieved in a PAST semester. So we should increment!
        if (details.length === 0 && (!nilaiHeader || !nilaiHeader.id)) {
          tahap = maxTahap + 1;
          await loadNilaiTasmi(); // reload for the new tahap
        }
      } else {
        await loadNilaiTasmi();
      }`;

  const loadFunc = p === 'tasmi' ? 'loadNilaiTasmi' : 'loadNilaiTasmiV2';

  const newLogic = `// Auto-increment Tahap Logic
      if (riwayatList.length > 0) {
        let maxTahap = 0;
        for (const r of riwayatList) {
          if (r.tahap > maxTahap) maxTahap = r.tahap;
        }

        // Check if maxTahap exists in ACTIVE semester (without history appending)
        let isActive = false;
        try {
          const checkParams = new URLSearchParams({
            jenis_penilaian: 'semester',
            mata_pelajaran: 'Ujian Tasmi',
            tahap: String(maxTahap)
          });
          const checkRes = await fetch(\`\${apiUrl}/api/guru/siswa/\${siswaId}/nilai-tasmi?\${checkParams.toString()}\`, { credentials: 'include' });
          if (checkRes.ok) {
             const body = await checkRes.json();
             if ((body.details && body.details.length > 0) || (body.nilai && body.nilai.id)) {
               isActive = true;
             }
          }
        } catch (e) {}

        if (isActive) {
          tahap = maxTahap;
        } else {
          tahap = maxTahap + 1;
        }
        await ${loadFunc}();
      } else {
        await ${loadFunc}();
      }`;

  if (content.includes('// Try to load the maxTahap first')) {
     // I need to use regex or string replace accurately since I added loadNilaiTasmi / loadNilaiTasmiV2 depending on page
     let toReplace = content.substring(content.indexOf('// Auto-increment Tahap Logic'), content.indexOf('loading = false;') - 5);
     content = content.replace(toReplace, newLogic + '\n\n    ');
  }

  fs.writeFileSync(filePath, content, 'utf-8');
}
