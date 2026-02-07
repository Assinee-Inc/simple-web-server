package gov

type ReceitaFederalService interface {
	ConsultaCPF(nome, cpf, dataNascimento string) (*ReceitaFederalResponse, error)
}

type ConsultaData struct {
	NumeroDeCPF            string `json:"numero_de_cpf"`
	NomeDaPF               string `json:"nome_da_pf"`
	DataNascimento         string `json:"data_nascimento"`
	SituacaoCadastral      string `json:"situacao_cadastral"`
	DataInscricao          string `json:"data_inscricao"`
	DigitoVerificador      string `json:"digito_verificador"`
	ComprovanteEmitido     string `json:"comprovante_emitido"`
	ComprovanteEmitidoData string `json:"comprovante_emitido_data"`
}

type ReceitaFederalResponse struct {
	Status   bool         `json:"status"`
	Return   string       `json:"return"`
	Consumed int          `json:"consumed"`
	Result   ConsultaData `json:"result"`
}
