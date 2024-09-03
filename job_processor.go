package main

import (
	"cloud.google.com/go/bigquery"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"google.golang.org/api/iterator"
)

type JobDataRow struct {
	JobID                    string              `bigquery:"JOB_ID"`
	ChunkID                  float64             `bigquery:"CHUNK_ID"`
	UserID                   string              `bigquery:"userId"`
	AssetID                  bigquery.NullString `bigquery:"assetId"`
	TotalDuration            int64               `bigquery:"totalDuration"`
	TotalRewardsConsumer     float64             `bigquery:"totalRewardsConsumer"`
	TotalRewardsContentOwner float64             `bigquery:"totalRewardsContentOwner"`
	CreatedAtDay             time.Time           `bigquery:"createdAtDay"`
	Status                   bigquery.NullString `bigquery:"status"`
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
			userId,
			status
		FROM
			replay-353318.replayAnalytics.table_blockchain_chunked_data_of_the_day_and_asset
		WHERE
			createdAtDay = '%s' AND
		    status IS NULL
	`, dMinus1)

	// fmt.Println(queryStr)

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
		if errors.Is(err, iterator.Done) {
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

	chunkedData := make(map[float64][]JobDataRow)
	for _, job := range jobs {
		chunkedData[job.ChunkID] = append(chunkedData[job.ChunkID], job)
	}

	for chunkID, jobGroup := range chunkedData {
		jobGroup := jobGroup
		go func(chunkID float64, jobGroup []JobDataRow) {
			handleJobGroup(chunkID, jobGroup)
		}(chunkID, jobGroup)
	}

	return nil
}

func handleJobGroup(chunkID float64, jobs []JobDataRow) {
	data := make([]map[string]any, len(jobs))

	i := 0
	for i = 0; i <= len(jobs); i++ {
		job := jobs[i]
		var assetID any
		if job.AssetID.Valid {
			assetID = job.AssetID.StringVal
		} else {
			assetID = nil
		}

		var status any
		if job.Status.Valid {
			status = job.Status.StringVal
		} else {
			status = nil
		}

		data[i] = map[string]any{
			"chunk_id":                    job.ChunkID,
			"job_id":                      job.JobID,
			"asset_id":                    assetID,
			"created_at_day":              job.CreatedAtDay.Format("2006-01-02T15:04:05Z"),
			"total_duration":              job.TotalDuration,
			"total_rewards_consumer":      job.TotalRewardsConsumer,
			"total_rewards_content_owner": job.TotalRewardsContentOwner,
			"user_id":                     job.UserID,
			"status":                      status,
		}
	}

	response, err := sendToWebhookWithRetry(data, 5)
	if err != nil {
		log.Printf("Falha ao enviar jobs do chunk %f: %v", chunkID, err)
	} else {
		log.Printf("Resposta do webhook para jobs do chunk %f: %s", chunkID, response)
	}
}
