const fs = require('fs');

const paths = ['tasmi', 'tasmi-v2'];

for (const p of paths) {
  const filePath = `frontend/src/routes/guru/rapor/${p}/+page.svelte`;
  let content = fs.readFileSync(filePath, 'utf-8');

  // Add riwayatTasmiMap state
  if (!content.includes('let riwayatTasmiMap = $state({});')) {
    content = content.replace('let siswaList = $state([]);', 'let siswaList = $state([]);\n  let riwayatTasmiMap = $state({});');
  }

  // Update loadSiswa to also fetch rekap-tasmi
  const fetchSiswaRegex = /const data = await res\.json\(\);\s+siswaList = data\.data \|\| data \|\| \[\];/;
  const newFetchSiswa = `const data = await res.json();
      siswaList = data.data || data || [];

      // Fetch riwayat tasmi
      const resRekap = await fetch(\`\${apiUrl}/api/guru/kelas/\${kelas.id}/rekap-${p}?rombel=\${kelas.rombel || '__kosong__'}\`, {
        credentials: 'include'
      });
      if (resRekap.ok) {
        const rekapData = await resRekap.json();
        const rekapList = rekapData.rekap || [];
        const map = {};
        for (const r of rekapList) {
          if (!map[r.siswa_id] || r.id > map[r.siswa_id].id) {
             map[r.siswa_id] = r;
          }
        }
        riwayatTasmiMap = map;
      }`;
  content = content.replace(fetchSiswaRegex, newFetchSiswa);

  // Update HTML to show badge
  const htmlBadgeTarget = `<span class="siswa-nisn">{siswa.nisn || '-'}</span>
              </div>`;
  const htmlBadgeReplace = `<span class="siswa-nisn">{siswa.nisn || '-'}</span>
              </div>
              <div class="siswa-badge-container">
                {#if riwayatTasmiMap[siswa.id]}
                  <span class="badge-riwayat success">Tasmi Lulus (Tahap {riwayatTasmiMap[siswa.id].tahap})</span>
                {:else}
                  <span class="badge-riwayat neutral">Belum Tasmi</span>
                {/if}
              </div>`;
  if (content.includes(htmlBadgeTarget)) {
    content = content.replace(htmlBadgeTarget, htmlBadgeReplace);
  }

  // Add CSS for badge
  if (!content.includes('.badge-riwayat')) {
    content = content.replace('</style>', `
  .siswa-badge-container {
    margin-left: auto;
  }
  .badge-riwayat {
    font-size: 0.75rem;
    font-weight: 700;
    padding: 0.25rem 0.6rem;
    border-radius: 9999px;
  }
  .badge-riwayat.success {
    background: #dcfce7;
    color: #166534;
  }
  .badge-riwayat.neutral {
    background: #f1f5f9;
    color: #64748b;
  }
</style>`);
  }

  fs.writeFileSync(filePath, content, 'utf-8');
}
