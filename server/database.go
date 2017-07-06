package server

import (
	"log"
)

func LogQuery(query string) {
	// log.Println(query)
}

func logQueryCount(count int) {
	log.Println("Database queries:", count)
}
