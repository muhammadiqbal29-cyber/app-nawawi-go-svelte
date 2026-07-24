const fs = require('fs');

const paths = ['tasmi', 'tasmi-v2'];

for (const p of paths) {
  const filePath = `frontend/src/routes/guru/rapor/${p}/[siswaId]/+page@.svelte`;
  let content = fs.readFileSync(filePath, 'utf-8');

  // Add riwayat data state
  if (!content.includes('let riwayatList = $state([]);')) {
    content = content.replace('let error = $state(\'\');', 'let error = $state(\'\');\n  let riwayatList = $state([]);');
  }

  // Update Promise.all
  const loadFunc = p === 'tasmi' ? 'loadNilaiTasmi' : 'loadNilaiTasmiV2';
  const fetchRegex = new RegExp(`await Promise\\.all\\(\\[loadSekolah\\(\\), ${loadFunc}\\(\\), loadKonfigurasiTasmi\\(\\)\\]\\);`);
  content = content.replace(fetchRegex, `await Promise.all([loadSekolah(), ${loadFunc}(), loadKonfigurasiTasmi(), loadRiwayat()]);`);

  // Add loadRiwayat function
  const loadRiwayatCode = `
  async function loadRiwayat() {
    try {
      const res = await fetch(\`\${apiUrl}/api/guru/siswa/\${siswaId}/riwayat-${p}\`, {
        credentials: 'include'
      });
      if (res.ok) {
        const data = await res.json();
        riwayatList = data.riwayat || [];
      }
    } catch (e) {
      console.error(e);
    }
  }
`;
  if (!content.includes('function loadRiwayat()')) {
    content = content.replace(`async function loadKonfigurasiTasmi()`, loadRiwayatCode + `\n  async function loadKonfigurasiTasmi()`);
  }

  // Add HTML for riwayat table above settings-box
  const riwayatHTML = `
<!-- Riwayat Tasmi (Persisten Lintas Semester) -->
{#if riwayatList.length > 0}
<div class="riwayat-box no-print" style="max-width: 210mm; margin: 0 auto 1.5rem; background: white; padding: 1.5rem; border-radius: 8px; box-shadow: 0 4px 12px rgba(0,0,0,0.05); border: 1px solid #e2e8f0;">
  <h3 style="margin-top: 0; color: #1e293b; display: flex; align-items: center; gap: 0.5rem;">
    <Icon name="award" size={20} /> Riwayat Kelulusan Tasmi' Siswa
  </h3>
  <table style="width: 100%; border-collapse: collapse; margin-top: 1rem; text-align: left; font-size: 0.9rem;">
    <thead>
      <tr>
        <th style="padding: 0.75rem; background: #f8fafc; border-bottom: 2px solid #e2e8f0;">Tahap</th>
        <th style="padding: 0.75rem; background: #f8fafc; border-bottom: 2px solid #e2e8f0;">Juz yang Diujikan</th>
        <th style="padding: 0.75rem; background: #f8fafc; border-bottom: 2px solid #e2e8f0;">Predikat</th>
        <th style="padding: 0.75rem; background: #f8fafc; border-bottom: 2px solid #e2e8f0;">Nilai Akhir</th>
        <th style="padding: 0.75rem; background: #f8fafc; border-bottom: 2px solid #e2e8f0;">Catatan</th>
      </tr>
    </thead>
    <tbody>
      {#each riwayatList as r}
        <tr style="border-bottom: 1px solid #e2e8f0;">
          <td style="padding: 0.75rem;">Tahap {r.tahap}</td>
          <td style="padding: 0.75rem; font-weight: 600;">{r.juz_yang_diujikan || '-'}</td>
          <td style="padding: 0.75rem; font-family: 'Aref Ruqaa', serif; font-size: 1.25rem;">{r.predikat || '-'}</td>
          <td style="padding: 0.75rem;">{r.nilai_akhir || '0'}</td>
          <td style="padding: 0.75rem; color: #64748b;">{r.catatan || '-'}</td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>
{/if}

`;

  const settingsBoxTarget = `<div\n  class="settings-box no-print"`;
  if (content.includes(settingsBoxTarget) && !content.includes('Riwayat Kelulusan Tasmi')) {
    content = content.replace(settingsBoxTarget, riwayatHTML + settingsBoxTarget);
  }

  fs.writeFileSync(filePath, content, 'utf-8');
}
