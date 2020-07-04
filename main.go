package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/RediSearch/redisearch-go/redisearch"
	"github.com/go-co-op/gocron"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	strip "github.com/grokify/html-strip-tags-go"
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

	var opts redisearch.IndexingOptions
	opts.Replace = true
	opts.Partial = true

	if err := c.IndexOptions(opts, []redisearch.Document{doc}...); err != nil {
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

func reIndexLinkFromPath(path string) {
	var link Link
	var content, _ = ioutil.ReadFile(path)

	err := json.Unmarshal(content, &link)
	if err != nil {
		log.Fatal(err)
		return
	}

	contentPath := filepath.Join(filepath.Dir(path), "content.html")

	if _, err := os.Stat(contentPath); err == nil {
		linkContent, _ := ioutil.ReadFile(contentPath)
		link.Body = strip.StripTags(string(linkContent))
	}

	parent := filepath.Base(filepath.Dir(path))

	saveLinkToRedis(parent, link)
}

func readLinkFromPath(path string) (link Link) {
	var content, _ = ioutil.ReadFile(path)

	err := json.Unmarshal(content, &link)
	if err != nil {
		log.Fatal(err)
		return
	}

	return link
}

func readLink(id string) (link Link) {
	return readLinkFromPath(fmt.Sprintf("content/%s/link.json", id))
}

func saveLinkToPath(path string, link Link) {
	jsonString, _ := json.Marshal(link)
	ioutil.WriteFile(path, jsonString, 0644)
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
	} else {
		ioutil.WriteFile(fmt.Sprintf("content/%s/.needs-content", id), []byte("t"), 0644)
	}

	if link.Title == "" || link.Title == link.Url {
		ioutil.WriteFile(fmt.Sprintf("content/%s/.needs-title", id), []byte("t"), 0644)
	}

	filepath = fmt.Sprintf("%s/link.json", dir)

	saveLinkToPath(filepath, link)

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

func tokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		if token == os.Getenv("TOKEN") {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})
}

func task() {
	neededTitles, _ := filepath.Glob("./content/**/.needs-title")
	for _, file := range neededTitles {
		linkFile := filepath.Join(filepath.Dir(file), "link.json")
		link := readLinkFromPath(linkFile)

		response, error := http.Get(link.Url)
		if error != nil {
			log.Fatal(error)
			continue
		}

		defer response.Body.Close()

		body, error := ioutil.ReadAll(response.Body)
		if error != nil {
			log.Fatal(error)
			continue
		}

		r, _ := regexp.Compile("<title>(.*?)</title>")

		matches := r.FindStringSubmatch(string(body))
		title := html.UnescapeString(matches[1])

		link.Title = title

		saveLinkToPath(linkFile, link)
		os.Remove(file)

		reIndexLinkFromPath(linkFile)

		fmt.Println(title)
	}

	neededContent, _ := filepath.Glob("./content/**/.needs-content")
	for _, file := range neededContent {
		linkFile := filepath.Join(filepath.Dir(file), "link.json")
		contentFile := filepath.Join(filepath.Dir(file), "content.html")
		link := readLinkFromPath(linkFile)

		response, error := http.Get(link.Url)
		if error != nil {
			log.Fatal(error)
			continue
		}

		defer response.Body.Close()

		body, error := ioutil.ReadAll(response.Body)
		if error != nil {
			log.Fatal(error)
			continue
		}

		ioutil.WriteFile(contentFile, []byte(string(body)), 0644)

		os.Remove(file)

		reIndexLinkFromPath(linkFile)

		fmt.Println("content loaded")
	}
}

func main() {
	args := os.Args[1:]

	if len(args) > 0 && args[0] == "index" {
		log.Fatal("lool")
		return
	}

	s1 := gocron.NewScheduler(time.UTC)
	s1.Every(1).Seconds().Do(task)
	s1.StartAsync()

	r := mux.NewRouter()

	r.HandleFunc("/api/link", getLinks).Methods(http.MethodGet)
	r.HandleFunc("/api/link", createLink).Methods(http.MethodPost)
	r.Use(tokenMiddleware)

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	log.Fatal(http.ListenAndServe("127.0.0.1:8080", handlers.CORS(originsOk, headersOk, methodsOk)(r)))

}
