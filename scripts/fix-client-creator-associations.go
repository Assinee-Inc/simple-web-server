package main

import (
	"fmt"
	"log"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/pkg/database"
)

func main() {
	// Conectar ao banco de dados
	database.Connect()

	log.Println("Iniciando correção de associações cliente-creator...")

	// Buscar todos os clientes que fizeram compras mas não têm associação direta com creators
	var clientsToFix []models.Client
	err := database.DB.
		Preload("Creators").
		Preload("Purchases.Ebook.Creator").
		Where("id IN (SELECT DISTINCT client_id FROM purchases)").
		Find(&clientsToFix).Error

	if err != nil {
		log.Fatalf("Erro ao buscar clientes: %v", err)
	}

	log.Printf("Encontrados %d clientes com compras para verificar...", len(clientsToFix))

	fixed := 0
	for _, client := range clientsToFix {
		// Verificar se o cliente já tem associações diretas com creators
		if len(client.Creators) > 0 {
			log.Printf("Cliente %s (ID: %d) já tem %d creators associados, pulando...",
				client.Name, client.ID, len(client.Creators))
			continue
		}

		// Coletar todos os creators únicos dos ebooks que o cliente comprou
		creatorsMap := make(map[uint]*models.Creator)
		for _, purchase := range client.Purchases {
			if purchase.Ebook.ID > 0 && purchase.Ebook.Creator.ID > 0 {
				creatorsMap[purchase.Ebook.Creator.ID] = &purchase.Ebook.Creator
			}
		}

		if len(creatorsMap) == 0 {
			log.Printf("Cliente %s (ID: %d) não tem creators válidos nas compras, pulando...",
				client.Name, client.ID)
			continue
		}

		// Converter map para slice
		var creators []*models.Creator
		for _, creator := range creatorsMap {
			creators = append(creators, creator)
		}

		// Atualizar as associações
		err = database.DB.Model(&client).Association("Creators").Replace(creators)
		if err != nil {
			log.Printf("Erro ao atualizar associações do cliente %s (ID: %d): %v",
				client.Name, client.ID, err)
			continue
		}

		fixed++
		log.Printf("✅ Cliente %s (ID: %d) associado com %d creators",
			client.Name, client.ID, len(creators))
	}

	log.Printf("Correção concluída! %d clientes foram corrigidos.", fixed)

	// Verificação final
	var totalClients int64
	database.DB.Model(&models.Client{}).Count(&totalClients)

	var clientsWithCreators int64
	database.DB.Table("client_creators").Select("DISTINCT client_id").Count(&clientsWithCreators)

	fmt.Printf("\nResumo:\n")
	fmt.Printf("Total de clientes: %d\n", totalClients)
	fmt.Printf("Clientes com associações: %d\n", clientsWithCreators)
	fmt.Printf("Clientes sem associações: %d\n", totalClients-clientsWithCreators)
}
