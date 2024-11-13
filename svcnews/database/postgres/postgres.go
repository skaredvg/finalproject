package postgres

import (
	"context"
	"fmt"
	"log"
	"math"
	"skillfact/finalproject/svcnews/database"

	"github.com/jackc/pgx/v4"
)

// Структура - объект соединения с БД
type DB struct {
	con *pgx.Conn
}

// Конструктор объекта соединения с БД
func NewDB(pconn string) (*DB, error) {
	db, err := pgx.Connect(context.Background(), pconn)
	if err != nil {
		log.Fatal(err.Error())
		return nil, err
	}
	return &DB{con: db}, err
}

// Функция регистрации масива публикаций в БД
func (dba *DB) New(p []database.NewsDetailed) error {
	sql := `INSERT INTO public.publication(
			title, annotation, publication_time, publication_url
			) VALUES ($1, $2, $3, $4)
			ON CONFLICT (title) DO UPDATE SET title = EXCLUDED.title
			RETURNING id`
	b := new(pgx.Batch)
	for _, ent := range p {
		b.Queue(sql, ent.Title, ent.Annotation, ent.PublicationTime, ent.LinkNews)
	}
	if b.Len() > 0 {
		br := dba.con.SendBatch(context.Background(), b)
		if _, err := br.Query(); err != nil {
			fmt.Println(err)
			return err
		}
		br.Close()
	}

	return nil
}

func (db *DB) NewsLatest(n int) ([]database.NewsDetailed, error) {
	l := make([]database.NewsDetailed, 0)
	if n < 0 {
		return l, fmt.Errorf("Отрицательное число публикаций(%s)", n)
	}
	sql := `SELECT r.id, r.title, r.annotation, r.publication_time, r.publication_url FROM (
			 SELECT p.*, row_number() OVER () as rn
			 FROM postgres.public.publication p
			 ORDER BY p.publication_time DESC) r
			WHERE r.rn <= $1`
	rows, err := db.con.Query(context.Background(), sql, n)
	if err != nil {
		return []database.NewsDetailed{}, err
	}
	for rows.Next() {
		p := database.NewsDetailed{}
		err := rows.Scan(&p.Id, &p.Title, &p.Annotation, &p.PublicationTime, &p.LinkNews)
		if err != nil {
			return []database.NewsDetailed{}, err
		}
		l = append(l, p)
	}
	return l, nil
}

func (db *DB) NewsPage(n int, page int) ([]database.NewsDetailed, error) {
	l := make([]database.NewsDetailed, 0)
	if n < 0 {
		return l, fmt.Errorf("Отрицательное число публикаций(%s)", n)
	}
	sqlc := `SELECT COUNT(1) FROM postgres.public.publication p RETURNING $1`
	row := db.con.QueryRow(context.Background(), sqlc, n)
	cn := 0
	row.Scan(&cn)

	sql := `SELECT r.id, r.title, r.annotation, r.publication_time, r.publication_url FROM (
			 SELECT p.*, row_number() OVER () as rn
			 FROM postgres.public.publication p
			 ORDER BY p.publication_time DESC) r
			WHERE r.rn between $1 AND $2`

	ne := n * page
	nb := ne - n + 1

	rows, err := db.con.Query(context.Background(), sql, nb, ne)
	if err != nil {
		return []database.NewsDetailed{}, err
	}
	for rows.Next() {
		p := database.NewsDetailed{}
		err := rows.Scan(&p.Id, &p.Title, &p.Annotation, &p.PublicationTime, &p.LinkNews)
		if err != nil {
			return []database.NewsDetailed{}, err
		}
		l = append(l, p)
	}
	return l, nil
}

func (db *DB) NewsFilter(n int, f database.Filter) ([]database.NewsDetailed, error) {
	bd := int64(f.DateFrom)
	ed := int64(f.DateTo)
	title := f.Title

	l := make([]database.NewsDetailed, 0, n)

	sql := `SELECT r.id, r.title, r.annotation, r.publication_time, r.publication_url, r.rn
			FROM (SELECT p.*, row_number() OVER () as rn FROM (
			SELECT p.*
			FROM postgres.public.publication p
			ORDER BY p.publication_time DESC) p
	  		) r
			WHERE r.rn <= $1 AND r.publication_time between $2 AND $3 AND r.title ILIKE "%$4%"`

	if ed == 0 {
		ed = math.MaxInt64
	}
	rows, err := db.con.Query(context.Background(), sql, n, bd, ed, title)
	if err != nil {
		return []database.NewsDetailed{}, err
	}
	for rows.Next() {
		p := database.NewsDetailed{}
		err := rows.Scan(&p.Id, &p.Title, &p.Annotation, &p.PublicationTime, &p.LinkNews)
		if err != nil {
			return []database.NewsDetailed{}, err
		}
		l = append(l, p)
	}
	return l, nil
}

func (db *DB) NewsDetailed(id int) (*database.NewsDetailed, error) {
	news := new(database.NewsDetailed)

	sql := `SELECT r.id, r.title, r.annotation, r.publication_time, r.publication_url
			FROM postgres.public.publication p
			WHERE r.id <= $1`

	row := db.con.QueryRow(context.Background(), sql, id)
	err := row.Scan(&news.Id, &news.Title, &news.Annotation, &news.PublicationTime, &news.LinkNews)
	if err != nil {
		return &database.NewsDetailed{}, err
	}
	return news, nil
}
