package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	usbdrivedetector "github.com/deepakjois/gousbdrivedetector"
	"github.com/ghodss/yaml"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type Info struct {
	Email string `yaml:"email"`
	Path  string `yaml:"path"`
}

type Config []Info

func (k Config) getPath(from string) string {
	for _, kk := range k {
		if kk.Email == from {
			return kk.Path
		}
	}
	return ""
}

func detectusb() {
	for range time.After(10 * time.Second) {
		if drives, err := usbdrivedetector.Detect(); err == nil {
			fmt.Printf("%d USB Devices Found\n", len(drives))
			for _, d := range drives {
				fmt.Println(d)
			}
		} else {
			fmt.Println(err)
		}
	}
}

func main() {

	go detectusb()

	configData, err := os.ReadFile("config.yaml")
	if err != nil {
		fmt.Println("Failed to read config file")
		return
	}

	var c Config
	err = yaml.Unmarshal(configData, &c)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	ctx := context.Background()
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	fromClause := ""
	for _, cc := range c {
		fromClause += cc.Email + ","
	}
	fromClause = strings.TrimSuffix(fromClause, ",")

	// process 2 days message
	query := fmt.Sprintf("from:%s,label:inbox", fromClause)
	req := srv.Users.Messages.List("me").Q(query)
	// if pageToken != "" {
	// 	req.PageToken(pageToken)
	// }
	req.MaxResults(50)
	r, err := req.Do()
	if err != nil {
		log.Fatalf("Unable to retrieve messages: %v", err)
	}

	log.Printf("Processing %v messages...\n", len(r.Messages))
	for _, m := range r.Messages {
		msg, err := srv.Users.Messages.Get("me", m.Id).Do()
		if err != nil {
			log.Fatalf("Unable to retrieve message %v: %v", m.Id, err)
		}

		from := ""
		date := ""
		subject := ""
		for _, h := range msg.Payload.Headers {
			if h.Name == "Date" {
				date = h.Value
			}
			if h.Name == "Subject" {
				subject = h.Value
			}
			if h.Name == "From" {
				from = strings.Split(strings.TrimSuffix(h.Value, ">"), "<")[1]
			}
		}
		fmt.Printf("Processing email from <%s> having subject <%s> on date <%s>\n", from, subject, date)

		for _, part := range msg.Payload.Parts {
			if part.MimeType == "application/octet-stream" {
				data, _ := base64.URLEncoding.DecodeString(part.Body.Data)

				tm := time.Unix(msg.InternalDate/1000, 0)
				year, month, day := tm.Date()

				stringdate := fmt.Sprintf("%v-%v-%v", day, month, year)
				fullPath := c.getPath(from) + "/" + stringdate
				_, err = os.Stat(fullPath)
				if os.IsNotExist(err) {
					err = os.MkdirAll(fullPath, 0755)
					if err != nil {
						fmt.Println("cannot create directory", err)
						return
					}
				}

				fileName := fmt.Sprintf("%s/%s", fullPath, part.Filename)
				// // If the file doesn't exist, create it, or append to the file
				f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					log.Fatal(err)
				}

				_, err = f.Write(data)
				if err != nil {
					log.Fatal(err)
				}

				f.Close()
			}
		}
	}

	for {

	}
}
