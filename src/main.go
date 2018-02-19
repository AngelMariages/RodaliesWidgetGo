package main

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type incidence struct {
	comunidad string
	category  string
	alert     string
}

// Aldiriko, Rodalies, Cercanias

func main() {
	alerts := make(map[string][]*incidence)
	doc, err := goquery.NewDocument("http://web02.renfe.es/u13/MTR/UltimaHora.nsf/Carga%20Vista%203A?OpenAgent")

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	doc.Find(".cuerpoavisos ul").Each(func(i int, node *goquery.Selection) {
		node.Find(".comunidad").Each(func(i int, comunity *goquery.Selection) {
			com := comunity.Text()
			node.Find(".avisos a").Each(func(i int, aviso *goquery.Selection) {
				cat := aviso.ChildrenFiltered(".str").Text()
				aviso.Children().Remove()
				alert := strings.TrimSpace(aviso.Text())
				alerts[com] = append(alerts[com], &incidence{
					com,
					cat,
					alert,
				})
			})
		})
	})

	for _, alert := range alerts {
		for _, inc := range alert {
			fmt.Println(inc)
		}
		fmt.Println()
		fmt.Println()
	}
}
