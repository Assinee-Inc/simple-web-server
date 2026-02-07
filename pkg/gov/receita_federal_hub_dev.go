package gov

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/anglesson/simple-web-server/internal/config"
)

type HubDevService struct {
}

func NewHubDevService() ReceitaFederalService {
	return &HubDevService{}
}

func (rf *HubDevService) ConsultaCPF(nome, cpf, dataNascimento string) (*ReceitaFederalResponse, error) {
	uri := fmt.Sprintf("%s/v2/cpf/?cpf=%s&data=%s&token=%s", config.AppConfig.HubDesenvolvedorApi, cpf, dataNascimento, config.AppConfig.HubDesenvolvedorToken)

	if !config.AppConfig.HubDesenvolvedorActive {
		log.Println("HubDesenvolvedorActive is false. Returning mock response.")
		return &ReceitaFederalResponse{
			Status: true,
			Return: "Mocked response",
			Result: ConsultaData{
				NumeroDeCPF:       cpf,
				NomeDaPF:          nome,
				DataNascimento:    dataNascimento,
				SituacaoCadastral: "REGULAR",
				DigitoVerificador: "00",
			},
		}, nil
	}

	request, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		log.Printf("Erro ao consultar dados na receita federal. Error: %s", err.Error())
		return nil, fmt.Errorf("erro ao preparar consulta: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Printf("Erro ao fazer requisição para receita federal. Error: %s", err.Error())
		return nil, fmt.Errorf("erro ao consultar receita federal: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Erro ao ler resposta da receita federal. Error: %s", err.Error())
		return nil, fmt.Errorf("erro ao ler resposta da receita federal: %w", err)
	}

	// Log the raw response
	log.Printf("Resposta bruta da API: %s", string(bodyBytes))

	// Criar um map para armazenar a resposta
	var responseMap map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &responseMap); err != nil {
		log.Printf("Erro ao fazer parse da resposta: %s", err.Error())
		return nil, fmt.Errorf("erro ao processar resposta da receita federal: %w", err)
	}

	// Verificar se a resposta tem o formato esperado
	if responseMap == nil {
		return nil, errors.New("resposta inválida da receita federal")
	}

	// Verificar status da resposta
	status, ok := responseMap["status"].(bool)
	if !ok {
		return nil, errors.New("status inválido na resposta da receita federal")
	}

	if !status {
		return &ReceitaFederalResponse{
			Status: false,
		}, nil
	}

	// Extrair o resultado
	resultMap, ok := responseMap["result"].(map[string]interface{})
	if !ok {
		log.Printf("Erro ao extrair resultado da resposta: %v", responseMap)
		return nil, errors.New("formato de resposta inválido da receita federal")
	}

	// Criar e popular o objeto
	response := &ReceitaFederalResponse{
		Status:   status,
		Return:   fmt.Sprintf("%v", responseMap["return"]),
		Consumed: int(responseMap["consumed"].(float64)),
		Result: ConsultaData{
			NumeroDeCPF:            fmt.Sprintf("%v", resultMap["numero_de_cpf"]),
			NomeDaPF:               fmt.Sprintf("%v", resultMap["nome_da_pf"]),
			DataNascimento:         fmt.Sprintf("%v", resultMap["data_nascimento"]),
			SituacaoCadastral:      fmt.Sprintf("%v", resultMap["situacao_cadastral"]),
			DataInscricao:          fmt.Sprintf("%v", resultMap["data_inscricao"]),
			DigitoVerificador:      fmt.Sprintf("%v", resultMap["digito_verificador"]),
			ComprovanteEmitido:     fmt.Sprintf("%v", resultMap["comprovante_emitido"]),
			ComprovanteEmitidoData: fmt.Sprintf("%v", resultMap["comprovante_emitido_data"]),
		},
	}

	log.Printf("Objeto populado: %+v", response)

	if response.Result.NomeDaPF == "" || response.Result.DataNascimento != dataNascimento {
		return nil, errors.New("nome não encontrado na receita federal")
	}

	return response, nil
}
