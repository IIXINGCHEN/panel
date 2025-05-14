package service

import (
	"net/http"

	"github.com/tnb-labs/panel/internal/biz"
	"github.com/tnb-labs/panel/internal/http/request"
	"github.com/tnb-labs/panel/pkg/tools"
)

type SettingService struct {
	settingRepo biz.SettingRepo
}

func NewSettingService(setting biz.SettingRepo) *SettingService {
	return &SettingService{
		settingRepo: setting,
	}
}

func (s *SettingService) Get(w http.ResponseWriter, r *http.Request) {
	setting, err := s.settingRepo.GetPanelSetting()
	if err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	Success(w, setting)
}

func (s *SettingService) Update(w http.ResponseWriter, r *http.Request) {
	req, err := Bind[request.PanelSetting](r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	restart := false
	if restart, err = s.settingRepo.UpdatePanelSetting(req); err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	if restart {
		tools.RestartPanel()
	}

	Success(w, nil)
}
