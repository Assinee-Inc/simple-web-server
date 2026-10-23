package handler

import (
	"net/http"
	"strings"

	accountmw "github.com/Assinee-Inc/docffy/internal/account/handler/middleware"
	accountrepo "github.com/Assinee-Inc/docffy/internal/account/repository"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	"github.com/Assinee-Inc/docffy/internal/shared/web"
)

// TopClientWithInitials extends TopClient with initials
type TopClientWithInitials struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	TotalPurchases int64  `json:"total_purchases"`
	Initials       string `json:"initials"`
}

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

type DashboardHandler struct{}

func NewDashboardHandler(_ ...any) *DashboardHandler {
	return &DashboardHandler{}
}

func (h *DashboardHandler) DashboardView(w http.ResponseWriter, r *http.Request) {
	var userID uint
	if uid, ok := authmw.GetAuthenticatedUserID(r.Context()); ok && uid > 0 {
		userID = uid
	} else if creator := accountmw.CreatorFromContext(r.Context()); creator != nil && creator.UserID > 0 {
		userID = creator.UserID
	} else if loggedUser := authmw.Auth(r); loggedUser != nil && loggedUser.ID > 0 {
		userID = loggedUser.ID
	} else {
		web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autorizado", nil)
		return
	}

	dashRepository := accountrepo.NewDashboardRepository(userID)

	totalEbooks := dashRepository.GetTotalEbooks()
	totalSendEbooks := dashRepository.GetTotalSendEbooks()
	totalClients := dashRepository.GetTotalClients()
	ebookStats, _ := dashRepository.GetEbookStats()

	dailyPurchases, _ := dashRepository.GetDailyPurchases()
	dailyDownloads, _ := dashRepository.GetDailyDownloads()
	topEbooks, _ := dashRepository.GetTopEbooks()
	topClients, _ := dashRepository.GetTopClients()
	topDownloadedEbooks, _ := dashRepository.GetTopDownloadedEbooks()

	topClientsWithInitials := make([]TopClientWithInitials, 0, len(topClients))
	for _, client := range topClients {
		topClientsWithInitials = append(topClientsWithInitials, TopClientWithInitials{
			Name:           client.Name,
			Email:          client.Email,
			TotalPurchases: client.TotalPurchases,
			Initials:       getInitials(client.Name),
		})
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success":               true,
		"total_ebooks":          totalEbooks,
		"total_sent_ebooks":     totalSendEbooks,
		"total_clients":         totalClients,
		"ebook_stats":           ebookStats,
		"daily_purchases":       dailyPurchases,
		"daily_downloads":       dailyDownloads,
		"top_ebooks":            topEbooks,
		"top_clients":           topClientsWithInitials,
		"top_downloaded_ebooks": topDownloadedEbooks,
	})
}
