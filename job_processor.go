package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/bigquery"
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

func processJobs(secretName string) error {
	fmt.Println("Processing jobs at:", time.Now())

	client, err := GetBigQueryClient(secretName)
	if err != nil {
		log.Println("Falha ao criar cliente do BigQuery: ", err)
		return err
	}
	defer client.Close()

	dMinus1 := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
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

	fmt.Printf("Total de jobs a serem processados: %d\n", len(jobs))
	fmt.Println("Todos os jobs: ", jobs)

	return nil

	// batchSize := 100
	// for i := 0; i < len(jobs); i += batchSize {
	// 	end := i + batchSize
	// 	if end > len(jobs) {
	// 		end = len(jobs)
	// 	}

	// 	batch := jobs[i:end]

	// 	err := addToBlockchain(batch)

	// 	if err != nil {
	// 		log.Printf("Erro ao processar o batch %d ao %d: %v", i, end, err)
	// 	}

	// 	fmt.Printf("Batch %d processado com sucesso\n", i)
	// }

	// fmt.Println("Todos os jobs foram processados com sucesso")

	// return nil
}
