package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
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

var cercaniasRegex = regexp.MustCompile("(rodalies)?(cercanias)?(aldiriko)?/i")

func main() {
	last, _ := strconv.ParseInt(os.Getenv("LAST_EXECUTION"), 10, 64)
	now := time.Now().Unix()
	fmt.Println("Now", now)
	fmt.Println("Last", last)
	fmt.Println("Now - last", now-last)
	if now-last > 60*10 {
		fmt.Println("now", strconv.Itoa(int(now)))
		fmt.Println("now", os.Getenv("LAST_EXECUTION"))
		err := os.Setenv("LAST_EXECUTION", strconv.Itoa(int(now)))
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

		/*fmt.Println(i.Alerts)
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
		}*/

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
			ref := cl.Collection("incidences").Doc(alert.ID)
			batch.Set(ref, alert)
		}

		_, err = batch.Commit(context.Background())
		if err != nil {
			fmt.Println("error!", err)
			return
		}
	} else {
		fmt.Println("Not enough seconds passed since last execution", now, last, now-last)
	}
}
