package checks

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"

	"github.com/mtojek/greenwall/middleware/monitoring"
)

const (
	httpCheckName              = "http_check"
	expectedPatternParameter   = "expectedPattern"
	basicAuthUsernameParameter = "basicAuthUsername"
	basicAuthPasswordParameter = "basicAuthPassword"
)

// HTTPCheck performs a simple check against an HTTP endpoint.
type HTTPCheck struct {
	monitoringConfiguration *monitoring.Configuration
	nodeConfig              *monitoring.Node

	client            *http.Client
	searchedPattern   []byte
	basicAuthUsername string
	basicAuthPassword string
}

// Initialize method initializes the check instance.
func (h *HTTPCheck) Initialize(monitoringConfiguration *monitoring.Configuration, nodeConfig *monitoring.Node) {
	h.monitoringConfiguration = monitoringConfiguration
	h.nodeConfig = nodeConfig

	h.client = &http.Client{Timeout: h.monitoringConfiguration.General.HTTPClientTimeout}

	h.searchedPattern = []byte(h.nodeConfig.ExpectedPattern) // Deprecation
	if len(h.searchedPattern) == 0 {
		h.searchedPattern = []byte(h.nodeConfig.Parameters[expectedPatternParameter])
	}

	h.basicAuthUsername = h.nodeConfig.Parameters[basicAuthUsernameParameter]
	h.basicAuthPassword = h.nodeConfig.Parameters[basicAuthPasswordParameter]
}

// Run method executes the check. This is invoked periodically.
func (h *HTTPCheck) Run() CheckResult {
	result := CheckResult{
		Status: StatusDanger,
	}

	req, err := http.NewRequest(http.MethodGet, h.nodeConfig.Endpoint, nil)
	if err != nil {
		log.Println(err)

		result.Message = err.Error()
		return result
	}

	if h.withBasicAuth() {
		req.SetBasicAuth(h.basicAuthUsername, h.basicAuthPassword)
	}

	response, err := h.client.Do(req)
	if err != nil {
		log.Println(err)

		result.Message = err.Error()
		return result
	}

	if h.isHTTPError(response.StatusCode) {
		e := fmt.Errorf("%d - %s", response.StatusCode, http.StatusText(response.StatusCode))
		log.Println(e)

		result.Message = e.Error()
		return result
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)

		result.Message = err.Error()
		return result
	}

	if response.Body != nil {
		errClosing := response.Body.Close()
		if err != nil {
			log.Println(errClosing)

			result.Message = errClosing.Error()
			return result
		}
	}

	if len(h.searchedPattern) > 0 && !bytes.Contains(body, h.searchedPattern) {
		result.Message = MessagePatternNotFound
		return result
	}

	result.Status = StatusSuccess
	result.Message = MessageOK
	return result
}

func (h *HTTPCheck) withBasicAuth() bool {
	return h.basicAuthUsername != "" && h.basicAuthPassword != ""
}

func (h *HTTPCheck) isHTTPError(code int) bool {
	return code >= http.StatusBadRequest
}

func init() {
	registerType(httpCheckName, reflect.TypeOf(HTTPCheck{}))
}
