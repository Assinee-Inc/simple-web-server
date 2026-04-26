package model

import (
	"fmt"
	"strings"

	"github.com/anglesson/simple-web-server/pkg/utils"
	"gorm.io/gorm"
)

type Client struct {
	gorm.Model
	PublicID  string `json:"public_id" gorm:"type:varchar(40);uniqueIndex"`
	Name      string `json:"name"`
	CPF       string `gorm:"unique" json:"cpf"`
	Birthdate string `json:"birthdate"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Validated bool   `json:"validated"`
	Purchases []*Purchase
}

func (c *Client) BeforeCreate(tx *gorm.DB) error {
	if c.PublicID == "" {
		c.PublicID = utils.GeneratePublicID("cli_")
	}
	return nil
}

func NewClient(name, cpf, birthDate, email, phone string) *Client {
	return &Client{
		Name:      name,
		CPF:       cpf,
		Birthdate: birthDate,
		Email:     email,
		Phone:     phone,
		Validated: false,
	}
}

func (c *Client) Update(name, cpf, email, phone string) {
	c.Name = name
	c.CPF = cpf
	c.Email = email
	c.Phone = phone
}

func (c *Client) TotalPurchasesByEbook(ebookID uint) int {
	var count int
	for _, purchase := range c.Purchases {
		if purchase.EbookID == ebookID {
			count++
		}
	}
	return count
}

func (c *Client) TotalDownloadsByEbook(ebookID uint) int {
	var count int
	for _, purchase := range c.Purchases {
		if purchase.EbookID == ebookID {
			count = +purchase.DownloadsUsed
		}
	}
	return count
}

func (c *Client) GetBirthdateBR() string {
	partsDate := strings.Split(c.Birthdate, "-")
	return fmt.Sprintf("%s/%s/%s", partsDate[2], partsDate[1], partsDate[0])
}

func (c *Client) GetInitials() string {
	names := strings.Fields(c.Name)
	if len(names) == 0 {
		return "?"
	}

	initials := ""
	if len(names) >= 1 {
		initials += string(names[0][0])
	}
	if len(names) >= 2 {
		initials += string(names[len(names)-1][0])
	}

	return strings.ToUpper(initials)
}
