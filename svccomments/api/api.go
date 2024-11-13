package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"skillfact/finalproject/svccomments/database"
	"strconv"
	"strings"

	grlmux "github.com/gorilla/mux"
)

type Comment struct {
	Id        int64
	NewsId    int64
	Author    string
	ParentId  int64
	Content   string
	Bad       bool
	BadReason string
	Comments  []Comment
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
	agate.mux.HandleFunc("/comments/news/{id}", agate.MiddleFunc(agate.handlecommentsonnews)).Methods(http.MethodGet)
	agate.mux.HandleFunc("/comments/news/{id}", agate.MiddleFunc(agate.handleaddcommentsonnews)).Methods(http.MethodPost)
}

func (agate *APIGate) Mux() *grlmux.Router {
	return agate.mux
}

func (agate *APIGate) GetDB() database.DBTestInterface {
	return agate.db
}

func deepConvertComments(src []database.Comment, pid int) []Comment {
	dest := make([]Comment, 0, 10)
	for _, v := range src {
		if v.ParentId == int64(pid) {
			note := Comment{}
			note.Id = v.Id
			note.NewsId = int64(v.NewsId)
			note.Author = v.Author
			note.ParentId = v.ParentId
			note.Content = v.Content
			note.Comments = deepConvertComments(src, int(v.Id))
			dest = append(dest, note)
		}
	}
	return dest
}

func (agate *APIGate) handlecommentsonnews(w http.ResponseWriter, r *http.Request) {
	Id := grlmux.Vars(r)["id"]
	id, err := strconv.Atoi(Id)
	if err != nil {
		m := fmt.Sprintf("Некорректно указан параметр id (%s)", Id)
		http.Error(w, m, http.StatusBadRequest)
		return
	}
	w.Header().Add("content-type", "application/json")
	obj, err := commentsonnews(agate, id)
	fmt.Fprintln(os.Stdout, obj)
	if err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNotFound)
		return
	}
	enc := json.NewEncoder(w)
	if err = enc.Encode(obj); err != nil {
		m := fmt.Sprintf("Ошибка получения данных (%s)", err.Error())
		http.Error(w, m, http.StatusNotFound)
		return
	}
}

func commentsonnews(agate *APIGate, id int) ([]Comment, error) {
	comms, err := agate.db.CommentsOnNewsId(id)
	if err != nil {
		return []Comment{}, err
	}
	return deepConvertComments(comms, 0), nil
}

func validateComment(in <-chan *Comment) <-chan any {
	out := make(chan any)
	go func() {
		defer close(out)
		c, ok := <-in
		if !ok {
			return
		}
		if strings.Contains(c.Content, "йцукен") {
			c.Bad = true
			c.BadReason = fmt.Sprintf("Запрещенная подстрока %s", "йцукен")
		}
	}()
	return out
}

func (agate *APIGate) handleaddcommentsonnews(w http.ResponseWriter, r *http.Request) {
	Id := grlmux.Vars(r)["id"]
	id, err := strconv.Atoi(Id)
	if err != nil {
		m := fmt.Sprintf("Некорректно указан параметр id (%s)", Id)
		http.Error(w, m, http.StatusBadRequest)
		return
	}

	b, err := io.ReadAll(r.Body.(io.Reader))
	defer r.Body.Close()
	if err != nil {
		m := fmt.Sprintf("Ошибка добавления комментариев к публикации id=%d (%s)", id, err.Error())
		http.Error(w, m, http.StatusInternalServerError)
		return
	}

	sr := bytes.Runes(b[:10])

	var (
		comm  *Comment
		comms []*Comment
		fOne  bool = true
	)

	if sr[0] == '[' {
		comms = []*Comment{}
		fOne = false
	} else {
		comm = new(Comment)
		fOne = true
	}

	dec := json.NewDecoder(bytes.NewReader(b))
	if fOne {
		err = dec.Decode(comm)
	} else {
		err = dec.Decode(&comms)
	}

	if err != nil {
		m := fmt.Sprintf("Ошибка добавления комментария(ев) к публикации id=%d (%s)", id, err.Error())
		http.Error(w, m, http.StatusInternalServerError)
		return
	}

	err = nil
	if fOne {
		ch := make(chan *Comment)
		ch2 := validateComment(ch)
		ch <- comm
		<-ch2
		if comm.Bad {
			err = errors.New("Не пройдена валидация")
		} else {
			err = addcommentonnews(agate, comm, id)
		}
	} else {
		ch := make(chan *Comment, len(comms))
		for _, v := range comms {
			ch2 := validateComment(ch)
			ch <- v
			<-ch2
			if v.Bad {
				err = errors.New("Не пройдена валидация")
			} else {
				err = addcommentonnews(agate, v, id)
			}
			break
		}
	}
	if err != nil {
		m := fmt.Sprintf("Ошибка добавления комментариев к публикации id=%d (%s)", id, err.Error())
		http.Error(w, m, http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)
	if fOne {
		err = enc.Encode(comm)
	} else {
		err = enc.Encode(comms)
	}
	if err != nil {
		m := fmt.Sprintf("Ошибка добавления комментариев к публикации id=%d (%s)", id, err.Error())
		http.Error(w, m, http.StatusInternalServerError)
		return
	}
}

func addcommentonnews(agate *APIGate, comm *Comment, id int) error {
	commdb := database.Comment{}
	commdb.NewsId = id
	commdb.ParentId = comm.ParentId
	commdb.Author = comm.Author
	commdb.Content = comm.Content
	cid, err := agate.db.SetCommentOnNews(commdb, commdb.NewsId)
	comm.Id = int64(cid)
	return err
}
