package main

import (
	"log"

	notify "github.com/pouyade/go-mock-sms"
)

func main() {
	client := notify.New()
	if err := client.Send(notify.Message{
		Title: "ExampleApp",
		Text:  "Your verification code is 123456",
	}); err != nil {
		log.Fatal(err)
	}
}
