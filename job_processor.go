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
		log.Println("Failed to create BigQuery client: ", err)
		return err
	}
	defer func(client *bigquery.Client) {
		err := client.Close()
		if err != nil {
			log.Println("Failed to close BigQuery client: ", err)
		}
	}(client)

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
		log.Println("Failed to execute query: ", err)
		return err
	}

	var jobs []JobDataRow
	var count int

	for {
		var row JobDataRow
		err := rows.Next(&row)
		if errors.Is(err, iterator.Done) {
			if count == 0 {
				log.Println("The query returned an empty result set.")
			} else {
				log.Printf("Total jobs read: %d", count)
			}
			break
		}
		if err != nil {
			log.Println("Failed to read results: ", err)
			return err
		}

		jobs = append(jobs, row)
		count++
	}

	batchSize := 100
	for i := 0; i < len(jobs); i += batchSize {
		end := i + batchSize
		if end > len(jobs) {
			end = len(jobs)
		}

		batch := jobs[i:end]

		err := addToBlockchain(batch)

		if err != nil {
			log.Printf("Error processing batch %d to %d: %v", i, end, err)
		}

		fmt.Printf("Batch %d processed successfully\n", i)
	}

	fmt.Println("All jobs were processed successfully")

	return nil
}
