package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"

	"google.golang.org/api/option"

	"cloud.google.com/go/firestore"
)

type incidences struct {
	Alerts []alert `xml:"aviso"`
}

type alert struct {
	ID       string `xml:"id"`
	Date     string `xml:"publicacion"`
	CA       string `xml:"categoria"`
	Affects  string `xml:"area"`
	Title    string `xml:"titular"`
	Subtitle string `xml:"entradilla"`
	Text     string `xml:"texto"`
	Sent     bool
}

// Aldiriko, Rodalies, Cercanias

var cercaniasRegex = regexp.MustCompile("(rodalies)?(cercanias)?(aldiriko)?/i")

func main() {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	res, err := client.Get("http://web02.renfe.es/u13/MTR/UltimaHora.nsf/xmlAvisosApp")

	if err != nil {
		fmt.Println("error!", err)
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		fmt.Println("error!", err)
		return
	}

	//fmt.Println("Read:\n", string(body))

	var i incidences
	err = xml.Unmarshal(body, &i)
	if err != nil {
		fmt.Println("error!", err)
		return
	}

	fmt.Println(i.Alerts)
	for _, alert := range i.Alerts {
		fmt.Println(alert.Date)
		fmt.Println(alert.CA)
		fmt.Println(alert.Affects)
		fmt.Println(alert.Title)
		fmt.Println(alert.Subtitle)
		fmt.Println(alert.Text)
		fmt.Println()
		fmt.Println()
		fmt.Println()
	}

	/*conf := &firebase.Config{ProjectID: "rodalieswidget"}
	app, err := firebase.NewApp(context.Background(), nil, option.WithAPIKey(os.Getenv("FIREBASE_API_KEY")))
	if err != nil {
		fmt.Println("error!", err)
		return
	}*/

	cl, err := firestore.NewClient(context.Background(), "rodalieswidget", option.WithAPIKey(os.Getenv("FIREBASE_API_KEY")))
	if err != nil {
		fmt.Println("error!", err)
		return
	}
	defer cl.Close()

	batch := cl.Batch()

	for _, alert := range i.Alerts {
		ref := cl.Collection("incidences").Doc(alert.ID)
		batch.Set(ref, alert)
	}

	_, err = batch.Commit(context.Background())
	if err != nil {
		fmt.Println("error!", err)
		return
	}
}
