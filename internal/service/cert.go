package service

import (
	"net/http"

	"github.com/go-rat/chix"

	"github.com/tnb-labs/panel/internal/biz"
	"github.com/tnb-labs/panel/internal/http/request"
	"github.com/tnb-labs/panel/pkg/acme"
	"github.com/tnb-labs/panel/pkg/types"
)

type CertService struct {
	certRepo biz.CertRepo
}

func NewCertService(cert biz.CertRepo) *CertService {
	return &CertService{
		certRepo: cert,
	}
}

func (s *CertService) CAProviders(w http.ResponseWriter, r *http.Request) {
	Success(w, []types.LV{
		{
			Label: "GoogleCN（推荐）",
			Value: "googlecn",
		},
		{
			Label: "Let's Encrypt",
			Value: "letsencrypt",
		},
		{
			Label: "ZeroSSL",
			Value: "zerossl",
		},
		{
			Label: "SSL.com",
			Value: "sslcom",
		},
		{
			Label: "Google",
			Value: "google",
		},
		{
			Label: "Buypass",
			Value: "buypass",
		},
	})

}

func (s *CertService) DNSProviders(w http.ResponseWriter, r *http.Request) {
	Success(w, []types.LV{
		{
			Label: "阿里云",
			Value: string(acme.AliYun),
		},
		{
			Label: "腾讯云",
			Value: string(acme.Tencent),
		},
		{
			Label: "华为云",
			Value: string(acme.Huawei),
		},
		{
			Label: "CloudFlare",
			Value: string(acme.CloudFlare),
		},
		{
			Label: "Godaddy",
			Value: string(acme.Godaddy),
		},
		{
			Label: "Gcore",
			Value: string(acme.Gcore),
		},
		{
			Label: "Porkbun",
			Value: string(acme.Porkbun),
		},
		{
			Label: "Namecheap",
			Value: string(acme.Namecheap),
		},
		{
			Label: "NameSilo",
			Value: string(acme.NameSilo),
		},
		{
			Label: "Name.com",
			Value: string(acme.Namecom),
		},
		{
			Label: "ClouDNS",
			Value: string(acme.ClouDNS),
		},
		{
			Label: "Duck DNS",
			Value: string(acme.DuckDNS),
		},
		{
			Label: "Hetzner",
			Value: string(acme.Hetzner),
		},
		{
			Label: "Linode",
			Value: string(acme.Linode),
		},
		{
			Label: "Vercel",
			Value: string(acme.Vercel),
		},
	})
}

func (s *CertService) Algorithms(w http.ResponseWriter, r *http.Request) {
	Success(w, []types.LV{
		{
			Label: "EC256",
			Value: string(acme.KeyEC256),
		},
		{
			Label: "EC384",
			Value: string(acme.KeyEC384),
		},
		{
			Label: "RSA2048",
			Value: string(acme.KeyRSA2048),
		},
		{
			Label: "RSA4096",
			Value: string(acme.KeyRSA4096),
		},
	})

}

func (s *CertService) List(w http.ResponseWriter, r *http.Request) {
	req, err := Bind[request.Paginate](r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	certs, total, err := s.certRepo.List(req.Page, req.Limit)
	if err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	Success(w, chix.M{
		"total": total,
		"items": certs,
	})
}

func (s *CertService) Upload(w http.ResponseWriter, r *http.Request) {
	req, err := Bind[request.CertUpload](r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	cert, err := s.certRepo.Upload(req)
	if err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	Success(w, cert)
}

func (s *CertService) Create(w http.ResponseWriter, r *http.Request) {
	req, err := Bind[request.CertCreate](r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	cert, err := s.certRepo.Create(req)
	if err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	Success(w, cert)
}

func (s *CertService) Update(w http.ResponseWriter, r *http.Request) {
	req, err := Bind[request.CertUpdate](r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	if err = s.certRepo.Update(req); err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	Success(w, nil)
}

func (s *CertService) Get(w http.ResponseWriter, r *http.Request) {
	req, err := Bind[request.ID](r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	cert, err := s.certRepo.Get(req.ID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	Success(w, cert)
}

func (s *CertService) Delete(w http.ResponseWriter, r *http.Request) {
	req, err := Bind[request.ID](r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	err = s.certRepo.Delete(req.ID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	Success(w, nil)
}

func (s *CertService) ObtainAuto(w http.ResponseWriter, r *http.Request) {
	req, err := Bind[request.ID](r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	if _, err = s.certRepo.ObtainAuto(req.ID); err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	Success(w, nil)
}

func (s *CertService) ObtainManual(w http.ResponseWriter, r *http.Request) {
	req, err := Bind[request.ID](r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	if _, err = s.certRepo.ObtainManual(req.ID); err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	Success(w, nil)
}

func (s *CertService) ObtainSelfSigned(w http.ResponseWriter, r *http.Request) {
	req, err := Bind[request.ID](r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	if err = s.certRepo.ObtainSelfSigned(req.ID); err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	Success(w, nil)
}

func (s *CertService) Renew(w http.ResponseWriter, r *http.Request) {
	req, err := Bind[request.ID](r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	_, err = s.certRepo.Renew(req.ID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	Success(w, nil)
}

func (s *CertService) ManualDNS(w http.ResponseWriter, r *http.Request) {
	req, err := Bind[request.ID](r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	dns, err := s.certRepo.ManualDNS(req.ID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	Success(w, dns)
}

func (s *CertService) Deploy(w http.ResponseWriter, r *http.Request) {
	req, err := Bind[request.CertDeploy](r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	err = s.certRepo.Deploy(req.ID, req.WebsiteID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	Success(w, nil)
}
