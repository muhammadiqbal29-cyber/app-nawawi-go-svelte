import os
import re

files = [
    "backend/internal/bankCatatanDiniyah/handler.go",
    "backend/internal/bankCatatanTahfizV2/handler.go",
    "backend/internal/bankCatatanKeseharian/handler.go"
]

block_to_remove = r'''	sekolahID, ok := r\.Context\(\)\.Value\("sekolah_id"\)\.\(int\)
	if !ok \{
		sekolahIDStr := r\.URL\.Query\(\)\.Get\("sekolah_id"\)
		if sekolahIDStr == "" \{
			response\.Error\(w, http\.(?:StatusBadRequest|StatusUnauthorized), ".*?", nil\)
			return
		\}
		sekolahID, _ = strconv\.Atoi\(sekolahIDStr\)
	\}'''

for f in files:
    with open(f, 'r') as file:
        content = file.read()
    
    # Remove context block
    content = re.sub(block_to_remove, "", content)
    
    # Remove SekolahID assignment in CreateHandler
    content = re.sub(r'SekolahID:\s*sekolahID,', '', content)
    
    # Replace Where("sekolah_id = ?", sekolahID).Find
    content = content.replace('.Where("sekolah_id = ?", sekolahID).Find', '.Find')
    
    # Replace Where("id = ? AND sekolah_id = ?", id, sekolahID)
    content = content.replace('.Where("id = ? AND sekolah_id = ?", id, sekolahID)', '.Where("id = ?", id)')

    with open(f, 'w') as file:
        file.write(content)
