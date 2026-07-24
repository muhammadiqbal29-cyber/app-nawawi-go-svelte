const fs = require('fs');
const paths = [
  'tasmi', 'tasmi-v2', 'kelulusan', 'kelulusan-v2'
];

for (const type of paths) {
  const filePath = `frontend/src/routes/guru/rapor/${type}/+page.svelte`;
  let content = fs.readFileSync(filePath, 'utf-8');

  // Insert the button HTML
  const targetHTML = `    <div class="selected-header">
      <div style="display: flex; gap: 0.5rem; align-items: center;">
        <button class="btn-back" onclick={resetKelas}>
          <Icon name="chevronLeft" size={16} /> Ganti Kelas
        </button>
        <div class="selected-kelas">
          <Icon name="book" size={16} />
          <strong>{kelasLabel(selectedKelas)}</strong>
          <span class="badge">{rombelLabel(selectedKelas.rombel)}</span>
        </div>
      </div>
    </div>`;

  const replaceHTML = `    <div class="selected-header">
      <div style="display: flex; gap: 0.5rem; align-items: center;">
        <button class="btn-back" onclick={resetKelas}>
          <Icon name="chevronLeft" size={16} /> Ganti Kelas
        </button>
        <div class="selected-kelas">
          <Icon name="book" size={16} />
          <strong>{kelasLabel(selectedKelas)}</strong>
          <span class="badge">{rombelLabel(selectedKelas.rombel)}</span>
        </div>
      </div>
      <div style="display: flex; gap: 0.5rem; align-items: center;">
        <a
          href={\`/guru/rapor/${type}/kelas/\${selectedKelas.id}/catatan-masal?rombel=\${selectedKelas.rombel || '__kosong__'}&sekolah_id=\${guruState.activeSekolahId}&kelas=\${encodeURIComponent(kelasLabel(selectedKelas))}\`}
          class="btn-catatan-masal"
        >
          <Icon name="edit" size={16} /> Isi Catatan Masal
        </a>
      </div>
    </div>`;

  content = content.replace(targetHTML, replaceHTML);

  // Insert the CSS class
  const cssTarget = `.btn-back:hover {
    color: #4338ca;
  }`;

  const cssReplace = `.btn-back:hover {
    color: #4338ca;
  }
  .btn-catatan-masal {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    background: #10b981;
    color: white;
    padding: 0.5rem 1rem;
    border-radius: 8px;
    font-weight: 600;
    font-size: 0.9rem;
    text-decoration: none;
    transition: all 0.2s;
  }
  .btn-catatan-masal:hover {
    background: #059669;
    transform: translateY(-1px);
    box-shadow: 0 4px 6px rgba(16, 185, 129, 0.2);
  }`;

  content = content.replace(cssTarget, cssReplace);

  fs.writeFileSync(filePath, content, 'utf-8');
}
