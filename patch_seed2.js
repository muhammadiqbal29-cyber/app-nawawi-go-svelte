const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'backend/internal/database/seed_bank.go');
let content = fs.readFileSync(file, 'utf8');

content = content.replace('{SekolahID: 1, Predikat: "Sangat Baik", TeksCatatan: "Alhamdulillah, Ananda memiliki kedisiplinan dan akhlak yang sangat baik selama di pondok. Terus pertahankan prestasi dan kelakuan baiknya!"},', '{SekolahID: 1, Predikat: "Sangat Baik", BatasBawah: 0, BatasAtas: 0, TeksCatatan: "Alhamdulillah, Ananda memiliki kedisiplinan dan akhlak yang sangat baik selama di pondok. Terus pertahankan prestasi dan kelakuan baiknya!"},');
content = content.replace('{SekolahID: 1, Predikat: "Baik", TeksCatatan: "Ananda menunjukkan sikap yang baik dan kooperatif, meskipun sesekali masih perlu diingatkan. Tingkatkan terus kedisiplinannya!"},', '{SekolahID: 1, Predikat: "Baik", BatasBawah: 1, BatasAtas: 20, TeksCatatan: "Ananda menunjukkan sikap yang baik dan kooperatif, meskipun sesekali masih perlu diingatkan. Tingkatkan terus kedisiplinannya!"},');
content = content.replace('{SekolahID: 1, Predikat: "Cukup", TeksCatatan: "Ananda cukup baik, namun perlu lebih memperhatikan aturan pondok. Mohon kerja sama orang tua untuk memotivasi Ananda agar lebih tertib."},', '{SekolahID: 1, Predikat: "Cukup", BatasBawah: 21, BatasAtas: 50, TeksCatatan: "Ananda cukup baik, namun perlu lebih memperhatikan aturan pondok. Mohon kerja sama orang tua untuk memotivasi Ananda agar lebih tertib."},');
content = content.replace('{SekolahID: 1, Predikat: "Kurang", TeksCatatan: "Ananda memerlukan perhatian dan bimbingan yang lebih intensif dari orang tua terkait kepatuhan terhadap tata tertib pondok pesantren."},', '{SekolahID: 1, Predikat: "Kurang", BatasBawah: 51, BatasAtas: 9999, TeksCatatan: "Ananda memerlukan perhatian dan bimbingan yang lebih intensif dari orang tua terkait kepatuhan terhadap tata tertib pondok pesantren."},');

fs.writeFileSync(file, content);
console.log('Seed updated.');
