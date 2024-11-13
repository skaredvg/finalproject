package inmemory

import (
	"fmt"
	"skillfact/finalproject/svccomments/database"
)

type DB struct {
	Comments []database.Comment
	idcomm   int64
}

func NewDB(_ string) *DB {
	m := new(DB)
	m.Comments = make([]database.Comment, 0, 10)
	m.idcomm = 1
	return m
}

func (db *DB) CommentsOnNewsId(newsid int) ([]database.Comment, error) {
	cs := make([]database.Comment, 0, 10)
	for _, v := range db.Comments {
		if v.NewsId == newsid {
			cs = append(cs, v)
		}
	}
	return cs, nil
}

func (db *DB) SetCommentOnNews(comm database.Comment, nid int) (int, error) {
	comm.Id = db.idcomm
	db.idcomm++
	db.Comments = append(db.Comments, comm)
	return int(comm.Id), nil
}

func (db *DB) AddTestingComments(n int) {
	for k := range make([]any, 10) {
		if k%2 == 0 {
			nc := database.Comment{Id: db.idcomm,
				NewsId:   k + 1,
				Author:   fmt.Sprintf("Автор %d", k+1),
				ParentId: 0,
				Content:  "Просто комментарий",
			}
			db.idcomm++
			db.Comments = append(db.Comments, nc)
			for i := range make([]any, n) {
				nc := database.Comment{Id: db.idcomm,
					NewsId:   k + 1,
					Author:   fmt.Sprintf("Автор %d", i),
					ParentId: nc.Id,
					Content:  "Просто комментарий",
				}
				db.Comments = append(db.Comments, nc)
				db.idcomm++
			}
		}
	}
}
