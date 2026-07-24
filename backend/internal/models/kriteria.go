package models

import "encoding/json"

type KriteriaPenilaianItem struct {
	Min  int    `json:"min"`
	Max  int    `json:"max"`
	Arab string `json:"arab"`
	Indo string `json:"indo"`
}

func GetDefaultKriteria() []KriteriaPenilaianItem {
	return []KriteriaPenilaianItem{
		{Min: 90, Max: 100, Arab: "مُمْتَاز", Indo: "Istimewa"},
		{Min: 80, Max: 89, Arab: "جَيِّد جِدًّا", Indo: "Sangat Baik"},
		{Min: 70, Max: 79, Arab: "جَيِّد", Indo: "Baik"},
		{Min: 60, Max: 69, Arab: "مَقْبُول", Indo: "Cukup"},
		{Min: 0, Max: 59, Arab: "رَاسِب", Indo: "Mengulang"},
	}
}

func EncodeKriteria(kriteria []KriteriaPenilaianItem) string {
	data, err := json.Marshal(kriteria)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func DecodeKriteria(raw string) []KriteriaPenilaianItem {
	if raw == "" || raw == "[]" {
		return GetDefaultKriteria()
	}
	var kriteria []KriteriaPenilaianItem
	if err := json.Unmarshal([]byte(raw), &kriteria); err != nil {
		return GetDefaultKriteria()
	}
	if len(kriteria) == 0 {
		return GetDefaultKriteria()
	}
	return kriteria
}

func GetPredikatNilai(nilaiAkhir int, kriteria []KriteriaPenilaianItem) string {
	if len(kriteria) == 0 {
		kriteria = GetDefaultKriteria()
	}
	for _, k := range kriteria {
		if nilaiAkhir >= k.Min && nilaiAkhir <= k.Max {
			return k.Arab
		}
	}
	return ""
}
