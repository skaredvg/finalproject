package database

type DBInterface interface {
	CommentsOnNewsId(int) ([]Comment, error)
	SetCommentOnNews(Comment, int) (int, error)
}

type DBTestInterface interface {
	DBInterface
	AddTestingComments(int)
}

type Comment struct {
	Id       int64
	NewsId   int
	Author   string
	ParentId int64
	Content  string
}
