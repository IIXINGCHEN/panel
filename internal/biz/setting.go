package biz

import (
	"context"
	"time"

	"github.com/tnb-labs/panel/internal/http/request"
)

type SettingKey string

const (
	SettingKeyName              SettingKey = "name"
	SettingKeyVersion           SettingKey = "version"
	SettingKeyMonitor           SettingKey = "monitor"
	SettingKeyMonitorDays       SettingKey = "monitor_days"
	SettingKeyBackupPath        SettingKey = "backup_path"
	SettingKeyWebsitePath       SettingKey = "website_path"
	SettingKeyMySQLRootPassword SettingKey = "mysql_root_password"
	SettingKeyOfflineMode       SettingKey = "offline_mode"
	SettingKeyAutoUpdate        SettingKey = "auto_update"
)

type Setting struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Key       SettingKey `gorm:"not null;default:'';unique" json:"key"`
	Value     string     `gorm:"not null;default:''" json:"value"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type SettingRepo interface {
	Get(key SettingKey, defaultValue ...string) (string, error)
	GetBool(key SettingKey, defaultValue ...bool) (bool, error)
	Set(key SettingKey, value string) error
	Delete(key SettingKey) error
	GetPanelSetting(ctx context.Context) (*request.PanelSetting, error)
	UpdatePanelSetting(ctx context.Context, setting *request.PanelSetting) (bool, error)
}
