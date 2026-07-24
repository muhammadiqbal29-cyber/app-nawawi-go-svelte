const fs = require('fs');
const path = require('path');

const file = path.join(__dirname, 'frontend/src/routes/guru/konfigurasi/KeseharianKriteriaTab.svelte');
let content = fs.readFileSync(file, 'utf8');

content = content.replace("let catatanList = [];", "let catatanList = $state([]);");
content = content.replace("let loading = true;", "let loading = $state(true);");
content = content.replace("let error = '';", "let error = $state('');");
content = content.replace("let showModal = false;", "let showModal = $state(false);");
content = content.replace("let isEditing = false;", "let isEditing = $state(false);");
content = content.replace("let editId = null;", "let editId = $state(null);");
content = content.replace("let saving = false;", "let saving = $state(false);");
content = content.replace("let formData = {", "let formData = $state({");
content = content.replace("let sekolahList = [];", "let sekolahList = $state([]);"); // not used anyway, but just in case

fs.writeFileSync(file, content);
console.log('Fixed reactivity.');
