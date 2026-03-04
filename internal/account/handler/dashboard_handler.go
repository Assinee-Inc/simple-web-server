package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	authmw "github.com/anglesson/simple-web-server/internal/auth/handler/middleware"
	accountrepo "github.com/anglesson/simple-web-server/internal/account/repository"
	"github.com/anglesson/simple-web-server/pkg/template"
)

// TopClientWithInitials extends TopClient with initials
type TopClientWithInitials struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	TotalPurchases int64  `json:"total_purchases"`
	Initials       string `json:"initials"`
}

// getInitials generates initials from a name
func getInitials(name string) string {
	names := strings.Fields(name)
	if len(names) == 0 {
		return ""
	}
	if len(names) == 1 {
		return strings.ToUpper(string(names[0][0]))
	}
	return strings.ToUpper(string(names[0][0]) + string(names[len(names)-1][0]))
}

type DashboardHandler struct {
	templateRenderer template.TemplateRenderer
}

func NewDashboardHandler(templateRenderer template.TemplateRenderer) *DashboardHandler {
	return &DashboardHandler{
		templateRenderer: templateRenderer,
	}
}

func (h *DashboardHandler) DashboardView(w http.ResponseWriter, r *http.Request) {
	loggedUser := authmw.Auth(r)
	dashRepository := accountrepo.NewDashboardRepository(loggedUser.ID)

	totalEbooks := dashRepository.GetTotalEbooks()
	totalSendEbooks := dashRepository.GetTotalSendEbooks()
	totalClients := dashRepository.GetTotalClients()
	ebookStats, _ := dashRepository.GetEbookStats()

	dailyPurchases, _ := dashRepository.GetDailyPurchases()
	dailyDownloads, _ := dashRepository.GetDailyDownloads()
	topEbooks, _ := dashRepository.GetTopEbooks()
	topClients, _ := dashRepository.GetTopClients()
	topDownloadedEbooks, _ := dashRepository.GetTopDownloadedEbooks()

	var topClientsWithInitials []TopClientWithInitials
	for _, client := range topClients {
		topClientsWithInitials = append(topClientsWithInitials, TopClientWithInitials{
			Name:           client.Name,
			Email:          client.Email,
			TotalPurchases: client.TotalPurchases,
			Initials:       getInitials(client.Name),
		})
	}

	dailyPurchasesJSON, err := json.Marshal(dailyPurchases)
	if err != nil {
		log.Printf("Error marshaling daily purchases data to JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	dailyDownloadsJSON, err := json.Marshal(dailyDownloads)
	if err != nil {
		log.Printf("Error marshaling daily downloads data to JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	topEbooksJSON, err := json.Marshal(topEbooks)
	if err != nil {
		log.Printf("Error marshaling top ebooks data to JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	topDownloadedEbooksJSON, err := json.Marshal(topDownloadedEbooks)
	if err != nil {
		log.Printf("Error marshaling top downloaded ebooks data to JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.templateRenderer.View(w, r, "dashboard", map[string]any{
		"TotalEbooks":             totalEbooks,
		"GetTotalSendEbooks":      totalSendEbooks,
		"GetTotalClients":         totalClients,
		"EbookStats":              ebookStats,
		"DailyPurchasesJSON":      string(dailyPurchasesJSON),
		"DailyDownloadsJSON":      string(dailyDownloadsJSON),
		"TopEbooksJSON":           string(topEbooksJSON),
		"TopClients":              topClientsWithInitials,
		"TopDownloadedEbooksJSON": string(topDownloadedEbooksJSON),
	}, "admin")
}
