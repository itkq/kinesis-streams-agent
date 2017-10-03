package api

import (
	"errors"
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

	errorstr := "error!"
	errorExporter := &TestExporter{
		endpoint: "/error",
		metrics: &ErrorMetrics{
			errorstr: errorstr,
		},
	}

	api.Register(helloExporter)
	api.Register(errorExporter)
	go api.Run()

	// test helloExporter
	endpoint := fmt.Sprintf("http://%s%s", address, helloExporter.Endpoint())
	resp, err := http.Get(endpoint)
	assert.Equal(t, nil, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	data, err := ioutil.ReadAll(resp.Body)
	b, _ := helloExporter.Export().ToJSON()
	assert.Equal(t, nil, err)
	assert.Equal(t, b, data)

	// test errorExporter
	endpoint = fmt.Sprintf("http://%s%s", address, errorExporter.Endpoint())
	resp, err = http.Get(endpoint)
	assert.Equal(t, nil, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	data, err = ioutil.ReadAll(resp.Body)
	assert.Equal(t, nil, err)
	assert.Equal(t, data, []byte(fmt.Sprintf("{\"error\":\"%s\"}", errorstr)))
}

type TestExporter struct {
	metrics  Metrics
	endpoint string
}

func (e *TestExporter) Export() Metrics {
	return e.metrics
}

func (e *TestExporter) Endpoint() string {
	return e.endpoint
}

type HelloMetrics struct{}

func (m *HelloMetrics) ToJSON() ([]byte, error) {
	return []byte("hello"), nil
}

type ErrorMetrics struct {
	errorstr string
}

func (m *ErrorMetrics) ToJSON() ([]byte, error) {
	return []byte{}, errors.New(m.errorstr)
}
