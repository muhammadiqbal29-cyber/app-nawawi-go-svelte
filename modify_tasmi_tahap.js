const fs = require('fs');

const paths = ['tasmi', 'tasmi-v2'];

for (const p of paths) {
  const filePath = `frontend/src/routes/guru/rapor/${p}/[siswaId]/+page@.svelte`;
  let content = fs.readFileSync(filePath, 'utf-8');
  const loadFunc = p === 'tasmi' ? 'loadNilaiTasmi' : 'loadNilaiTasmiV2';

  // Find the onMount Promise.all
  const promiseAllStr = `await Promise.all([loadSekolah(), ${loadFunc}(), loadKonfigurasiTasmi(), loadRiwayat()]);`;
  
  if (content.includes(promiseAllStr)) {
    const newOnMount = `
      await Promise.all([loadSekolah(), loadKonfigurasiTasmi(), loadRiwayat()]);
      
      // Auto-increment Tahap Logic
      if (riwayatList.length > 0) {
        let maxTahap = 0;
        for (const r of riwayatList) {
          if (r.tahap > maxTahap) maxTahap = r.tahap;
        }
        
        // Try to load the maxTahap first to see if it belongs to the current semester
        tahap = maxTahap;
        await ${loadFunc}();
        
        // If it's empty (meaning the current semester has no record for maxTahap), 
        // it means maxTahap was achieved in a PAST semester. So we should increment!
        if (details.length === 0 && (!nilaiHeader || !nilaiHeader.id)) {
          tahap = maxTahap + 1;
          await ${loadFunc}(); // reload for the new tahap
        }
      } else {
        await ${loadFunc}();
      }
`;
    content = content.replace(promiseAllStr, newOnMount);
  }

  // Add the badge to UI near the Tahap dropdown
  const selectStr = `<select bind:value={tahap} onchange={handleTahapChange} class="form-input">`;
  const badgeHtml = `
      <div style="display: flex; justify-content: space-between; align-items: center;">
        Tahap Ujian:
        <span style="font-size: 0.7rem; background: #e0f2fe; color: #0284c7; padding: 2px 6px; border-radius: 4px; font-weight: bold;">Auto-set</span>
      </div>
      <select bind:value={tahap} onchange={handleTahapChange} class="form-input">`;
  
  if (content.includes(selectStr) && !content.includes('Auto-set')) {
    content = content.replace(`Tahap Ujian:\n      <select bind:value={tahap} onchange={handleTahapChange} class="form-input">`, badgeHtml);
  }

  fs.writeFileSync(filePath, content, 'utf-8');
}
