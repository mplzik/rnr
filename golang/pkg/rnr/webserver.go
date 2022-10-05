package rnr

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/mplzik/rnr/golang/pkg/pb"
	"github.com/mplzik/rnr/ui"
	"google.golang.org/protobuf/encoding/protojson"
)

type RnrWebServer struct {
	job *Job
}

func NewRnrWebserver(job *Job) *RnrWebServer {
	ret := &RnrWebServer{
		job: job,
	}

	return ret
}

func (rnr *RnrWebServer) tasksHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		m := protojson.MarshalOptions{
			EmitUnpopulated: true,
		}
		b, err := m.Marshal(rnr.job.Proto(nil))
		if err != nil {
			log.Fatal("Failed to convert a task to json:", err.Error())
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(b) //nolint:errcheck

	case "POST":
		defer r.Body.Close()
		b, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to read request body: %v", err)
			return
		}

		tr := &pb.TaskRequest{}
		err = protojson.Unmarshal(b, tr)
		if err != nil {
			log.Printf("Failed to convert body to JSON: %s", err.Error())
			return
		}
		fmt.Println(tr)
		err = rnr.job.TaskRequest(tr)
		if err != nil {
			log.Printf("Failed to process task request %s: %s", tr, err.Error())
		}
		w.Write([]byte{}) //nolint:errcheck
	}
}

func (rnr *RnrWebServer) RegisterHttp(urlPrefix string) {
	fs := http.FileServer(http.FS(ui.Content))
	http.Handle(urlPrefix+"/", fs)
	http.HandleFunc(urlPrefix+"/tasks", rnr.tasksHandler)
}
