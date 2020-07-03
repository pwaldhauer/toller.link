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

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/RediSearch/redisearch-go/redisearch"
)

func saveLinkToRedis(id string, link Link) {
	c := redisearch.NewClient("localhost:6379", "tollerlink")
	sc := redisearch.NewSchema(redisearch.DefaultOptions).
		AddField(redisearch.NewTextFieldOptions("body", redisearch.TextFieldOptions{Weight: 2.0})).
		AddField(redisearch.NewTextField("url")).
		AddField(redisearch.NewTextFieldOptions("tags", redisearch.TextFieldOptions{Weight: 10.0, Sortable: true})).
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
		Set("body", link.Body).
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
		Limit(0, 50).
		SetSortBy("date", false).
		SetReturnFields("title", "body").
		Highlight([]string{"title", "body"}, "<b>", "</b>").
		Summarize("title", "body"))

	if total == 0 {
		return links
	}

	for _, doc := range docs {
		link := readLink(doc.Id)

		link.ContextTitle = fmt.Sprintf("%s", doc.Properties["title"])
		link.ContextBody = fmt.Sprintf("%s", doc.Properties["body"])

		links = append(links, link)
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
	/*
		contentPath := fmt.Sprintf("content/%s/content.html", id)

		if _, err := os.Stat(contentPath); err == nil {
			linkContent, _ := ioutil.ReadFile(contentPath)
			link.Body = fmt.Sprintf("%s", linkContent)
		}
	*/
	return link
}

type Link struct {
	Date         int64
	Url          string
	Title        string
	Tags         string
	Body         string
	ContextTitle string
	ContextBody  string
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
	saveLinkToRedis(id, link)

	dir = fmt.Sprintf("content/%s", id)

	_ = os.Mkdir(dir, 0755)

	contentPath := fmt.Sprintf("content/%s/content.html", id)
	if link.Body != "" {
		ioutil.WriteFile(contentPath, []byte(link.Body), 0644)
		link.Body = ""
	}

	filepath = fmt.Sprintf("%s/link.json", dir)

	jsonString, _ := json.Marshal(link)
	ioutil.WriteFile(filepath, jsonString, 0644)

	fmt.Fprintf(w, "Link: %+v, %s", link, filepath)

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "success"}`))
}

func getLinks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	query := r.URL.Query()

	var links []Link
	links = findInRedis(fmt.Sprintf("%s*", query.Get("q")))

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

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	log.Fatal(http.ListenAndServe(":8080", handlers.CORS(originsOk, headersOk, methodsOk)(r)))
}
