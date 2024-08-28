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
	ChunkID                  float64
	JobID                    string
	AssetID                  string
	CreatedAtDay             time.Time
	TotalDuration            float64
	TotalRewardsConsumer     float64
	TotalRewardsContentOwner float64
	UserID                   string
}

// processJobs realiza um select em d -1 para obter os dados e criar uma goroutine para cada JobData
func processJobs(datetime time.Time) error {
	client, err := bigquery.NewClient(context.Background(), "seu-projeto-id")
	if err != nil {
		return fmt.Errorf("falha ao criar cliente do BigQuery: %v", err)
	}
	defer client.Close()

	dMinus1 := datetime.AddDate(0, 0, -1).Format("2006-01-02")

	query := client.Query(fmt.Sprintf(`
		SELECT
			CHUNK_ID,
			JOB_ID,
			assetId,
			createdAtDay,
			totalDuration,
			totalRewardsConsumer,
			totalRewardsContentOwner,
			userId
		FROM
			replay-staging-353318.replayAnalyticsStaging.table_blockchain_chunked_data_of_the_day_and_asset 
		WHERE
			status <> 'FINISHED'
			AND createdAtDay = '%s'
	`, dMinus1))

	it, err := query.Read(context.Background())
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
		job := job // Captura o valor atual de `job`
		go func(job JobData) {
			handleJob(&job) // Passa o job para handleJob
		}(job)
	}

	return nil
}

func handleJob(job *JobData) {
	data := []map[string]interface{}{
		{
			"chunk_id":                    job.ChunkID,
			"job_id":                      job.JobID,
			"asset_id":                    job.AssetID,
			"created_at_day":              job.CreatedAtDay.Format("2006-01-02T15:04:05Z"), // Formata a data para ISO 8601
			"total_duration":              job.TotalDuration,
			"total_rewards_consumer":      job.TotalRewardsConsumer,
			"total_rewards_content_owner": job.TotalRewardsContentOwner,
			"user_id":                     job.UserID,
		},
	}

	// Envia os dados para o webhook com retentativa
	response, err := sendToWebhookWithRetry(data, 5) // Tenta 5 vezes em caso de falha
	if err != nil {
		// TODO: Notificar por email, slack ...
		log.Printf("Falha ao enviar job %s: %v", job.JobID, err)
	} else {
		// TODO: Atualizar o status do job no BigQuery
		log.Printf("Resposta do webhook para job %s: %s", job.JobID, response)
	}
}
