#!/bin/bash

# Remove sekolahID check blocks from handlers
for file in backend/internal/bankCatatanDiniyah/handler.go backend/internal/bankCatatanTahfizV2/handler.go backend/internal/bankCatatanKeseharian/handler.go; do
    # Remove context checking block for sekolahID
    sed -i '/sekolahID, ok := r.Context().Value("sekolah_id").(int)/,/sekolahID, _ = strconv.Atoi(sekolahIDStr)/d' "$file"
    sed -i '/}/ { N; /^[[:space:]]*$/ { N; s/}\n[[:space:]]*\n/}\n/ } }' "$file"
    # Remove the extra closing brace that got left behind
    sed -i '/sekolahIDStr == "" {/{N;N;d}' "$file" # just in case
    # Actually, the block is:
    # sekolahID, ok := r.Context().Value("sekolah_id").(int)
    # if !ok {
    #     sekolahIDStr := r.URL.Query().Get("sekolah_id")
    #     if sekolahIDStr == "" {
    #         response.Error(w, http.StatusBadRequest, "sekolah_id diperlukan", nil)
    #         return
    #     }
    #     sekolahID, _ = strconv.Atoi(sekolahIDStr)
    # }
    
    # A cleaner way using awk:
    awk '
    /sekolahID, ok := r.Context\(\).Value\("sekolah_id"\)/ { skip = 1 }
    skip && /}/ {
        brace_count++
        if (brace_count == 2) {
            skip = 0
            brace_count = 0
            next
        }
    }
    !skip { print $0 }
    ' "$file" > "$file.tmp" && mv "$file.tmp" "$file"

    # Remove SekolahID: sekolahID, from CreateHandler
    sed -i '/SekolahID:[[:space:]]*sekolahID,/d' "$file"

    # Change Where("sekolah_id = ?", sekolahID) to Find
    sed -i 's/\.Where("sekolah_id = ?", sekolahID)\.Find/\.Find/g' "$file"

    # Change Where("id = ? AND sekolah_id = ?", id, sekolahID) to Where("id = ?", id)
    sed -i 's/\.Where("id = ? AND sekolah_id = ?", id, sekolahID)/\.Where("id = ?", id)/g' "$file"

done
