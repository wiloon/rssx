package utils

import (
	"net/http"

	"fmt"
	"io/ioutil"
)

func Download() {
	url := "http://www.oschina.net/news/rss"
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	//f, err := os.Create("osChina.xml")
	//if err != nil {
	//	panic(err)
	//}

	//io.Copy(f, res.Body)

	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		fmt.Print(bodyString)
	}

}
