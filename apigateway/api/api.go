package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	grlmux "github.com/gorilla/mux"
)

type NewsShortDetailed struct {
	Id              int64
	Title           string
	PublicationTime int64
	LinkNews        string
	SiteNews        string
	Annotation      string
}

type NewsFullDetailed struct {
	NewsShortDetailed
	Comments []Comment
}

type Comment struct {
	Id       int64
	NewsId   int64
	Author   string
	ParentId int64
	Content  string
	Comments []Comment
}

type APIGate struct {
	grlmux *grlmux.Router
	client *http.Client
	lnk    map[string]string
}

func New(l map[string]string, client *http.Client) *APIGate {
	g := new(APIGate)
	if client == nil {
		client = http.DefaultClient
	}
	g.grlmux = grlmux.NewRouter()
	g.client = client
	g.lnk = l
	return g
}

func (agate *APIGate) MiddleFunc(f http.HandlerFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Сервер: %s сквозной id= %d\n", r.RemoteAddr, 1)
		f.ServeHTTP(w, r)
	}
}

func (agate *APIGate) RegistryAPI() {
	agate.grlmux.HandleFunc("/news/latest", agate.MiddleFunc(agate.handlelatest)).Methods(http.MethodGet)
	agate.grlmux.HandleFunc("/news", agate.MiddleFunc(agate.handlepage)).Methods(http.MethodGet)
	agate.grlmux.HandleFunc("/news/filter", agate.MiddleFunc(agate.handlefilter)).Methods(http.MethodGet)
	agate.grlmux.HandleFunc("/news/{id}", agate.MiddleFunc(agate.handlenewsview)).Methods(http.MethodGet)
	agate.grlmux.HandleFunc("/comments/news/{id}", agate.MiddleFunc(agate.handlecommentsonnews)).Methods(http.MethodGet)
	agate.grlmux.HandleFunc("/comments/news/{id}", agate.MiddleFunc(agate.handleaddcommentsonnews)).Methods(http.MethodPost)
}

func (agate *APIGate) Mux() *grlmux.Router {
	return agate.grlmux
}

// Вернуть порцию последних новостей
// Параметры запроса:
// nc - размер порции
func (agate *APIGate) handlelatest(w http.ResponseWriter, r *http.Request) {
	nc := r.URL.Query().Get("nc")
	n, err := strconv.Atoi(nc)
	if err != nil {
		m := fmt.Sprintf("Некорректно указан параметр nc (%d)", n)
		http.Error(w, m, http.StatusBadRequest)
		return
	}

	resp, err := agate.client.Get(agate.lnk["svcnews"] + r.RequestURI)
	if err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNotFound)
		return
	}

	if ct := resp.Header.Get("content-type"); ct != "" {
		w.Header().Add("content-type", ct)
	}

	b, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNotFound)
		return
	}

	w.Write(b)
}

// Вернуть заданную страницу новостей
// Параметры запроса:
// nc - размер окна новостей
// page - номер страницы новостей
func (agate *APIGate) handlepage(w http.ResponseWriter, r *http.Request) {
	nc := r.URL.Query().Get("nc")
	n, err := strconv.Atoi(nc)
	if err != nil {
		m := fmt.Sprintf("Некорректно указан параметр nc (%d)", n)
		http.Error(w, m, http.StatusBadRequest)
		return
	}

	pg := r.URL.Query().Get("page")
	if pg == "" {
		pg = "1"
	}
	page, err := strconv.Atoi(pg)
	if err != nil {
		m := fmt.Sprintf("Некорректно указан параметр page (%d)", page)
		http.Error(w, m, http.StatusBadRequest)
		return
	}

	resp, err := agate.client.Get(agate.lnk["svcnews"] + r.RequestURI)
	if err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNotFound)
		return
	}

	if ct := resp.Header.Get("content-type"); ct != "" {
		w.Header().Add("content-type", ct)
	}

	b, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNotFound)
		return
	}

	w.Write(b)
}

// Вернуть результаты фильтра
// Параметры запроса:
// bpd - (begin publication date) начало времени публикации
// epd - (end publication date) конец времени публикации
// tl - (title) подстрока в заголовке публикации
// nc - размер порции данных
func (agate *APIGate) handlefilter(w http.ResponseWriter, r *http.Request) {
	bpd := r.URL.Query().Get("bpd")
	bd, err := strconv.Atoi(bpd)
	if err != nil {
		m := fmt.Sprintf("Некорректно указан параметр bpd (%d)", bd)
		http.Error(w, m, http.StatusBadRequest)
	}

	epd := r.URL.Query().Get("epd")
	ed, err := strconv.Atoi(epd)
	if err != nil {
		m := fmt.Sprintf("Некорректно указан параметр epd (%d)", ed)
		http.Error(w, m, http.StatusBadRequest)
	}

	nc := r.URL.Query().Get("nc")
	n, err := strconv.Atoi(nc)
	if err != nil {
		m := fmt.Sprintf("Некорректно указан параметр nc (%d)", n)
		http.Error(w, m, http.StatusBadRequest)
	}

	resp, err := agate.client.Get(agate.lnk["svcnews"] + r.RequestURI)
	if err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNotFound)
		return
	}

	if ct := resp.Header.Get("content-type"); ct != "" {
		w.Header().Add("content-type", ct)
	}

	b, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNotFound)
		return
	}

	w.Write(b)
}

// Вернуть клиенту полную информацию о публикации
func (agate *APIGate) handlenewsview(w http.ResponseWriter, r *http.Request) {
	Id := grlmux.Vars(r)["id"]
	_, err := strconv.Atoi(Id)
	if err != nil {
		m := fmt.Sprintf("Некорректно указан Id публикации (%s)", Id)
		http.Error(w, m, http.StatusBadRequest)
	}

	resp, err := agate.client.Get(agate.lnk["svcnews"] + r.RequestURI)
	if err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNoContent)
		return
	}

	if ct := resp.Header.Get("content-type"); ct != "" {
		w.Header().Add("content-type", ct)
	}

	bnews, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNoContent)
		return
	}

	news := NewsFullDetailed{Comments: []Comment{}}
	dec := json.NewDecoder(bytes.NewReader(bnews))
	if err := dec.Decode(&news); err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNoContent)
		return
	}

	resp, err = agate.client.Get(fmt.Sprintf("%s/comments/news/%s", agate.lnk["svccomments"], Id))
	if err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNoContent)
		return
	}

	bcom, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNoContent)
		return
	}

	comms := []Comment{}
	dec = json.NewDecoder(bytes.NewReader(bcom))
	if err = dec.Decode(&comms); err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNoContent)
		return
	}

	news.Comments = comms
	enc := json.NewEncoder(w)
	if err = enc.Encode(news); err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNoContent)
		return
	}
}

func (agate *APIGate) handlecommentsonnews(w http.ResponseWriter, r *http.Request) {
	Id := grlmux.Vars(r)["id"]
	_, err := strconv.Atoi(Id)
	if err != nil {
		m := fmt.Sprintf("Некорректно указан Id публикации (%s)", Id)
		http.Error(w, m, http.StatusBadRequest)
	}

	resp, err := agate.client.Get(agate.lnk["svccomments"] + r.RequestURI)
	if err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNotFound)
		return
	}

	if ct := resp.Header.Get("content-type"); ct != "" {
		w.Header().Add("content-type", ct)
	}

	b, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNotFound)
		return
	}

	w.Write(b)
}

func (agate *APIGate) handleaddcommentsonnews(w http.ResponseWriter, r *http.Request) {
	Id := grlmux.Vars(r)["id"]
	_, err := strconv.Atoi(Id)
	if err != nil {
		m := fmt.Sprintf("Некорректно указан Id публикации (%s)", Id)
		http.Error(w, m, http.StatusNotFound)
	}

	ct := r.Header.Get("content-type")
	if ct == "" {
		ct = "plain/text"
	}

	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	resp, err := agate.client.Post(agate.lnk["svccomments"]+r.RequestURI, ct, bytes.NewReader(b))
	if err != nil {
		m := fmt.Sprintf("Ошибка добавления данных (%s)", err.Error())
		http.Error(w, m, http.StatusNotFound)
		return
	}

	resp.Write(w)
}
