const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'frontend/src/routes/(dashboard)/admin/bank-catatan/KeseharianTab.svelte');
let content = fs.readFileSync(file, 'utf8');

// Add to default formData
content = content.replace("    predikat: 'Sangat Baik',\n    teks_catatan: ''\n  };", "    predikat: 'Sangat Baik',\n    batas_bawah: 0,\n    batas_atas: 0,\n    teks_catatan: ''\n  };");
content = content.replace("      predikat: 'Sangat Baik',\n      teks_catatan: ''\n    };\n    showModal = true;", "      predikat: 'Sangat Baik',\n      batas_bawah: 0,\n      batas_atas: 0,\n      teks_catatan: ''\n    };\n    showModal = true;");
content = content.replace("      predikat: catatan.predikat,\n      teks_catatan: catatan.teks_catatan\n    };\n    showModal = true;", "      predikat: catatan.predikat,\n      batas_bawah: catatan.batas_bawah,\n      batas_atas: catatan.batas_atas,\n      teks_catatan: catatan.teks_catatan\n    };\n    showModal = true;");

// Add columns to table
content = content.replace("              <th>Predikat</th>\n              <th>Teks Catatan</th>", "              <th>Predikat</th>\n              <th>Batas Bawah</th>\n              <th>Batas Atas</th>\n              <th>Teks Catatan</th>");
content = content.replace("                  <td>\n                    <span class=\"badge {c.predikat === 'Sangat Baik' ? 'badge-primary' : 'badge-neutral'}\">\n                      {c.predikat}\n                    </span>\n                  </td>\n                  <td>\n                    <div class=\"text-catatan\">\n                      {c.teks_catatan}\n                    </div>\n                  </td>", "                  <td>\n                    <span class=\"badge {c.predikat === 'Sangat Baik' ? 'badge-primary' : 'badge-neutral'}\">\n                      {c.predikat}\n                    </span>\n                  </td>\n                  <td>{c.batas_bawah}</td>\n                  <td>{c.batas_atas}</td>\n                  <td>\n                    <div class=\"text-catatan\">\n                      {c.teks_catatan}\n                    </div>\n                  </td>");

// Add inputs to modal
const inputs = `
          <div class="form-group">
            <label for="batas_bawah">Batas Bawah Poin <span class="required">*</span></label>
            <input type="number" id="batas_bawah" class="form-input" bind:value={formData.batas_bawah} required min="0" />
          </div>

          <div class="form-group">
            <label for="batas_atas">Batas Atas Poin <span class="required">*</span></label>
            <input type="number" id="batas_atas" class="form-input" bind:value={formData.batas_atas} required min="0" />
          </div>
`;
content = content.replace('          <div class="form-group">\n            <label for="teks_catatan">Teks Catatan', inputs + '          <div class="form-group">\n            <label for="teks_catatan">Teks Catatan');

fs.writeFileSync(file, content);
console.log('Svelte updated.');
