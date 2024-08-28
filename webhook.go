package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

// sendToWebhookWithRetry tenta enviar dados para o webhook com retentativa limitada
func sendToWebhookWithRetry(data []map[string]interface{}, retries int) (string, error) {
	var lastError error
	for attempt := 1; attempt <= retries; attempt++ {
		response, err := sendToWebhook(data)
		if err != nil {
			lastError = err
			log.Printf("Tentativa %d/%d falhou: %v", attempt, retries, err)
			if attempt < retries {
				time.Sleep(time.Duration(attempt) * time.Second) // Espera exponencial
			}
		} else {
			log.Println("Dados enviados com sucesso para o webhook.")
			return response, nil
		}
	}
	// Retorna o último erro após todas as tentativas falharem
	log.Println("Número máximo de tentativas atingido. Abortando.")
	return "", lastError
}

// sendToWebhook envia os dados para o webhook como uma requisição HTTP POST
func sendToWebhook(data []map[string]interface{}) (string, error) {
	jsonData, err := sonic.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar dados para JSON: %w", err)
	}

	req, err := http.NewRequest("POST", "node_api_url_here", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição HTTP: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao enviar requisição para o webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("webhook retornou status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler corpo da resposta: %w", err)
	}

	return string(body), nil
}
