package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/api/iterator"
)

type JobDataRow struct {
	ChunkID                  int64     `bigquery:"CHUNK_ID"`
	JobID                    string    `bigquery:"JOB_ID"`
	AssetID                  *string   `bigquery:"assetId"`
	CreatedAtDay             time.Time `bigquery:"createdAtDay"`
	TotalDuration            float64   `bigquery:"totalDuration"`
	TotalRewardsConsumer     float64   `bigquery:"totalRewardsConsumer"`
	TotalRewardsContentOwner float64   `bigquery:"totalRewardsContentOwner"`
	UserID                   string    `bigquery:"userId"`
}

// processJobs realiza um select em d-1 para obter os dados e criar uma goroutine para cada JobData
func processJobs(datetime time.Time, secretName string) error {

	client, err := GetBigQueryClient(secretName)
	if err != nil {
		log.Println("Falha ao criar cliente do BigQuery: ", err)
		return err
	}

	defer client.Close()

	dMinus1 := datetime.AddDate(0, 0, -1).Format("2006-01-02")
	queryStr := fmt.Sprintf(`
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
	`, dMinus1)

	query := client.Query(queryStr)

	rows, err := query.Read(context.Background())
	if err != nil {
		log.Println("Falha ao executar query: ", err)
		return err
	}

	var jobs []JobDataRow
	var count int

	for {
		var row JobDataRow
		err := rows.Next(&row)
		if err == iterator.Done {
			if count == 0 {
				log.Println("A query retornou um conjunto de resultados vazio.")
			} else {
				log.Printf("Total de jobs lidos: %d", count)
			}
			break
		}
		if err != nil {
			log.Println("Falha ao ler resultados: ", err)
			return err
		}
		jobs = append(jobs, row)
		count++
	}

	chunkedData := make(map[int64][]JobDataRow)
	for _, job := range jobs {
		chunkedData[job.ChunkID] = append(chunkedData[job.ChunkID], job)
	}

	for chunkID, jobGroup := range chunkedData {
		jobGroup := jobGroup
		go func(chunkID int64, jobGroup []JobDataRow) {
			handleJobGroup(chunkID, jobGroup)
		}(chunkID, jobGroup)
	}

	return nil
}

func handleJobGroup(chunkID int64, jobs []JobDataRow) {
	data := make([]map[string]interface{}, len(jobs))
	for i, job := range jobs {
		data[i] = map[string]interface{}{
			"chunk_id":                    job.ChunkID,
			"job_id":                      job.JobID,
			"asset_id":                    job.AssetID,
			"created_at_day":              job.CreatedAtDay.Format("2006-01-02T15:04:05Z"), // Formata a data para ISO 8601
			"total_duration":              job.TotalDuration,
			"total_rewards_consumer":      job.TotalRewardsConsumer,
			"total_rewards_content_owner": job.TotalRewardsContentOwner,
			"user_id":                     job.UserID,
		}
	}

	response, err := sendToWebhookWithRetry(data, 5)
	if err != nil {
		// TODO: Notificar por email, slack ...
		log.Printf("Falha ao enviar jobs do chunk %d: %v", chunkID, err)
	} else {
		// TODO: Atualizar o status do chunk no BigQuery
		log.Printf("Resposta do webhook para jobs do chunk %d: %s", chunkID, response)
	}
}
