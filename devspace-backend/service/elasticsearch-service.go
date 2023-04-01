package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func CreateElasticsearchIndex(indexName string) error {
	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %s", err)
	}

	// Check if index already exists
	exists, err := es.Indices.Exists([]string{indexName})
	if err != nil {
		return fmt.Errorf("error checking if index exists: %s", err)
	}
	if exists.StatusCode == 200 {
		return fmt.Errorf("index %s already exists", indexName)
	}

	// Create index
	req := esapi.IndicesCreateRequest{
		Index: indexName,
	}

	res, err := req.Do(context.Background(), es)
	if err != nil {
		return fmt.Errorf("error creating index: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var errorResponse struct {
			Error struct {
				Type   string `json:"type"`
				Reason string `json:"reason"`
			} `json:"error"`
		}

		if err := json.NewDecoder(res.Body).Decode(&errorResponse); err != nil {
			return fmt.Errorf("error parsing error response: %s", err)
		}

		return fmt.Errorf("error creating index: %s", errorResponse.Error.Reason)
	}

	log.Printf("Successfully created Elasticsearch index %s", indexName)
	return nil
}
