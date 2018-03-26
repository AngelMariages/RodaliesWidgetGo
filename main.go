package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"time"

	firebase "firebase.google.com/go"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
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

//var cercaniasRegex = regexp.MustCompile("(rodalies)?(cercanias)?(aldiriko)?/i")

func main() {
	//doIncidencesRequest()
	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, os.Interrupt)

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("Running update")
			//go doIncidencesRequest()
		case <-ctrlC:
			ticker.Stop()
			fmt.Println("Finished from Ctrl + C")
			return
		}
	}
}

func doIncidencesRequest() {
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

	var i incidences
	err = xml.Unmarshal(body, &i)
	if err != nil {
		fmt.Println("error!", err)
		return
	}

	conf, err := google.JWTConfigFromJSON([]byte(os.Getenv("FIREBASE_CONFIG")),
		"https://www.googleapis.com/auth/cloud-platform",
		"https://www.googleapis.com/auth/datastore",
		"https://www.googleapis.com/auth/devstorage.full_control",
		"https://www.googleapis.com/auth/firebase",
		"https://www.googleapis.com/auth/identitytoolkit",
		"https://www.googleapis.com/auth/userinfo.email")
	ts := conf.TokenSource(context.Background())

	app, err := firebase.NewApp(context.Background(), &firebase.Config{ProjectID: os.Getenv("FIREBASE_PROJECT_ID")}, option.WithTokenSource(ts))
	if err != nil {
		fmt.Println("error!", err)
		return
	}

	cl, err := app.Firestore(context.Background())
	if err != nil {
		fmt.Println("error!", err)
		return
	}
	defer cl.Close()

	batch := cl.Batch()

	for _, alert := range i.Alerts {
		ref := cl.Collection("incidences/" + alert.CA).Doc(alert.ID)
		batch.Set(ref, alert)
	}

	_, err = batch.Commit(context.Background())
	if err != nil {
		fmt.Println("error!", err)
		return
	}
}
