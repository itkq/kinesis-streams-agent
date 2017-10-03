package api

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

type Metrics interface {
	ToJSON() ([]byte, error)
}

type Exporter interface {
	Export() Metrics
	Endpoint() string
}

type API struct {
	listener  net.Listener
	exporters []Exporter
}

func NewAPI(address string) (*API, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	return &API{
		listener:  listener,
		exporters: make([]Exporter, 0),
	}, nil
}

func (a *API) Run() {
	mux := http.NewServeMux()

	for _, e := range a.exporters {
		mux.HandleFunc(e.Endpoint(), a.Handler(e))
	}

	server := http.Server{
		Handler: mux,
	}

	log.Printf("info: api server listening on http://%s/\n", a.listener.Addr())
	server.Serve(a.listener)
}

func (m *API) Register(e Exporter) {
	m.exporters = append(m.exporters, e)
}

func (m *API) Handler(e Exporter) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := e.Export().ToJSON()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err.Error())))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(b)
		}
	}
}
