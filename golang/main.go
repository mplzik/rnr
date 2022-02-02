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
		n.Add(rnr.NewSimpleCallbackTask(fmt.Sprintf("Hello %d", i), func(_ *rnr.SimpleCallbackTask, ctx context.Context) (bool, error) {
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
