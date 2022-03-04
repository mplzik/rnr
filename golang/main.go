package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	rnr "github.com/mplzik/rnr/golang/pkg/rnr"
)

func main() {
	fmt.Println("Hello, world!")
	// Create a faux job
	n := rnr.NewNestedTask("Root task", 1)
	job := rnr.NewJob(n)

	for i := 0; i < 100; i++ {
		n.Add(rnr.NewCallbackTask(fmt.Sprintf("Hello %d", i), func(context.Context, *rnr.CallbackTask) (bool, error) {
			if rand.Intn(3) > 1 {
				return true, nil
			} else {
				return true, fmt.Errorf("bad luck")
			}
		}))
	}

	// Listen and serve

	rnr := rnr.NewRnrWebserver(job)
	rnr.RegisterHttp("")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
