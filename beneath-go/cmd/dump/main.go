package main

import (
	"context"
	"encoding/hex"
	"log"
	"os"

	"cloud.google.com/go/bigtable"
)

func main() {
	os.Setenv("BIGTABLE_PROJECT_ID", "")
	os.Setenv("BIGTABLE_EMULATOR_HOST", "localhost:8086")

	// prepare BigTable client
	ctx := context.Background()
	client, err := bigtable.NewClient(ctx, "", "")
	if err != nil {
		log.Fatalf("Could not create bigtable client: %v", err)
	}

	// create table
	table := client.Open("records")

	// dump all rows
	rr := bigtable.PrefixRange("")
	err = table.ReadRows(context.Background(), rr, func(row bigtable.Row) bool {
		item := row["cf0"][0]
		log.Printf("\tKey: %s; Value: %s; Timestamp: %v\n", hex.EncodeToString([]byte(item.Row)), hex.EncodeToString(item.Value), item.Timestamp)
		return true
	})
	if err != nil {
		log.Fatalf("Could not create dump rows: %v", err)
	}
}
