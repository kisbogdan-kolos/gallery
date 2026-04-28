package main

import (
	"fmt"
	"log"

	"github.com/kisbogdan-kolos/gallery/backend/db"
	"github.com/kisbogdan-kolos/gallery/backend/queue"
	"github.com/kisbogdan-kolos/gallery/backend/storage"
	"github.com/kisbogdan-kolos/gallery/ocr/ocr"
)

func main() {
	err := db.DbConnect()
	if err != nil {
		log.Fatal(err)
	}

	err = storage.Init()
	if err != nil {
		log.Fatal(err)
	}

	err = queue.Connect()
	defer queue.Disconnect()
	if err != nil {
		log.Fatal(err)
	}

	msgs, err := queue.GetConsumer()
	if err != nil {
		log.Fatal(err)
	}

	for d := range msgs {
		id := uint(0)
		fmt.Sscanf(string(d.Body), "%d", &id)

		err := ocr.Run(id)
		if err != nil {
			log.Print(err)
		}
	}
}
