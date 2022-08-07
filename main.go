package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// TODO: Error handling
// TODO: Writing logs to a file
// TODO: Add readme file
// How to run it
// How to build it
// Hot to operate it

type Info struct {
	Email string `yaml:"email"`
	Path  string `yaml:"path"`
}

type Config struct {
	Name          string `yaml:"name"`
	NotifierEmail string `yaml:"notifierEmail"`
	Checks        []Info `yaml:"checks"`
}

func (k Config) getPath(from string) string {
	for _, kk := range k.Checks {
		if kk.Email == from {
			return kk.Path
		}
	}
	return ""
}

func main() {

	f, err := os.OpenFile("cad-creation-automation.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	wrt := io.MultiWriter(os.Stdout, f)
	log.SetOutput(wrt)

	configData, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Println("Failed to read config file")
		return
	}

	var c Config
	err = yaml.Unmarshal(configData, &c)
	if err != nil {
		log.Printf("err: %v\n", err)
		return
	}

	log.Println("Starting the process", time.Now().Local().String(), c.Name)

	ctx := context.Background()
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
		return
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailModifyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
		return
	}
	checkMailClient := getClient(config, "checkMailClientToken")

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(checkMailClient))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
		return
	}

	llist, err := srv.Users.Labels.List("me").Do()
	if err != nil {
		log.Println("failed to list label", err)
		return
	}

	cadCreationLabelID := ""
	for _, label := range llist.Labels {
		if label.Name == "cad-creation-automation" {
			cadCreationLabelID = label.Id
		}
	}

	if cadCreationLabelID == "" {
		res, err := srv.Users.Labels.Create("me", &gmail.Label{Name: "cad-creation-automation"}).Do()
		if err != nil {
			log.Println("failed to create label", err)
			return
		}
		cadCreationLabelID = res.Id
	}

	fromClause := ""
	for _, cc := range c.Checks {
		fromClause += cc.Email + ","
	}
	fromClause = strings.TrimSuffix(fromClause, ",")

	tick := time.NewTicker(30 * time.Second)
	defer tick.Stop()

	profile, err := srv.Users.GetProfile("me").Do()
	if err != nil {
		log.Println("cannot get profile", err)
		return
	}

	log.Println("email address is ", profile.EmailAddress)

	for range tick.C {

		// process 2 days message
		query := fmt.Sprintf("from:%s,label:(inbox,-cad-creation-automation),newer_than:2d", fromClause)
		req := srv.Users.Messages.List("me").Q(query)
		// if pageToken != "" {
		// 	req.PageToken(pageToken)
		// }
		req.MaxResults(50)
		r, err := req.Do()
		if err != nil {
			log.Printf("Unable to retrieve messages: %v\n", err)
			continue
		}

		log.Println()
		log.Printf("Processing %v messages...\n", len(r.Messages))
		for _, m := range r.Messages {
			msg, err := srv.Users.Messages.Get("me", m.Id).Do()
			if err != nil {
				log.Printf("Unable to retrieve message %v: %v\n", m.Id, err)
				log.Println("One attachement is being skipped")
				continue
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
			log.Printf("Processing email from <%s> having subject <%s> on date <%s>\n", from, subject, date)

			for _, part := range msg.Payload.Parts {
				if part.MimeType == "application/octet-stream" {

					attachmentID := part.Body.AttachmentId

					res, err := srv.Users.Messages.Attachments.Get("me", msg.Id, attachmentID).Do()
					if err != nil {
						log.Println("failed to download attachement", err)
						continue
					}

					data, err := base64.URLEncoding.DecodeString(res.Data)
					if err != nil {
						log.Println("error decoding body data", err)
						continue
					}

					tm := time.Unix(msg.InternalDate/1000, 0)
					year, month, day := tm.Date()

					stringdate := fmt.Sprintf("%v-%v-%v", day, month, year)
					fullPath := c.getPath(from) + "/" + stringdate
					_, err = os.Stat(fullPath)
					if os.IsNotExist(err) {
						err = os.MkdirAll(fullPath, 0755)
						if err != nil {
							log.Println("cannot create directory", err)
							continue
						}
					}

					fileName := fmt.Sprintf("%s/%s", fullPath, part.Filename)
					// // If the file doesn't exist, create it, or append to the file
					f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						log.Println("failed to open file", err)
						continue
					}

					_, err = f.Write([]byte(data))
					if err != nil {
						log.Println("failed to write data into file", err)
						continue
					}

					f.Close()
				}
			}

			_, err = srv.Users.Messages.Modify("me", msg.Id, &gmail.ModifyMessageRequest{AddLabelIds: []string{cadCreationLabelID}}).Do()
			if err != nil {
				log.Println("failed to modify message", err)
				continue
			}
		}

		log.Println("Checking USB")
		detectusb(c.Name, profile.EmailAddress, c.NotifierEmail, srv)
	}
}
