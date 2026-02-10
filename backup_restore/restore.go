package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const sqlFileZip = "backup.sql.gz"
const sqlFile = "backup.sql"

type Response struct {
	Dump string `json:"dump"`
}

type Solution struct {
	AliveSSN []string `json:"alive_ssns"`
}

func main() {

	res, err := http.Get("https://hackattic.com/challenges/backup_restore/problem?access_token=8f3a9cb1bc4572d2")
	if err != nil {
		log.Fatalf("Couldn't initiate GET request: %v\n", err)
	}
	defer res.Body.Close()

	var dump Response
	err = json.NewDecoder(res.Body).Decode(&dump)
	if err != nil {
		log.Fatalf("Couldn't decode json: %v\n", err)
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(dump.Dump)
	if err != nil {
		log.Fatalf("base64 decode error: %v", err)
	}

	err = createBackup(decodedBytes, sqlFileZip)
	if err != nil {
		log.Fatalf("Couldn't create backup zip file: %v\n", err)
	}

	err = decompressBackup()
	if err != nil {
		log.Fatalf("Couldn't decompress backup file: %v\n", err)
	}

	cmd := exec.Command("psql", "-d", "dump", "-f", sqlFile)
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Exec command failed: %v\n", err)
	}

	sol, err := getSSN()
	if err != nil {
		log.Fatalf("Couldn't get ssn: %v", err)
	}

	b, err := json.Marshal(sol)
	if err != nil {
		log.Fatalf("Couldn't encode json: %v\n", err)
	}
	reader := bytes.NewReader(b)

	req, err := http.NewRequest("POST", "https://hackattic.com/challenges/backup_restore/solve?access_token=8f3a9cb1bc4572d2", reader)
	if err != nil {
		log.Fatalf("Couldn't create new request: %v\n", err)
	}

	client := &http.Client{
		Timeout: 7 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Couldn't initiate request-response: %v\n", err)
	}
	defer resp.Body.Close()

	var result any
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Fatalf("Couldn't decode json: %v\n", err)
	}

	fmt.Println(result)
}

func getSSN() (Solution, error) {
	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		return Solution{}, fmt.Errorf("unable to create connection pool: %v", err)
	}
	defer dbpool.Close()

	rows, err := dbpool.Query(context.Background(), "SELECT ssn FROM criminal_records WHERE status=$1", "alive")
	if err != nil {
		return Solution{}, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	var aliveSSNs []string
	for rows.Next() {
		var ssn string
		if err := rows.Scan(&ssn); err != nil {
			return Solution{}, fmt.Errorf("scanning row field: %v", err)
		}
		aliveSSNs = append(aliveSSNs, ssn)
	}

	return Solution{
		AliveSSN: aliveSSNs,
	}, nil
}

func decompressBackup() error {
	f, err := os.Open(sqlFileZip)
	if err != nil {
		return err
	}
	defer f.Close()

	outFile, err := os.Create(sqlFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	reader, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer reader.Close()

	_, err = io.Copy(outFile, reader)
	if err != nil {
		return err
	}

	return nil
}

func createBackup(rawBytes []byte, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	f.Write(rawBytes)

	return nil
}
