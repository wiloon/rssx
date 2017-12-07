package rss

import (
	"net/http"
	"io/ioutil"
	"fmt"
	"encoding/xml"
	"wiloon.com/rssx/data"
	"github.com/wiloon/wiloon-log/log"
)

func Sync() {
	//find all feeds
	feeds := data.FindFeeds()
	for _, feed := range feeds {
		res, err := http.Get(feed.Url)
		if err != nil {
			panic(err)
		}

		defer res.Body.Close()

		var remoteFeedBody []byte
		if res.StatusCode == http.StatusOK {
			remoteFeedBody, _ = ioutil.ReadAll(res.Body)
			bodyString := string(remoteFeedBody)
			fmt.Print(bodyString)
		}

		v := Rss{}
		err = xml.Unmarshal(remoteFeedBody, &v)
		if err != nil {
			fmt.Printf("error: %v", err)
			return
		}

		log.Info(v)
		log.Info("xml name:", v.XMLName)
		log.Info("Version:", v.Version)
		log.Info("Description:", v.Description)
		log.Info("channel:", v.Chan)
		log.Info("channel.title:", v.Chan.Title)
		log.Info("channel.link:", v.Chan.Link)
		log.Info("channel.items:", v.Chan.Items[0].Title)
		log.Info("channel.items:", v.Chan.Items[1].Title)

		for i, v := range v.Chan.Items {
			log.Info("index:%v, title:%v", i, v.Title)

			data.SaveNews(feed.Id, v.Title, v.Link, v.Description)

		}
	}

}
