package service

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anglesson/simple-web-server/pkg/storage"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func getFilename(original string) string {
	now := time.Now()

	timestamp := now.Format("20060102_150405")

	log.Println(timestamp)

	fileName := filepath.Base(original)

	ext := filepath.Ext(fileName)
	name := strings.TrimSuffix(fileName, ext)

	newFileName := fmt.Sprintf("%s_%s%s", name, timestamp, ext)

	log.Println(newFileName)

	return newFileName
}

// ApplyWatermark aplica uma marca d'água ao PDF com as informações do usuário
func ApplyWatermark(s3Key, content string) (string, error) {
	log.Printf("Baixando arquivo do S3: %s", s3Key)
	localFilePath, err := storage.GetFile(s3Key)
	if err != nil {
		return "", fmt.Errorf("erro ao baixar arquivo do S3: %w", err)
	}
	defer func() {
		if err := os.Remove(localFilePath); err != nil {
			log.Printf("Erro ao remover arquivo temporário %s: %v", localFilePath, err)
		}
	}()

	log.Printf("Arquivo baixado para: %s", localFilePath)

	return ApplyWatermarkToLocalFile(localFilePath, content, s3Key)
}

// ApplyWatermarkToLocalFile aplica marca d'água a um arquivo local
func ApplyWatermarkToLocalFile(localFilePath, content, originalName string) (string, error) {
	outputPDF := getFilename(originalName)

	if err := os.MkdirAll("./temp", 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	conf := model.NewDefaultConfiguration()

	watermarkStrings := make([]string, 0)
	watermarkStrings = append(watermarkStrings, "font:Helvetica, points:20, pos:c, fillc:#000000, scale:1.0, rot:45, op:0.1")
	watermarkStrings = append(watermarkStrings, "font:Helvetica, points:20, pos:bc, fillc:#000000, scale:1.0, rot:0, op:0.1")
	watermarkStrings = append(watermarkStrings, "font:Helvetica, points:20, pos:l, fillc:#000000, scale:1.0, rot:90, op:0.1")
	watermarkStrings = append(watermarkStrings, "font:Helvetica, points:20, pos:r, fillc:#000000, scale:1.0, rot:-90, op:0.1")
	watermarkStrings = append(watermarkStrings, "font:Helvetica, points:20, pos:tc, fillc:#000000, scale:1.0, rot:0, op:0.1")

	currentInputPath := localFilePath
	for key, wms := range watermarkStrings {
		if key > 0 {
			currentInputPath = outputPDF
		}
		log.Printf("Adicionando marca d'água: %s", wms)

		wm, errParse := pdfcpu.ParseTextWatermarkDetails(
			content,
			wms,
			true,
			types.POINTS,
		)

		if errParse != nil {
			log.Fatal(errParse)
		}

		err := api.AddWatermarksFile(currentInputPath, outputPDF, nil, wm, conf)
		if err != nil {
			fmt.Println("Erro ao configurar o stamp:", err)
			return "", err
		}
	}

	fmt.Println("Stamp aplicado com sucesso!")
	return outputPDF, nil
}
