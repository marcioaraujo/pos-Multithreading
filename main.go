package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Estruturas para armazenar a resposta das APIs
type Endereco struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
}

type Resultado struct {
	API      string
	Endereco Endereco
	Err      error
}

func main() {
	cep := "01153000"
	timeout := 1 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Canal para capturar o primeiro resultado
	resultChan := make(chan Resultado, 2)

	// Dispara goroutines para as duas APIs
	go buscaEndereco(ctx, resultChan, "BrasilAPI", "https://brasilapi.com.br/api/cep/v1/"+cep)
	go buscaEndereco(ctx, resultChan, "ViaCEP", "http://viacep.com.br/ws/"+cep+"/json/")

	// Captura o primeiro resultado ou timeout
	select {
	case result := <-resultChan:
		if result.Err != nil {
			fmt.Printf("Erro: %v\n", result.Err)
		} else {
			fmt.Printf("API: %s\nEndereço: %+v\n", result.API, result.Endereco)
		}
	case <-ctx.Done():
		fmt.Println("Erro: Timeout, nenhuma API respondeu em 1 segundo")
	}
}

// Função que busca o endereço e envia o resultado pelo canal
func buscaEndereco(ctx context.Context, resultChan chan<- Resultado, api string, url string) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		resultChan <- Resultado{API: api, Err: err}
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		resultChan <- Resultado{API: api, Err: err}
		return
	}
	defer resp.Body.Close()

	// Decodifica a resposta JSON
	var endereco Endereco
	if err := json.NewDecoder(resp.Body).Decode(&endereco); err != nil {
		resultChan <- Resultado{API: api, Err: err}
		return
	}

	// Envia o resultado pelo canal
	resultChan <- Resultado{API: api, Endereco: endereco}
}
