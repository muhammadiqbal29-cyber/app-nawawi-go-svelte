import fs from 'fs';
import path from 'path';

function fixFile(filePath) {
  let content = fs.readFileSync(filePath, 'utf8');

  if (filePath.includes('admin/guru/+page.svelte')) {
    content = content.replace(
      /\{#each guruList as guru, index\}/g,
      '{#each guruList as guru, index (guru.id || index)}'
    );
  }

  if (filePath.includes('guru/profil/+page.svelte')) {
    content = content.replace(
      /\{#each profileData.hakAkses as jenjang, index\}/g,
      '{#each profileData.hakAkses as jenjang, index (jenjang)}'
    );
  }

  if (filePath.includes('materi-tahfiz/+page.svelte')) {
    content = content.replace(
      /\{#each activeSurah.ayatInfo as ayat\}/g,
      '{#each activeSurah.ayatInfo as ayat (ayat.nomorAyat || ayat)}'
    );
  }

  if (filePath.includes('guru/+layout.svelte')) {
    content = content.replace(/catch \{/g, 'catch (e) { /* ignore */');
  }

  if (filePath.includes('guru/rapot/kelulusan/[siswaId]/+page@.svelte')) {
    content = content.replace(
      /\{#each riwayatUjian as ujian, i\}/g,
      '{#each riwayatUjian as ujian, i (ujian.id || i)}'
    );
  }

  if (filePath.includes('guru/rapot/ujian-tahfiz/[siswaId]/+page@.svelte')) {
    content = content.replace(
      /\{#each subMateriList as m, i\}/g,
      '{#each subMateriList as m, i (m.id || i)}'
    );
  }

  fs.writeFileSync(filePath, content, 'utf8');
}

const files = [
  'src/routes/(dashboard)/admin/guru/+page.svelte',
  'src/routes/(dashboard)/guru/profil/+page.svelte',
  'src/routes/(dashboard)/materi-tahfiz/+page.svelte',
  'src/routes/guru/+layout.svelte',
  'src/routes/guru/rapot/kelulusan/[siswaId]/+page@.svelte',
  'src/routes/guru/rapot/ujian-tahfiz/[siswaId]/+page@.svelte'
];

files.forEach((f) => fixFile(path.join(process.cwd(), f)));
