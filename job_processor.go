package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

type JobData struct {
	JobID   string
	ChunkID float64
}

func processJobs(datetime time.Time) error {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "seu-projeto-id")
	if err != nil {
		return fmt.Errorf("falha ao criar cliente do BigQuery: %v", err)
	}
	defer client.Close()

	// Executa a primeira consulta para obter JOB_ID e CHUNK_ID
	// De acordo com o dia do cronjob
	query := client.Query(`
		SELECT JOB_ID, CHUNK_ID
		FROM sua_tabela
		WHERE EXTRACT(DATE FROM data) = DATE_SUB(CURRENT_DATE(), INTERVAL 1 DAY)
	`)

	it, err := query.Read(ctx)
	if err != nil {
		return fmt.Errorf("falha ao executar query: %v", err)
	}

	var jobs []JobData

	for {
		var row JobData
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("falha ao ler resultados: %v", err)
		}
		jobs = append(jobs, row)
	}

	// Para cada JobData, dispara uma goroutine
	for _, job := range jobs {
		go handleJob(ctx, client, &job)
	}

	return nil
}

func handleJob(ctx context.Context, client *bigquery.Client, job *JobData) {
	query := client.Query(fmt.Sprintf(`
		SELECT *
		FROM tabela_block
		WHERE JOB_ID = '%s' AND CHUNK_ID = %.1f
	`, job.JobID, job.ChunkID))

	it, err := query.Read(ctx)
	if err != nil {
		log.Printf("Erro ao executar query para JOB_ID: %s, CHUNK_ID: %.1f: %v", job.JobID, job.ChunkID, err)
		return
	}

	// Processa os dados retornados (mockado para este exemplo)
	var data []map[string]interface{}

	for {
		var row map[string]interface{}
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Erro ao ler resultados para JOB_ID: %s, CHUNK_ID: %.1f: %v", job.JobID, job.ChunkID, err)
			return
		}
		data = append(data, row)
	}

	// Envia os dados para o webhook com retentativa
	sendToWebhookWithRetry(data)
}
