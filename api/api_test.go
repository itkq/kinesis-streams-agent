package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPI(t *testing.T) {
	invalidAddress := "hogehoge"
	api, err := NewAPI(invalidAddress)
	assert.NotEqual(t, err, nil)

	address := "localhost:8080"
	api, err = NewAPI(address)
	assert.Equal(t, err, nil)

	helloExporter := &TestExporter{
		endpoint: "/hello",
		metrics:  &HelloMetrics{},
	}

	api.Register(helloExporter)
	go api.Run()

	endpoint := fmt.Sprintf("http://%s%s", address, helloExporter.Endpoint())
	resp, err := http.Get(endpoint)
	assert.Equal(t, nil, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	data, err := ioutil.ReadAll(resp.Body)
	b, _ := json.Marshal(helloExporter.Export())
	assert.Equal(t, nil, err)
	assert.Equal(t, b, data)
}

type TestExporter struct {
	metrics  interface{}
	endpoint string
}

func (e *TestExporter) Export() interface{} {
	return e.metrics
}

func (e *TestExporter) Endpoint() string {
	return e.endpoint
}

type HelloMetrics struct{}
