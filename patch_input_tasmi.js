const fs = require('fs');
const path = 'frontend/src/routes/(blank)/guru/input-nilai-tasmi-v2/[siswaId]/+page.svelte';
let content = fs.readFileSync(path, 'utf8');

// We need to inject loadRiwayat in init()
const initRegex = /async function init\(\) \{[\s\S]*?\/\/ 3\. Fetch nilai yang sudah ada\s*await loadNilaiTasmiV2\(\);/g;
const match = initRegex.exec(content);
if (!match) {
    console.error("Could not find init() block to patch");
    process.exit(1);
}

const newInit = `
      // 3. Fetch riwayat to determine maxTahap
      try {
        const riwayatRes = await fetch(\`\${apiUrl}/api/guru/siswa/\${siswaId}/riwayat-tasmi-v2\`, {
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
      await loadNilaiTasmiV2();
`;

content = content.replace('// 3. Fetch nilai yang sudah ada\n      await loadNilaiTasmiV2();', newInit);

fs.writeFileSync(path, content, 'utf8');
console.log("Patched successfully");
