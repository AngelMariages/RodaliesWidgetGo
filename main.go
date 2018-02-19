package main

import (
	"log"
	"time"

	"github.com/mmcdole/gofeed"
)

type incidence struct {
	title       string
	description string
	published   time.Time
}

func main() {
	var incidences []incidence

	loc, _ := time.LoadLocation("Europe/Spain")

	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL("http://www.gencat.cat/rodalies/incidencies_rodalies_rss_ca_ES.xml")
	for _, item := range feed.Items {
		if item.Title != "Normalitat al servei." {
			t, err := time.ParseInLocation("Mon, 2 Jan 2006 15:04:05 -0700", item.Published, loc)
			if err != nil {
				log.Fatal("Couldn't parse time", err)
				return
			}

			incidences = append(incidences, incidence{item.Title, item.Description, t})
		}
	}

	for _, incidence := range incidences {
		log.Println(incidence.title)
		log.Println(incidence.description)
		log.Println(incidence.published)
	}
}
