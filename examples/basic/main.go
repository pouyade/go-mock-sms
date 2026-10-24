package main

import (
	"log"

	notify "github.com/pouyade/go-mock-sms"
)

func main() {
	if err := notify.Send("Verification code", "Your code is 123456"); err != nil {
		log.Fatal(err)
	}
}
