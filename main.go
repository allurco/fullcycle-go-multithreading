package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type CepData interface {
}

type BrasilApiData struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}

type ViaCepData struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

type OutputCepData struct {
	Api          string `json:"api"`
	Street       string `json:"street"`
	Neighborhood string `json:"neighborhood"`
	Complement   string `json:"complement"`
	City         string `json:"city"`
	State        string `json:"state"`
}

func main() {

	var cep string
	fmt.Print("Digite o CEP: ")
	fmt.Scan(&cep)
	result := GetAddress(cep)

	fmt.Println(result)
}

func GetAddress(cep string) string {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ch := make(chan string)

	go fetchCepApis("https://brasilapi.com.br/api/cep/v1/"+cep, ch, ctx)
	go fetchCepApis("http://viacep.com.br/ws/"+cep+"/json/", ch, ctx)

	select {
	case result := <-ch:
		return result
	case <-ctx.Done():
		return "request timeout"
	}
}

func fetchCepApis(cepUrl string, ch chan string, ctx context.Context) {

	u, err := url.Parse(cepUrl)
	if err != nil {
		ch <- fmt.Sprintf("Error: %s", err)
	}

	apiType := u.Host

	request, err := http.NewRequestWithContext(ctx, "GET", cepUrl, nil)
	if err != nil {
		ch <- fmt.Sprintf("Error: %s", err)
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			ch <- "Time out exceeded"
			return
		} else {
			ch <- fmt.Sprintf("Error: %s", err)
		}

	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- fmt.Sprintf("Error: %s", err)
	}

	var data CepData

	if apiType == "brasilapi.com.br" {
		data = new(BrasilApiData)
	} else {
		data = new(ViaCepData)
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		ch <- fmt.Sprintf("Error: %s", err)
	}

	var outputCepData OutputCepData
	outputCepData.Api = apiType

	switch v := data.(type) {
	case *BrasilApiData:
		outputCepData.Street = v.Street
		outputCepData.Complement = ""
		outputCepData.Neighborhood = v.Neighborhood
		outputCepData.City = v.City
		outputCepData.State = v.State
	case *ViaCepData:
		outputCepData.Street = v.Logradouro
		outputCepData.Complement = v.Complemento
		outputCepData.Neighborhood = v.Bairro
		outputCepData.City = v.Localidade
		outputCepData.State = v.Uf
	}

	ch <- fmt.Sprintf("Api: %s \nLougradouro: %s \nComplemento: %s\nBairro: %s\nCidade: %s\nEstado: %s", outputCepData.Api, outputCepData.Street, outputCepData.Complement, outputCepData.Neighborhood, outputCepData.City, outputCepData.State)
}
