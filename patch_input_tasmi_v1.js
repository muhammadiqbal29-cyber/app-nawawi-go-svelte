const fs = require('fs');
const path = 'frontend/src/routes/(blank)/guru/input-nilai-tasmi/[siswaId]/+page.svelte';
if (!fs.existsSync(path)) {
    console.log("No V1 file found.");
    process.exit(0);
}
let content = fs.readFileSync(path, 'utf8');

const initRegex = /async function init\(\) \{[\s\S]*?\/\/ 3\. Fetch nilai yang sudah ada\s*await loadNilaiTasmi\(\);/g;
const match = initRegex.exec(content);
if (!match) {
    console.error("Could not find init() block to patch in V1");
    process.exit(0); // non fatal
}

const newInit = `
      // 3. Fetch riwayat to determine maxTahap
      try {
        const riwayatRes = await fetch(\`\${apiUrl}/api/guru/siswa/\${siswaId}/riwayat-tasmi\`, {
          credentials: 'include'
        });
        if (riwayatRes.ok) {
          const rData = await riwayatRes.json();
          const rList = rData.riwayat || [];
          if (rList.length > 0) {
            let maxTahap = 0;
            for (const r of rList) {
              if (r.tahap > maxTahap) maxTahap = r.tahap;
            }
            if (maxTahap < 5) {
                tahap = maxTahap + 1;
            } else {
                tahap = 5;
            }
          }
        }
      } catch (e) {
        console.error(e);
      }

      // 4. Fetch nilai yang sudah ada
      await loadNilaiTasmi();
`;

content = content.replace('// 3. Fetch nilai yang sudah ada\n      await loadNilaiTasmi();', newInit);
fs.writeFileSync(path, content, 'utf8');
console.log("Patched V1 successfully");
