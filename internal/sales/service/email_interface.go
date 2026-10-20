package service

import (
	salesmodel "github.com/Assinee-Inc/docffy/internal/sales/model"
	salesdto "github.com/Assinee-Inc/docffy/internal/sales/service/dto"
)

// IEmailService defines the email operations needed by the sales module
type IEmailService interface {
	SendLinkToDownload(purchases []*salesmodel.Purchase)
	ResendDownloadLink(dto *salesdto.ResendDownloadLinkDTO) error
}
