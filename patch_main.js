const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'backend/main.go');
let content = fs.readFileSync(file, 'utf8');

content = content.replace("guruRouter.HandleFunc(\"/bank-catatan-keseharian\", bankCatatanKeseharian.GetAllHandler).Methods(\"GET\")\n", "");

// Add guruRouter for konfigurasi keseharian
const addRoute = `
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/konfigurasi-nilai-keseharian", konfigurasiKeseharian.GetKonfigurasiNilaiKeseharianHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/konfigurasi-nilai-keseharian", konfigurasiKeseharian.UpsertKonfigurasiNilaiKeseharianHandler).Methods("POST")
`;

if (!content.includes('GetKonfigurasiNilaiKeseharianHandler')) {
    content = content.replace("guruRouter.HandleFunc(\"/sekolah/{sekolah_id}/konfigurasi-nilai-diniyah\", konfigurasiDiniyah.UpsertKonfigurasiNilaiDiniyahHandler).Methods(\"POST\")", "guruRouter.HandleFunc(\"/sekolah/{sekolah_id}/konfigurasi-nilai-diniyah\", konfigurasiDiniyah.UpsertKonfigurasiNilaiDiniyahHandler).Methods(\"POST\")\n" + addRoute);
}

// Add import for konfigurasiKeseharian
if (!content.includes('"siakad-nawawi-be/internal/konfigurasiKeseharian"')) {
    content = content.replace('"siakad-nawawi-be/internal/konfigurasiDiniyah"', '"siakad-nawawi-be/internal/konfigurasiDiniyah"\n\t"siakad-nawawi-be/internal/konfigurasiKeseharian"');
}

fs.writeFileSync(file, content);
console.log('Fixed main.go');
