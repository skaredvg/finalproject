package inmemory

import (
	"fmt"
	"skillfact/finalproject/svcnews/database"
	"strings"
	"time"
)

type DB struct {
	News   []database.NewsDetailed
	idnews int
	uq     map[string]int // для устранения дублирования публикаций

}

func NewDB(_ string) *DB {
	m := new(DB)
	m.News = make([]database.NewsDetailed, 0, 10)
	m.idnews = 1
	m.uq = map[string]int{}
	return m
}

func (db *DB) NewsLatest(n int) ([]database.NewsDetailed, error) {
	s := make([]database.NewsDetailed, 0, n)
	count := 0
	for i := len(db.News) - 1; i >= 0; i-- {
		if count >= n {
			break
		}
		s = append(s, db.News[i])
		count++
	}
	return s, nil
}

func (db *DB) NewsPage(n int, page int) ([]database.NewsDetailed, error) {
	s := make([]database.NewsDetailed, 0, n)
	l := len(db.News)
	j := l - (page-1)*n - 1
	count := 0
	for i := j; i >= 0; i-- {
		if count >= n {
			break
		}
		s = append(s, db.News[i])
		count++
	}
	return s, nil
}

func (db *DB) NewsFilter(n int, f database.Filter) ([]database.NewsDetailed, error) {
	bd := int64(f.DateFrom)
	ed := int64(f.DateTo)
	title := f.Title

	s := make([]database.NewsDetailed, 0, n)
	count := 0
	for i := len(db.News) - 1; i >= 0; i-- {
		if count >= n {
			break
		}
		if bd == 0 {
			bd = db.News[i].PublicationTime
		}
		if ed == 0 {
			ed = db.News[i].PublicationTime
		}
		if title == "" {
			title = db.News[i].Title
		}
		if db.News[i].PublicationTime >= bd && db.News[i].PublicationTime <= ed && strings.Contains(db.News[i].Title, title) {
			s = append(s, db.News[i])
			count++
		}
	}
	return s, nil
}

func (db *DB) NewsDetailed(id int) (*database.NewsDetailed, error) {
	news := new(database.NewsDetailed)
	for _, v := range db.News {
		if v.Id == id {
			news.Id = v.Id
			news.Title = v.Title
			news.PublicationTime = v.PublicationTime
			news.LinkNews = v.LinkNews
			news.SiteNews = v.SiteNews
			news.Annotation = v.Annotation
		}
	}
	if news.Id == 0 {
		return nil, fmt.Errorf("Публикация id=%d не найдена", id)
	}
	return news, nil
}

// Функция регистрации масива публикаций в БД
func (dba *DB) New(p []database.NewsDetailed) error {
	for _, ent := range p {
		if _, ok := dba.uq[ent.Title]; !ok {
			ent.Id = dba.idnews
			dba.News = append(dba.News, ent)
			dba.uq[ent.Title] = ent.Id
			dba.idnews++
		}
	}

	return nil
}

func (db *DB) AddTestingNews(n int) {
	for range make([]any, n) {
		n := database.NewsDetailed{Id: db.idnews,
			Title:           fmt.Sprintf("Публикация %d", db.idnews),
			PublicationTime: time.Now().UnixNano(),
			LinkNews:        fmt.Sprintf("localhost/%d", db.idnews),
			SiteNews:        "localhost",
			Annotation:      fmt.Sprintf("Аннотация к публикации %d", db.idnews)}
		db.idnews++
		db.News = append(db.News, n)
	}
}
