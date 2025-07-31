package gov

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anglesson/simple-web-server/internal/config"
)

// Testes para o HubDevService que cobrem especificamente a validação na linha 97
// do arquivo receita_federal_hub_dev.go:
// if response.Result.NomeDaPF == "" || response.Result.DataNascimento != dataNascimento {
//     return nil, errors.New("nome não encontrado na receita federal")
// }

func TestHubDevService_ConsultaCPF_ValidationErrors(t *testing.T) {
	tests := []struct {
		name           string
		cpf            string
		dataNascimento string
		mockResponse   map[string]interface{}
		expectedError  string
	}{
		{
			name:           "should return error when nome is empty",
			cpf:            "12345678901",
			dataNascimento: "01/01/1990",
			mockResponse: map[string]interface{}{
				"status":   true,
				"return":   "success",
				"consumed": 1,
				"result": map[string]interface{}{
					"numero_de_cpf":            "12345678901",
					"nome_da_pf":               "", // Nome vazio
					"data_nascimento":          "01/01/1990",
					"situacao_cadastral":       "REGULAR",
					"data_inscricao":           "01/01/2010",
					"digito_verificador":       "12",
					"comprovante_emitido":      "SIM",
					"comprovante_emitido_data": "01/01/2020",
				},
			},
			expectedError: "nome não encontrado na receita federal",
		},
		{
			name:           "should return error when data nascimento does not match",
			cpf:            "12345678901",
			dataNascimento: "01/01/1990",
			mockResponse: map[string]interface{}{
				"status":   true,
				"return":   "success",
				"consumed": 1,
				"result": map[string]interface{}{
					"numero_de_cpf":            "12345678901",
					"nome_da_pf":               "João Silva",
					"data_nascimento":          "02/01/1990", // Data diferente
					"situacao_cadastral":       "REGULAR",
					"data_inscricao":           "01/01/2010",
					"digito_verificador":       "12",
					"comprovante_emitido":      "SIM",
					"comprovante_emitido_data": "01/01/2020",
				},
			},
			expectedError: "nome não encontrado na receita federal",
		},
		{
			name:           "should return error when both nome is empty and data nascimento does not match",
			cpf:            "12345678901",
			dataNascimento: "01/01/1990",
			mockResponse: map[string]interface{}{
				"status":   true,
				"return":   "success",
				"consumed": 1,
				"result": map[string]interface{}{
					"numero_de_cpf":            "12345678901",
					"nome_da_pf":               "",           // Nome vazio
					"data_nascimento":          "02/01/1990", // Data diferente
					"situacao_cadastral":       "REGULAR",
					"data_inscricao":           "01/01/2010",
					"digito_verificador":       "12",
					"comprovante_emitido":      "SIM",
					"comprovante_emitido_data": "01/01/2020",
				},
			},
			expectedError: "nome não encontrado na receita federal",
		},
		{
			name:           "should return success when nome is not empty and data nascimento matches",
			cpf:            "12345678901",
			dataNascimento: "01/01/1990",
			mockResponse: map[string]interface{}{
				"status":   true,
				"return":   "success",
				"consumed": 1,
				"result": map[string]interface{}{
					"numero_de_cpf":            "12345678901",
					"nome_da_pf":               "João Silva",
					"data_nascimento":          "01/01/1990", // Data correta
					"situacao_cadastral":       "REGULAR",
					"data_inscricao":           "01/01/2010",
					"digito_verificador":       "12",
					"comprovante_emitido":      "SIM",
					"comprovante_emitido_data": "01/01/2020",
				},
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Criar servidor mock
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verificar se a requisição contém os parâmetros esperados
				if r.URL.Query().Get("cpf") != tt.cpf || r.URL.Query().Get("data") != tt.dataNascimento {
					t.Errorf("Parâmetros da requisição não correspondem aos esperados")
				}

				// Retornar resposta mock
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Criar serviço com URL mock
			service := &HubDevService{}

			// Substituir temporariamente a URL da API pela URL do servidor mock
			originalURL := config.AppConfig.HubDesenvolvedorApi
			config.AppConfig.HubDesenvolvedorApi = server.URL
			defer func() { config.AppConfig.HubDesenvolvedorApi = originalURL }()

			// Executar consulta
			response, err := service.ConsultaCPF(tt.cpf, tt.dataNascimento)

			// Verificar resultado
			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Esperava erro '%s', mas não houve erro", tt.expectedError)
					return
				}
				if err.Error() != tt.expectedError {
					t.Errorf("Erro esperado '%s', mas recebeu '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Não esperava erro, mas recebeu: %s", err.Error())
					return
				}
				if response == nil {
					t.Errorf("Esperava resposta válida, mas recebeu nil")
					return
				}
				if response.Result.NomeDaPF != tt.mockResponse["result"].(map[string]interface{})["nome_da_pf"] {
					t.Errorf("Nome esperado '%s', mas recebeu '%s'",
						tt.mockResponse["result"].(map[string]interface{})["nome_da_pf"],
						response.Result.NomeDaPF)
				}
			}
		})
	}
}

func TestHubDevService_ConsultaCPF_EdgeCases(t *testing.T) {
	// Teste para casos extremos e comportamentos atuais da validação
	tests := []struct {
		name           string
		cpf            string
		dataNascimento string
		mockResponse   map[string]interface{}
		expectedError  string
	}{
		{
			name:           "should return success when nome is only whitespace (current behavior - TODO: improve validation)",
			cpf:            "12345678901",
			dataNascimento: "01/01/1990",
			mockResponse: map[string]interface{}{
				"status":   true,
				"return":   "success",
				"consumed": 1,
				"result": map[string]interface{}{
					"numero_de_cpf":            "12345678901",
					"nome_da_pf":               "   ", // Apenas espaços em branco - atual validação não trata como vazio
					"data_nascimento":          "01/01/1990",
					"situacao_cadastral":       "REGULAR",
					"data_inscricao":           "01/01/2010",
					"digito_verificador":       "12",
					"comprovante_emitido":      "SIM",
					"comprovante_emitido_data": "01/01/2020",
				},
			},
			expectedError: "",
		},
		{
			name:           "should return error when data nascimento format is different",
			cpf:            "12345678901",
			dataNascimento: "01/01/1990",
			mockResponse: map[string]interface{}{
				"status":   true,
				"return":   "success",
				"consumed": 1,
				"result": map[string]interface{}{
					"numero_de_cpf":            "12345678901",
					"nome_da_pf":               "João Silva",
					"data_nascimento":          "1990-01-01", // Formato diferente
					"situacao_cadastral":       "REGULAR",
					"data_inscricao":           "01/01/2010",
					"digito_verificador":       "12",
					"comprovante_emitido":      "SIM",
					"comprovante_emitido_data": "01/01/2020",
				},
			},
			expectedError: "nome não encontrado na receita federal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Criar servidor mock
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Criar serviço com URL mock
			service := &HubDevService{}

			// Substituir temporariamente a URL da API pela URL do servidor mock
			originalURL := config.AppConfig.HubDesenvolvedorApi
			config.AppConfig.HubDesenvolvedorApi = server.URL
			defer func() { config.AppConfig.HubDesenvolvedorApi = originalURL }()

			// Executar consulta
			_, err := service.ConsultaCPF(tt.cpf, tt.dataNascimento)

			// Verificar resultado
			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Esperava erro '%s', mas não houve erro", tt.expectedError)
					return
				}
				if err.Error() != tt.expectedError {
					t.Errorf("Erro esperado '%s', mas recebeu '%s'", tt.expectedError, err.Error())
				}
			}
		})
	}
}
