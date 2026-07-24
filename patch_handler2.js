const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'backend/internal/bankCatatanKeseharian/handler.go');
let content = fs.readFileSync(file, 'utf8');

content = content.replace(/var payload struct {\n                Predikat    string `json:"predikat"`\n                TeksCatatan string `json:"teks_catatan"`\n        }/g, 'var payload struct {\n                Predikat    string `json:"predikat"`\n                BatasBawah  int    `json:"batas_bawah"`\n                BatasAtas   int    `json:"batas_atas"`\n                TeksCatatan string `json:"teks_catatan"`\n        }');

content = content.replace(/Predikat:    payload\.Predikat,\n                TeksCatatan: payload\.TeksCatatan,/g, 'Predikat:    payload.Predikat,\n                BatasBawah:  payload.BatasBawah,\n                BatasAtas:   payload.BatasAtas,\n                TeksCatatan: payload.TeksCatatan,');

content = content.replace(/data\.Predikat = payload\.Predikat\n        data\.TeksCatatan = payload\.TeksCatatan/g, 'data.Predikat = payload.Predikat\n        data.BatasBawah = payload.BatasBawah\n        data.BatasAtas = payload.BatasAtas\n        data.TeksCatatan = payload.TeksCatatan');

fs.writeFileSync(file, content);
console.log('Handler updated.');
