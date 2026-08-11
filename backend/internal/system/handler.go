package system

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	
	"gorm.io/gorm/clause"
)

var (
	// In-memory cache for maintenance mode to avoid querying DB on every request
	IsMaintenanceMode       bool
	MaintenanceMessage      string
	MaintenanceAllowedGurus string
	cacheMutex              sync.RWMutex
	cacheInitialized        bool
)

// InitSystemConfig loads the system configuration from the database into memory
func InitSystemConfig() {
	var confMode models.SystemConfig
	if err := database.DB.Where("`key` = ?", "maintenance_mode").First(&confMode).Error; err == nil {
		cacheMutex.Lock()
		IsMaintenanceMode = confMode.Value == "true"
		cacheMutex.Unlock()
	}

	var confMsg models.SystemConfig
	if err := database.DB.Where("`key` = ?", "maintenance_message").First(&confMsg).Error; err == nil {
		cacheMutex.Lock()
		MaintenanceMessage = confMsg.Value
		cacheMutex.Unlock()
	}

	var confGurus models.SystemConfig
	if err := database.DB.Where("`key` = ?", "maintenance_allowed_gurus").First(&confGurus).Error; err == nil {
		cacheMutex.Lock()
		MaintenanceAllowedGurus = confGurus.Value
		cacheMutex.Unlock()
	}

	cacheMutex.Lock()
	cacheInitialized = true
	cacheMutex.Unlock()
}

// GetMaintenanceStatus returns the current maintenance status from memory
func GetMaintenanceStatus() (bool, string, string) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	if !cacheInitialized {
		// Fallback if not initialized (though it shouldn't happen)
		return false, "", ""
	}
	return IsMaintenanceMode, MaintenanceMessage, MaintenanceAllowedGurus
}

// GetSystemStatusHandler — public: get current system status (e.g. maintenance mode)
func GetSystemStatusHandler(w http.ResponseWriter, r *http.Request) {
	isMaint, msg, _ := GetMaintenanceStatus()
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"maintenance_mode":    isMaint,
		"maintenance_message": msg,
		"server_time":         time.Now().Format(time.RFC3339),
	})
}

type SystemConfigPayload struct {
	MaintenanceMode         bool   `json:"maintenance_mode"`
	MaintenanceMessage      string `json:"maintenance_message"`
	MaintenanceAllowedGurus string `json:"maintenance_allowed_gurus"`
}

// GetSystemConfigHandler — admin: get system configuration including sensitive data
func GetSystemConfigHandler(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value("role").(string)
	if !ok || (role != "admin" && role != "superadmin") {
		response.Error(w, http.StatusForbidden, "Tidak memiliki izin", nil)
		return
	}

	isMaint, msg, gurus := GetMaintenanceStatus()
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"maintenance_mode":          isMaint,
		"maintenance_message":       msg,
		"maintenance_allowed_gurus": gurus,
	})
}

// UpdateSystemConfigHandler — admin: update system configuration
func UpdateSystemConfigHandler(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value("role").(string)
	if !ok || (role != "admin" && role != "superadmin") {
		response.Error(w, http.StatusForbidden, "Tidak memiliki izin", nil)
		return
	}

	var payload SystemConfigPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	// Save MaintenanceMode
	modeStr := "false"
	if payload.MaintenanceMode {
		modeStr = "true"
	}
	confMode := models.SystemConfig{
		Key:   "maintenance_mode",
		Value: modeStr,
	}
	database.DB.Clauses(clause.OnConflict{UpdateAll: true}).Create(&confMode)

	// Save MaintenanceMessage
	confMsg := models.SystemConfig{
		Key:   "maintenance_message",
		Value: payload.MaintenanceMessage,
	}
	database.DB.Clauses(clause.OnConflict{UpdateAll: true}).Create(&confMsg)

	// Save MaintenanceAllowedGurus
	confGurus := models.SystemConfig{
		Key:   "maintenance_allowed_gurus",
		Value: payload.MaintenanceAllowedGurus,
	}
	database.DB.Clauses(clause.OnConflict{UpdateAll: true}).Create(&confGurus)

	// Update memory cache
	cacheMutex.Lock()
	IsMaintenanceMode = payload.MaintenanceMode
	MaintenanceMessage = payload.MaintenanceMessage
	MaintenanceAllowedGurus = payload.MaintenanceAllowedGurus
	cacheMutex.Unlock()

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"message":                   "Konfigurasi sistem berhasil diperbarui",
		"maintenance_mode":          payload.MaintenanceMode,
		"maintenance_message":       payload.MaintenanceMessage,
		"maintenance_allowed_gurus": payload.MaintenanceAllowedGurus,
	})
}
