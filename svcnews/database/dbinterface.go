package database

type DBInterface interface {
	NewsLatest(int) ([]NewsDetailed, error)
	NewsPage(int, int) ([]NewsDetailed, error)
	NewsFilter(int, Filter) ([]NewsDetailed, error)
	NewsDetailed(int) (*NewsDetailed, error)
	New([]NewsDetailed) error
}

type DBTestInterface interface {
	DBInterface
	AddTestingNews(int)
}

type NewsDetailed struct {
	Id              int
	Title           string
	PublicationTime int64
	LinkNews        string
	SiteNews        string
	Annotation      string
}

type Filter struct {
	Title    string
	DateFrom int
	DateTo   int
}
