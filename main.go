package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	"github.com/RediSearch/redisearch-go/redisearch"
)

func saveLinkToRedis(id string, link Link) {
	c := redisearch.NewClient("localhost:6379", "tollerlink")
	sc := redisearch.NewSchema(redisearch.DefaultOptions).
		AddField(redisearch.NewTextField("body")).
		AddField(redisearch.NewTextField("url")).
		AddField(redisearch.NewTextFieldOptions("tags", redisearch.TextFieldOptions{Sortable: true})).
		AddField(redisearch.NewTextFieldOptions("title", redisearch.TextFieldOptions{Weight: 5.0, Sortable: true})).
		AddField(redisearch.NewNumericField("date"))

	if _, err := c.Info(); err != nil {
		if err := c.CreateIndex(sc); err != nil {
			log.Fatal(err)
		}
	}

	doc := redisearch.NewDocument(id, 1.0)
	doc.Set("title", link.Title).
		Set("url", link.Url).
		Set("tags", link.Tags).
		Set("date", link.Date)

	if err := c.Index([]redisearch.Document{doc}...); err != nil {
		log.Fatal(err)
	}
}

func findInRedis(query string) (links []Link) {
	c := redisearch.NewClient("localhost:6379", "tollerlink")

	var docs []redisearch.Document
	var total int

	docs, total, _ = c.Search(redisearch.NewQuery(query).
		SetLanguage("german").
		Limit(0, 10).
		SetSortBy("date", false).
		SetReturnFields("title"))

	if total == 0 {
		return links
	}

	for _, doc := range docs {
		links = append(links, readLink(doc.Id))
	}

	return links
}

func readLink(id string) (link Link) {
	var filepath string = fmt.Sprintf("content/%s/link.json", id)

	var content, _ = ioutil.ReadFile(filepath)

	err := json.Unmarshal(content, &link)
	if err != nil {
		log.Fatal(err)
		return
	}

	return link
}

/*
func ExampleClient() {

	// Create a client. By default a client is schemaless
	// unless a schema is provided when creating the index


	// Create a schema


	// Drop an existing index. If the index does not exist an error is returned
	c.Drop()

	// Create the index with the given schema


	// Create a document with an id and given score


	// Searching with limit and sorting
	docs, total, err := c.Search(redisearch.NewQuery("hello world").
		Limit(0, 2).
		SetReturnFields("title"))

	fmt.Println(docs[0].Id, docs[0].Properties["title"], total, err)
	// Output: doc1 Hello world 1 <nil>
}
*/
// https://dev.to/moficodes/build-your-first-rest-api-with-go-2gcj

type Link struct {
	Date  int64
	Url   string
	Title string
	Tags  string
}

func createLink(w http.ResponseWriter, r *http.Request) {
	var link Link

	err := json.NewDecoder(r.Body).Decode(&link)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	link.Date = time.Now().Unix()

	t := time.Now()

	h := sha1.New()
	h.Write([]byte(link.Url))

	var id string
	var dir string
	var filepath string

	id = fmt.Sprintf("%s_%x", t.Format("20060102150405"), h.Sum(nil))

	dir = fmt.Sprintf("content/%s", id)

	_ = os.Mkdir(dir, 0755)

	filepath = fmt.Sprintf("%s/link.json", dir)

	jsonString, _ := json.Marshal(link)
	ioutil.WriteFile(filepath, jsonString, 0644)

	fmt.Fprintf(w, "Link: %+v, %s", link, filepath)

	saveLinkToRedis(id, link)

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "success"}`))
}

func getLinks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	query := r.URL.Query()

	var links []Link
	links = findInRedis(query.Get("q"))

	jsonString, _ := json.Marshal(links)

	w.Write(jsonString)
}

func main() {
	args := os.Args[1:]

	if len(args) > 0 && args[0] == "index" {
		log.Fatal("lool")
		return
	}

	r := mux.NewRouter()

	r.HandleFunc("/api/link", getLinks).Methods(http.MethodGet)
	r.HandleFunc("/api/link", createLink).Methods(http.MethodPost)

	log.Fatal(http.ListenAndServe(":8080", r))
}
