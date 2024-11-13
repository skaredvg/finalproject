package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"skillfact/finalproject/svcnews/database"
	"strconv"

	grlmux "github.com/gorilla/mux"
)

type NewsShortDetailed struct {
	Id              int64
	Title           string
	PublicationTime int64
}

type NewsFullDetailed struct {
	NewsShortDetailed
	LinkNews   string
	SiteNews   string
	Annotation string
}

type APIGate struct {
	mux *grlmux.Router
	db  database.DBTestInterface
}

func New(db database.DBTestInterface) *APIGate {
	g := new(APIGate)
	g.mux = grlmux.NewRouter()
	g.db = db
	return g
}

func (agate *APIGate) MiddleFunc(f http.HandlerFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Сервер: %s сквозной id= %d\n", r.RemoteAddr, 1)
		f.ServeHTTP(w, r)
	}
}

func (agate *APIGate) RegistryAPI() {
	agate.mux.HandleFunc("/news/latest", agate.MiddleFunc(agate.handlelatest)).Methods(http.MethodGet)
	agate.mux.HandleFunc("/news/pages/{page}", agate.MiddleFunc(agate.handlepage)).Methods(http.MethodGet)
	agate.mux.HandleFunc("/news/filter", agate.MiddleFunc(agate.handlefilter)).Methods(http.MethodGet)
	agate.mux.HandleFunc("/news/{id}", agate.MiddleFunc(agate.handlenewsview)).Methods(http.MethodGet)
}

func (agate *APIGate) Mux() *grlmux.Router {
	return agate.mux
}

func (agate *APIGate) GetDB() database.DBTestInterface {
	return agate.db
}

// Вернуть порцию последних новостей
// Параметры запроса:
// nc - размер порции
func (agate *APIGate) handlelatest(w http.ResponseWriter, r *http.Request) {
	nc := r.URL.Query().Get("nc")
	n, err := strconv.Atoi(nc)
	if err != nil || n < 1 {
		m := fmt.Sprintf("Некорректно указан параметр nc (%d)", n)
		http.Error(w, m, http.StatusBadRequest)
		return
	}

	w.Header().Add("content-type", "application/json")
	s, err := newslatest(agate, n)
	if err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNotFound)
		return
	}
	enc := json.NewEncoder(w)
	if err := enc.Encode(s); err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNotFound)
	}
}

func newslatest(agate *APIGate, n int) ([]NewsShortDetailed, error) {
	s, err := agate.db.NewsLatest(n)
	if err != nil {
		return []NewsShortDetailed{}, err
	}

	return sliceNewsShortDetailed(s), nil
}

func sliceNewsShortDetailed(s []database.NewsDetailed) []NewsShortDetailed {
	r := make([]NewsShortDetailed, 0, len(s))
	for _, v := range s {
		obj := NewsShortDetailed{}
		obj.Id = int64(v.Id)
		obj.Title = v.Title
		obj.PublicationTime = v.PublicationTime
		r = append(r, obj)
	}
	return r
}

// Вернуть заданную страницу новостей
// Параметры запроса:
// nc - размер окна новостей
// page - номер страницы новостей
func (agate *APIGate) handlepage(w http.ResponseWriter, r *http.Request) {
	nc := r.URL.Query().Get("nc")
	n, err := strconv.Atoi(nc)
	if err != nil || n < 1 {
		m := fmt.Sprintf("Некорректно указан параметр nc (%d)", n)
		http.Error(w, m, http.StatusBadRequest)
		return
	}

	Page := grlmux.Vars(r)["page"]
	page, err := strconv.Atoi(Page)
	if err != nil || page < 1 {
		m := fmt.Sprintf("Некорректно указан параметр page (%d)", page)
		http.Error(w, m, http.StatusBadRequest)
		return
	}

	w.Header().Add("content-type", "application/json")
	s, err := newspage(agate, page, n)
	if err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNotFound)
		return
	}
	enc := json.NewEncoder(w)
	if err := enc.Encode(s); err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNotFound)
		return
	}
}

func newspage(agate *APIGate, page int, n int) ([]NewsShortDetailed, error) {
	s, err := agate.db.NewsPage(n, page)
	if err != nil {
		return []NewsShortDetailed{}, err
	}
	return sliceNewsShortDetailed(s), nil
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
	if err != nil || bd < 0 {
		m := fmt.Sprintf("Некорректно указан параметр bpd (%d)", bd)
		w.Write([]byte(m))
	}

	epd := r.URL.Query().Get("epd")
	ed, err := strconv.Atoi(epd)
	if err != nil || ed < 0 {
		m := fmt.Sprintf("Некорректно указан параметр epd (%d)", ed)
		w.Write([]byte(m))
	}

	tl := r.URL.Query().Get("tl")

	nc := r.URL.Query().Get("nc")
	n, err := strconv.Atoi(nc)
	if err != nil || n < 1 {
		m := fmt.Sprintf("Некорректно указан параметр nc (%d)", n)
		w.Write([]byte(m))
	}

	f := database.Filter{DateFrom: bd,
		DateTo: ed,
		Title:  tl}
	w.Header().Add("content-type", "application/json")
	s, err := newsfilter(agate, n, f)
	if err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNotFound)
		return
	}
	if err := json.NewEncoder(w).Encode(s); err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNotFound)
		return
	}
}

func newsfilter(agate *APIGate, n int, f database.Filter) ([]NewsShortDetailed, error) {
	s, err := agate.db.NewsFilter(n, f)
	if err != nil {
		return []NewsShortDetailed{}, err
	}
	return sliceNewsShortDetailed(s), nil
}

// Вернуть информацию о публикации
func (agate *APIGate) handlenewsview(w http.ResponseWriter, r *http.Request) {
	Id := grlmux.Vars(r)["id"]
	id, err := strconv.Atoi(Id)
	if err != nil || id < 1 {
		m := fmt.Sprintf("Некорректно указан Id публикации (%s)", Id)
		w.Write([]byte(m))
	}

	w.Header().Add("content-type", "application/json")
	obj, err := newsfulldetailed(agate, id)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	enc := json.NewEncoder(w)
	err = enc.Encode(obj)
	if err != nil {
		fmt.Fprintf(os.Stdout, "Ответ %s", err.Error())
	}
}

func newsfulldetailed(agate *APIGate, id int) (*NewsFullDetailed, error) {
	nf := new(NewsFullDetailed)

	nd, err := agate.db.NewsDetailed(id)
	if err != nil {
		return nil, err
	}

	nf.Id = int64(nd.Id)
	nf.Title = nd.Title
	nf.PublicationTime = nd.PublicationTime
	nf.LinkNews = nd.LinkNews
	nf.SiteNews = nd.SiteNews
	nf.Annotation = nd.Annotation
	return nf, nil
}
