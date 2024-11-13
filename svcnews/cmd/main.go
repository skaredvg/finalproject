package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"skillfact/finalproject/svcnews/api"
	"skillfact/finalproject/svcnews/database"
	"skillfact/finalproject/svcnews/database/inmemory"
	"strings"
	"sync"
	"time"
)

// Структура хранения конфигурации
type config struct {
	TimePeriodScan uint
	Sites          []struct {
		Url string
	}
	DBUser     string
	DBPassword string
}

// Загрузка конфигурации
func load_conf(fn string) config {
	f, err := os.OpenFile(fn, os.O_RDONLY, 0111)
	if err != nil {
		log.Fatalf("Не найден файл конфигурации%s", fn)
	}
	b, err := io.ReadAll(f)
	if err != nil {
		log.Fatalf("Ошибка чтения конфигурации %s", fn)
	}

	cfg := config{DBUser: "postgres", DBPassword: "06041972"}
	if json.Unmarshal(b, &cfg) != nil {
		log.Fatalf("Ошибка чтения конфигурации %s", fn)
	}

	//fmt.Printf("%v", cfg)
	return cfg
}

// Обработка RSS-ссылки
func processRSSLinksChan(l string, tps uint, chdb chan<- []api.RSSPublication, chlog chan<- error, w *sync.WaitGroup) {
	defer w.Done()
	for {
		rnf := api.NewRSSNewsFeed(l)
		err := rnf.ProcessLink()
		if err != nil {
			chlog <- err
		}
		chdb <- rnf.Channel.Publications
		<-time.After(time.Duration(tps) * time.Second)
	}
}

// Сохранение постов в БД
func processRSSPostToDatabase(db database.DBInterface, chdb <-chan []api.RSSPublication, chlog chan<- error, w *sync.WaitGroup) {
	defer w.Done()
	for v := range chdb {
		dbp := []database.NewsDetailed{}
		for _, ent := range v {
			p := database.NewsDetailed{
				Title:      ent.Title,
				Annotation: ent.Description,
				LinkNews:   ent.Link,
				SiteNews:   "",
			}

			t := time.Now()

			if strings.Contains(ent.PubTime, "GMT") {
				t, _ = time.Parse(time.RFC1123, ent.PubTime)
			} else {
				t, _ = time.Parse(time.RFC1123Z, ent.PubTime)
			}
			p.PublicationTime = t.UnixMilli()
			dbp = append(dbp, p)
		}
		if len(dbp) == 0 {
			continue
		}
		err := db.New(dbp)
		if err != nil {
			chlog <- err
		}
	}
}

// Вывод ошибок в консоль
func processLog(cherr <-chan error) {
	for err := range cherr {
		log.Println(err.Error())
	}
}

func main() {
	cfg := load_conf("config.json")
	//conn := fmt.Sprintf("postgres://%s:%s@%s/postgres", cfg.DBUser, cfg.DBPassword, "localhost")
	conn := ""

	db := inmemory.NewDB(conn)
	db.AddTestingNews(10)

	/*------------------------*/
	chdb := make(chan []api.RSSPublication)
	cherr := make(chan error)
	w := new(sync.WaitGroup)
	go processLog(cherr)
	for _, v := range cfg.Sites {
		w.Add(1)
		go processRSSLinksChan(v.Url, cfg.TimePeriodScan, chdb, cherr, w)
	}
	w.Add(1)
	go processRSSPostToDatabase(db, chdb, cherr, w)
	/*-------------------------*/

	api := api.New(db)
	api.RegistryAPI()
	mux := api.Mux()
	fmt.Println(1)
	fmt.Println(http.ListenAndServe("127.0.0.1:8086", mux))
	fmt.Println(2)
}
