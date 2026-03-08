package handler

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/anglesson/simple-web-server/internal/config"
	salesvc "github.com/anglesson/simple-web-server/internal/sales/service"
)

func WatermarkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("x-app-key") != config.AppConfig.AppKey {
		http.Error(w, "Invalid app key", http.StatusUnauthorized)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "Invalid content", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		log.Printf("Erro ao obter arquivo: %v", err)
		http.Error(w, "Erro ao obter arquivo", http.StatusBadRequest)
		return
	}
	defer file.Close()

	tempFile, err := os.CreateTemp("", "upload-*.pdf")
	if err != nil {
		log.Printf("Erro ao criar arquivo temporário: %v", err)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())

	_, err = io.Copy(tempFile, file)
	if err != nil {
		log.Printf("Erro ao copiar arquivo: %v", err)
		http.Error(w, "Erro ao processar arquivo", http.StatusInternalServerError)
		return
	}
	tempFile.Close()

	outputPath, err := salesvc.ApplyWatermarkToLocalFile(tempFile.Name(), content, fileHeader.Filename)
	if err != nil {
		log.Printf("Erro ao aplicar marca d'água: %v", err)
		http.Error(w, "Erro ao processar arquivo", http.StatusInternalServerError)
		return
	}
	defer os.Remove(outputPath)

	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(fileHeader.Filename))
	w.Header().Set("Content-Type", "application/pdf")
	http.ServeFile(w, r, outputPath)
}
