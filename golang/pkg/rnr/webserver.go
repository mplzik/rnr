package rnr

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/golang/protobuf/jsonpb"
	"github.com/mplzik/rnr/golang/pkg/pb"
)

type RnrWebServer struct {
	job *Job
}

// The web UI used to display the data
//go:embed ui
var content embed.FS

func NewRnrWebserver(job *Job) *RnrWebServer {
	ret := &RnrWebServer{
		job: job,
	}

	return ret
}

func (rnr *RnrWebServer) tasksHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		m := jsonpb.Marshaler{
			EmitDefaults: true,
		}
		w.Header().Set("Content-Type", "application/json")
		err := m.Marshal(w, rnr.job.GetProto())
		if err != nil {
			log.Fatal("Failed to convert a task to json:", err.Error())
		}

	case "POST":
		tr := &pb.TaskRequest{}
		err := jsonpb.Unmarshal(r.Body, tr)
		if err != nil {
			log.Printf("Failed to convert body to JSON: %s", err.Error())
			return
		}
		fmt.Println(tr)
		err = rnr.job.TaskRequest(tr)
		if err != nil {
			log.Printf("Failed to process task request %s: %s", tr, err.Error())
		}
		w.Write([]byte{})
	}
}

func (rnr *RnrWebServer) RegisterHttp(urlPrefix string) {

	fsRoot, err := fs.Sub(content, "ui")
	if err != nil {
		log.Fatalf("Vendored data doesn't contain subdir 'ui', something went wrong: %s", err.Error())
	}
	fs := http.FileServer(http.FS(fsRoot))
	http.Handle(urlPrefix+"/", fs)
	http.HandleFunc(urlPrefix+"/tasks", rnr.tasksHandler)

}
