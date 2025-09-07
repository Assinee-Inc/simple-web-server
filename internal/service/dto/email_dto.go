package dto

import "fmt"

// ResendDownloadLinkDTO contém os dados necessários para reenvio do link de download
type ResendDownloadLinkDTO struct {
	ClientName   string
	ClientEmail  string
	EbookTitle   string
	EbookFiles   []FileDTO
	DownloadLink string
	AppName      string
	ContactEmail string
}

// FileDTO representa um arquivo do ebook
type FileDTO struct {
	OriginalName string
	Size         string
}

// ValidateResendDownloadLinkDTO valida os dados do DTO
func (dto *ResendDownloadLinkDTO) Validate() error {
	if dto.ClientEmail == "" {
		return fmt.Errorf("email do cliente é obrigatório")
	}

	if dto.ClientName == "" {
		return fmt.Errorf("nome do cliente é obrigatório")
	}

	if dto.EbookTitle == "" {
		return fmt.Errorf("título do ebook é obrigatório")
	}

	if dto.DownloadLink == "" {
		return fmt.Errorf("link de download é obrigatório")
	}

	if len(dto.EbookFiles) == 0 {
		return fmt.Errorf("ebook deve ter pelo menos um arquivo")
	}

	return nil
}
