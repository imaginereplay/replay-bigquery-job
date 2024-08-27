package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func sendToWebhookWithRetry(data []map[string]interface{}) {
	// WHILE TRUE (com retentativa)
	for {
		err := sendToWebhook(data)
		if err != nil {
			log.Println("Erro ao enviar dados para o webhook, tentando novamente:", err)
			time.Sleep(2 * time.Second) // Aguarda 2 segundos antes de tentar novamente
			continue
		}
		break
	}

}

func sendToWebhook(data []map[string]interface{}) error {
	// Mock do webhook
	// Aqui você enviaria uma requisição HTTP real para o webhook
	log.Println("Enviando dados para o webhook:", data)

	// Simulando sucesso no envio
	resp, err := http.Post("http://mocked-webhook-url", "application/json", nil)
	if err != nil {
		return fmt.Errorf("erro ao enviar requisição para webhook: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("webhook retornou status code %d", resp.StatusCode)
	}

	return nil
}
