package main

import (
	"encoding/base64"
	"fmt"
	"log"

	usbdrivedetector "github.com/deepakjois/gousbdrivedetector"
	"google.golang.org/api/gmail/v1"
)

func detectusb(computerName, senderEmail, receiverEmail string, srv *gmail.Service) {
	if drives, err := usbdrivedetector.Detect(); err == nil {
		log.Printf("%d USB Devices Found\n", len(drives))
		for _, d := range drives {
			log.Println(d)
		}

		if len(drives) > 0 {
			// New message for our gmail service to send
			var message gmail.Message

			// Compose the message
			messageStr := []byte(
				fmt.Sprintf("From: %s\r\n", senderEmail) +
					fmt.Sprintf("To: mailto:%s\r\n", receiverEmail) +
					"Subject: USB detected message\r\n\r\n" +
					fmt.Sprintf("Your computer (%s) has %v usb devices attached", computerName, len(drives)))

			// Place messageStr into message.Raw in base64 encoded format
			message.Raw = base64.URLEncoding.EncodeToString(messageStr)

			// Send the message
			_, err = srv.Users.Messages.Send("me", &message).Do()
			if err != nil {
				log.Printf("error sending email: %v", err)
			} else {
				log.Println("Message sent!")
			}
		}

	} else {
		log.Println("error while reading usb device")
	}
}
