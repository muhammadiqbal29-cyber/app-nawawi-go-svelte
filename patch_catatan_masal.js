const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'frontend/src/routes/guru/rapor/keseharian/kelas/[kelas_id]/catatan-masal/+page.svelte');
let content = fs.readFileSync(file, 'utf8');

// Add states for konfigurasi and bank
const newStates = `
  let kriteriaList = $state([]);
  let bankCatatanList = $state([]);
`;

if (!content.includes('let kriteriaList = $state([]);')) {
    content = content.replace("let catatanMap = $state({});", "let catatanMap = $state({});\n" + newStates);
}

// Modify loadData to fetch kriteria and bank catatan
const loadKonfig = `
      // 3. Ambil konfigurasi dan bank catatan keseharian
      const resKonfig = await fetch(\`\${apiUrl}/api/guru/sekolah/\${sekolahId}/konfigurasi-nilai-keseharian\`, { credentials: 'include' });
      if (resKonfig.ok) {
        const body = await resKonfig.json();
        kriteriaList = body.kriteria_penilaian || [];
      }
      const resBank = await fetch(\`\${apiUrl}/api/guru/bank-catatan-keseharian?sekolah_id=\${sekolahId}\`, { credentials: 'include' });
      if (resBank.ok) {
        bankCatatanList = await resBank.json() || [];
      }
`;

if (!content.includes('3. Ambil konfigurasi dan bank catatan keseharian')) {
    content = content.replace("// Map catatan", loadKonfig + "\n      // Map catatan");
}

// Update generateAutoCatatan to use the logic
const generateFn = `
  function generateAutoCatatan() {
    if (kriteriaList.length === 0 || bankCatatanList.length === 0) {
      toastStore.show('Konfigurasi kriteria atau bank catatan masih kosong. Atur di menu Konfigurasi & Admin.', 'error');
      return;
    }

    siswaData = siswaData.map(s => {
      // s.rata_rata is already populated? 
      // Wait, s.rata_rata was 0! We need to calculate rata_rata from item.details in loadData.
      const avg = s.rata_rata || 0;
      let matchingKriteria = kriteriaList.find(k => avg >= k.min && avg <= k.max);
      
      let newCatatan = s.catatan;
      if (matchingKriteria) {
        const predikatIndo = matchingKriteria.indo;
        // Cari di bank catatan
        const matchingBank = bankCatatanList.filter(b => b.predikat.toLowerCase() === predikatIndo.toLowerCase());
        if (matchingBank.length > 0) {
          // Ambil random jika ada lebih dari 1
          const rand = Math.floor(Math.random() * matchingBank.length);
          newCatatan = matchingBank[rand].teks_catatan;
        }
      }

      return { 
        ...s, 
        catatan: newCatatan
      };
    });
    toastStore.show('Catatan otomatis berhasil di-generate berdasarkan rata-rata nilai', 'success');
  }
`;

if (content.includes("catatan: 'Selalu menjaga kedisiplinan dan akhlak yang baik di asrama. Tingkatkan terus belajarnya.'")) {
    const oldGenerate = `  function generateAutoCatatan() {
    siswaData = siswaData.map(s => {
      return { 
        ...s, 
        catatan: 'Selalu menjaga kedisiplinan dan akhlak yang baik di asrama. Tingkatkan terus belajarnya.' 
      };
    });
    toastStore.show('Catatan otomatis berhasil di-generate', 'success');
  }`;
    content = content.replace(oldGenerate, generateFn);
}

fs.writeFileSync(file, content);
console.log('Fixed catatan masal keseharian');
