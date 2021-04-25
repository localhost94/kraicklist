package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"context"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// define http handlers
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/search", LoadDatabase())
	// define port, we need to set it as env for Heroku deployment
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}
	// start server
	fmt.Printf("Server is listening on %s...", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatalf("unable to start server due: %v", err)
	}
}

func handleSearch(s *Searcher) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// fetch query string from query params
			q := r.URL.Query().Get("q")
			if len(q) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("missing search query in query params"))
				return
			}
			// search relevant records
			records, err := s.Search(q)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			// output success response
			buf := new(bytes.Buffer)
			encoder := json.NewEncoder(buf)
			encoder.Encode(records)
			w.Header().Set("Content-Type", "application/json")
			w.Write(buf.Bytes())
		},
	)
}

type Searcher struct {
	records []Record
}

func LoadDatabase() http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// fetch query string from query params
			q := r.URL.Query().Get("q")
			if len(q) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("missing search query in query params"))
				return
			}
			credential := options.Credential{
					Username: "root",
					Password: "bismillah",
			}
			uri := "mongodb://mongo:27017"
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri)).SetAuth(credential)
			if err != nil { log.Fatal(err) }

			collection := client.Database("kraicklist").Collection("data")

			query := bson.M{
				"$text": bson.M{
					"$search": q,
				},
			}
			cur, err := collection.Find(ctx, query)
			defer cur.Close(ctx)

			var records []Record
			for cur.Next(ctx) {
				var result Record
				err := cur.Decode(&result)
				if err != nil { log.Fatal(err) }
				
				records = append(records, result)
			}
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			// output success response
			buf := new(bytes.Buffer)
			encoder := json.NewEncoder(buf)
			encoder.Encode(records)
			w.Header().Set("Content-Type", "application/json")
			w.Write(buf.Bytes())
		},
	)
}

func (s *Searcher) Load(filepath string) error {
	// open file
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("unable to open source file due: %v", err)
	}
	defer file.Close()
	// read as gzip
	reader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("unable to initialize gzip reader due: %v", err)
	}
	// read the reader using scanner to contstruct records
	var records []Record
	cs := bufio.NewScanner(reader)
	for cs.Scan() {
		var r Record
		err = json.Unmarshal(cs.Bytes(), &r)
		if err != nil {
			continue
		}
		records = append(records, r)
	}
	s.records = records

	return nil
}

func (s *Searcher) Search(query string) ([]Record, error) {
	var result []Record
	for _, record := range s.records {
		if strings.Contains(record.Title, query) || strings.Contains(record.Content, query) {
			result = append(result, record)
		}
	}
	return result, nil
}

type Record struct {
	ID        int64    `json:"id"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	ThumbURL  string   `json:"thumb_url"`
	Tags      []string `json:"tags"`
	UpdatedAt int64    `json:"updated_at"`
	ImageURLs []string `json:"image_urls"`
}
