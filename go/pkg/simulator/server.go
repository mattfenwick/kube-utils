package simulator

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/rand"
	"net/http"
	"time"
)

func RunServer(jaegerTracerProvider *tracesdk.TracerProvider) {
	port := 19999
	server := &Server{Provider: jaegerTracerProvider}
	SetupHTTPServer(server)

	stop := make(chan struct{})

	addr := fmt.Sprintf(":%d", port)
	go func() {
		logrus.Infof("starting HTTP server on port %d", port)
		utils.DoOrDie(http.ListenAndServe(addr, nil))
	}()
	<-stop
}

type Server struct {
	Provider *tracesdk.TracerProvider
}

func (s *Server) StartScan(scan *StartScan) {
	_, span := s.Provider.Tracer("server").Start(context.Background(), "start")
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	logrus.Infof("scan: %s of %d\n", scan.Data[:15], len(scan.Data))
	span.End()
}

func (s *Server) FetchScanResults(scanId string) (*ScanResults, error) {
	_, span := s.Provider.Tracer("server").Start(context.Background(), "fetch")
	defer span.End()
	return &ScanResults{
		IsDone: false,
		Data:   scanId + rand.String(40_000),
	}, nil
}

func (s *Server) NotFound(w http.ResponseWriter, r *http.Request) {
	_, span := s.Provider.Tracer("server").Start(context.Background(), "not-found")
	defer span.End()

	w.WriteHeader(404)
	_, err := w.Write([]byte("not found"))
	utils.DoOrDie(err)
}

func (s *Server) Error(w http.ResponseWriter, r *http.Request, httpError error, statusCode int) {
	_, span := s.Provider.Tracer("server").Start(context.Background(), "error")
	defer span.End()

	w.WriteHeader(statusCode)
	_, err := w.Write([]byte(httpError.Error()))
	utils.DoOrDie(err)
}

// Responder .....
type Responder interface {
	StartScan(scan *StartScan)
	FetchScanResults(scanId string) (*ScanResults, error)

	NotFound(w http.ResponseWriter, r *http.Request)
	Error(w http.ResponseWriter, r *http.Request, err error, statusCode int)
}

// SetupHTTPServer .....
func SetupHTTPServer(responder Responder) {
	http.HandleFunc("/scan", func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())
		span.AddEvent("server received")
		switch r.Method {
		case "POST":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				logrus.Errorf("unable to read body for pod POST: %s", err.Error())
				responder.Error(w, r, err, 400)
				return
			}
			var request StartScan
			err = json.Unmarshal(body, &request)
			if err != nil {
				logrus.Errorf("unable to ummarshal JSON for scan POST: %s", err.Error())
				responder.Error(w, r, err, 400)
				return
			}
			responder.StartScan(&request)
			_, err = fmt.Fprint(w, "")
			utils.DoOrDie(err)
		case "GET":
			ids, ok := r.URL.Query()["scan-id"]
			if !ok || len(ids) == 0 {
				responder.Error(w, r, errors.Errorf("missing scan-id parameter"), 400)
				return
			}

			scanId := ids[0]

			results, err := responder.FetchScanResults(scanId)
			if err != nil {
				responder.Error(w, r, err, 400)
				return
			}

			bytes, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				responder.Error(w, r, err, 500)
				return
			}
			w.Header().Set(http.CanonicalHeaderKey("content-type"), "application/json")
			_, err = fmt.Fprint(w, string(bytes))
			utils.DoOrDie(err)
		default:
			responder.NotFound(w, r)
		}
	})
}
