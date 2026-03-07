package service

import (
	salesdto "github.com/anglesson/simple-web-server/internal/sales/service/dto"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
)

// IEmailService defines the email operations needed by the sales module
type IEmailService interface {
	SendLinkToDownload(purchases []*salesmodel.Purchase)
	ResendDownloadLink(dto *salesdto.ResendDownloadLinkDTO) error
}
