package main

import (
	"encoding/base64"
	"fmt"
	"log"

	usbdrivedetector "github.com/deepakjois/gousbdrivedetector"
	"google.golang.org/api/gmail/v1"
)

func detectusb(name string, srv *gmail.Service) {
	if drives, err := usbdrivedetector.Detect(); err == nil {
		fmt.Printf("%d USB Devices Found\n", len(drives))
		for _, d := range drives {
			fmt.Println(d)
		}

		if len(drives) > 0 {
			// New message for our gmail service to send
			var message gmail.Message

			// Compose the message
			messageStr := []byte(
				"From: sharadregoti15@gmail.com\r\n" +
					"To: mailto:cadramesha@gmail.com\r\n" +
					"Subject: USB detected message\r\n\r\n" +
					fmt.Sprintf("Your computer (%s) has %v usb devices attached", name, len(drives)))

			// Place messageStr into message.Raw in base64 encoded format
			message.Raw = base64.URLEncoding.EncodeToString(messageStr)

			// Send the message
			_, err = srv.Users.Messages.Send("me", &message).Do()
			if err != nil {
				log.Printf("error sending email: %v", err)
			} else {
				fmt.Println("Message sent!")
			}
		}

	} else {
		fmt.Println("error while reading usb device")
	}
}
