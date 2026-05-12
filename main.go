package main

import (
	"context"
	"log"

	"github.com/lakerszhy/thn/domain"
	"github.com/lakerszhy/thn/hn"
)

func main() {
	client, err := hn.New(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	p := domain.NewPagination(20)

	items, err := client.FetchItems(context.Background(), domain.CategoryTop, p)
	if err != nil {
		log.Fatal(err)
	}

	for _, item := range items {
		log.Println(item.Title)
	}

	p = p.Next()
	items, err = client.FetchItems(context.Background(), domain.CategoryTop, p)
	if err != nil {
		log.Fatal(err)
	}

	for _, item := range items {
		log.Println(item.Title)
	}
}
