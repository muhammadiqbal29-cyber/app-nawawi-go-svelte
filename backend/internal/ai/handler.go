package ai

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/tahunajaran"
)

// AnalitikSiswaDiniyahHandler memproses request HTTP untuk mendapatkan AI Analitik Diniyah Siswa
func AnalitikSiswaDiniyahHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siswaID, err := strconv.Atoi(vars["siswa_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Siswa tidak valid", err)
		return
	}

	forceRefresh := r.URL.Query().Get("force") == "true"

	taID, semID, err := tahunajaran.ResolveActivePeriod(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	sekolahIDStr, _ := vars["sekolah_id"]
	sekolahID, _ := strconv.Atoi(sekolahIDStr)

	// Panggil Service Diniyah AI
	aiHTML, err := GenerateAnalitikSiswaDiniyah(siswaID, sekolahID, taID, semID, forceRefresh)
	if err != nil {
		if err.Error() == "Sekolah tidak melanggan paket Diniyah" {
			response.Error(w, http.StatusForbidden, err.Error(), err)
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	// Return data html murni dalam json format yang siap disisipkan ke DOM
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"siswa_id": siswaID,
		"ai_html":  aiHTML,
	})
}

type MesoRequest struct {
	FilterJudul string `json:"filter_judul"`
	TotalSiswa  int    `json:"total_siswa"`
	SangatBaik  int    `json:"sangat_baik"`
	Baik        int    `json:"baik"`
	Cukup       int    `json:"cukup"`
	Risiko      int    `json:"risiko"`
	Remedial    int    `json:"remedial"`
}

// AnalitikKelasDiniyahHandler memproses request HTTP untuk mendapatkan AI Analitik Diniyah tingkat Kelas (Meso)
func AnalitikKelasDiniyahHandler(w http.ResponseWriter, r *http.Request) {
	var payload MesoRequest
	importJson := json.NewDecoder(r.Body)
	if err := importJson.Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Payload tidak valid", err)
		return
	}

	forceRefresh := r.URL.Query().Get("force") == "true"

	// Panggil Service Diniyah AI tingkat kelas
	aiHTML, err := GenerateAnalitikKelasDiniyah(payload, forceRefresh)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	// Return data json
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"ai_html":  aiHTML,
	})
}

