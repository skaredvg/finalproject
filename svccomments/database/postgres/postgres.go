package postgres

import (
	"context"
	"fmt"
	"log"
	"skillfact/finalproject/svccomments/database"

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
func (dba *DB) New(p []database.Comment) error {
	sql := `INSERT INTO public.publication(
			title, annotation, publication_time, publication_url
			) VALUES ($1, $2, $3, $4)
			ON CONFLICT (title) DO UPDATE SET title = EXCLUDED.title
			RETURNING id`
	b := new(pgx.Batch)
	for _, ent := range p {
		b.Queue(sql, ent.Author, ent.Content, ent.NewsId, ent.ParentId)
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

// Функция возвращает публикации в количестве n
func (dba *DB) Last(n int) ([]database.Comment, error) {
	l := make([]database.Comment, 0)
	if n < 0 {
		return l, fmt.Errorf("Отрицательное число публикаций(%s)", n)
	}
	sql := `SELECT r.id, r.title, r.annotation, r.publication_time, r.publication_url FROM (
			 SELECT p.*, row_number() OVER () as rn
			 FROM postgres.public.publication p
			 ORDER BY p.publication_time DESC) r
			WHERE r.rn <= $1`
	rows, err := dba.con.Query(context.Background(), sql, n)
	if err != nil {
		return []database.Comment{}, err
	}
	for rows.Next() {
		p := database.Comment{}
		err := rows.Scan(&p.Id, &p.Author, &p.Content, &p.ParentId, &p.NewsId)
		if err != nil {
			return []database.Comment{}, err
		}
		l = append(l, p)
	}
	return l, nil
}

func (db *DB) CommentsOnNewsId(newsid int) ([]database.Comment, error) {
	l := make([]database.Comment, 0)
	if newsid < 0 {
		return l, fmt.Errorf("Отрицательное число публикаций(%s)", newsid)
	}
	sql := `SELECT id, newsid, author, parentid, content
			FROM postgres.public.comments p
			WHERE pnewsid = $1`
	rows, err := db.con.Query(context.Background(), sql, newsid)
	if err != nil {
		return []database.Comment{}, err
	}
	for rows.Next() {
		p := database.Comment{}
		err := rows.Scan(&p.Id, &p.NewsId, &p.Author, &p.ParentId, &p.Content)
		if err != nil {
			return []database.Comment{}, err
		}
		l = append(l, p)
	}
	return l, nil
}

func (db *DB) SetCommentOnNews(comm database.Comment, nid int) (int, error) {
	p := 0

	sql := `insert into postgres.public.comments p(
			newsid, author, parentid, content
			) VALUES ($1, $2, $3, $4)
			RETURNING id`
	row := db.con.QueryRow(context.Background(), sql, nid)
	err := row.Scan(&p)
	if err != nil {
		return 0, err
	}
	return p, nil
}
