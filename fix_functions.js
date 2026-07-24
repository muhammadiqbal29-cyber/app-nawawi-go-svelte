const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'frontend/src/routes/guru/input-nilai/keseharian/kelas/[id]/+page.svelte');
let content = fs.readFileSync(file, 'utf8');

// Replace submitPelanggaran
const plgRegex = /function selectPelanggaran\(j\) \{[\s\S]*?async function submitPelanggaran\(\) \{[\s\S]*?\} finally \{\n\s*plgFormLoading = false;\n\s*\}\n\s*\}/;
const newPlg = `function selectPelanggaran(j) {
    plgFormJenisId = j.id;
    plgSelectedNama = j.nama;
    plgSearchQuery = '';
    plgShowDropdown = false;
    plgFormPoin = j.poin;
  }

  function tambahPendingPelanggaran() {
    if (!plgFormJenisId || !plgFormTanggal) {
      plgFormError = 'Pilih jenis pelanggaran dan tanggal terlebih dahulu';
      return;
    }
    plgFormError = '';
    pendingPelanggaranList = [...pendingPelanggaranList, {
      jenis_pelanggaran_id: plgFormJenisId,
      nama: plgSelectedNama,
      tanggal: plgFormTanggal,
      poin: plgFormPoin,
      iqob: plgFormIqob
    }];
    plgFormJenisId = '';
    plgFormPoin = 0;
    plgSearchQuery = '';
    plgShowDropdown = true;
    plgSelectedNama = '';
  }

  function removePendingPelanggaran(index) {
    pendingPelanggaranList = pendingPelanggaranList.filter((_, i) => i !== index);
  }

  async function submitPelanggaran() {
    if (pendingPelanggaranList.length === 0) {
      plgFormError = 'Tambahkan setidaknya satu pelanggaran ke daftar';
      return;
    }
    plgFormLoading = true;
    plgFormError = '';

    try {
      const url = new URL(\`\${apiUrl}/api/guru/pelanggaran/catatan\`);
      if (guruState.activeTahunId) url.searchParams.append('tahun_ajaran_id', guruState.activeTahunId);
      if (guruState.activeSemesterId) url.searchParams.append('semester_id', guruState.activeSemesterId);

      const promises = pendingPelanggaranList.map(item => {
        return fetch(url.toString(), {
          method: 'POST',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            siswa_id: Number(plgFormSiswaId),
            sekolah_id: Number(sekolahId),
            jenis_pelanggaran_id: Number(item.jenis_pelanggaran_id),
            tanggal: item.tanggal,
            poin: Number(item.poin),
            iqob: item.iqob || ''
          })
        }).then(async r => {
          if (!r.ok) {
            const b = await r.json().catch(() => ({}));
            throw new Error(b.message || \`Gagal menambah pelanggaran: \${item.nama}\`);
          }
        });
      });

      await Promise.all(promises);

      toastStore.show('Pelanggaran berhasil dicatat', 'success');
      pendingPelanggaranList = [];
      await loadPelanggaranKelas();
    } catch (err) {
      plgFormError = err.message;
    } finally {
      plgFormLoading = false;
    }
  }`;

content = content.replace(plgRegex, newPlg);

// Add applyRecommendedIqob
const addIqob = `function tambahPendingIqob() {`;
const replaceIqob = `function applyRecommendedIqob() {
    if (recommendedIqob) {
      iqobFormJenisId = recommendedIqob.id;
      iqobFormPoin = recommendedIqob.poin;
      iqobSelectedNama = recommendedIqob.nama;
      iqobShowDropdown = false;
      iqobSearchQuery = '';
    }
  }

  function tambahPendingIqob() {`;

content = content.replace(addIqob, replaceIqob);

fs.writeFileSync(file, content);
console.log('Fixed Functions');
