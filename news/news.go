package news

func init() {

}

var bucket = "NewsBucket"

const (
	NewsId      = "Id"
	FeedId      = "FeedId"
	Title       = "Title"
	Url         = "Url"
	Description = "Description"
	NextNewsId  = "NextId"
	PubDate     = "PubDate"
	Guid        = "Guid"
)

type Site struct {
	Title    string
	NewsList []News
}
type News struct {
	Id          string
	FeedId      int64
	Title       string
	Url         string
	Description string
	NextId      string
	PubDate     string
	Guid        string
}

func (site *Site) Append(title, url, description string) {
	site.NewsList = append(site.NewsList, News{Title: title, Url: url, Description: description})
}

//func (site *Site) Save() {
//
//	for _, v := range site.NewsList {
//		data.Save(v.Title, v.Url, v.Description)
//	}
//
//}
