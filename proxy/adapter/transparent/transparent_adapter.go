package transparent

import (
	"bytes"
	"github.com/orbs-network/membuffers/go"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
	"github.com/orbs-network/scribe/log"
	"github.com/orbs-network/trash-panda/proxy"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

type adapter struct {
	handlers []proxy.Handler
}

type builderFunc func(h *handler, input []byte) (message membuffers.Message, err *proxy.HttpErr)

type handler struct {
	name   string
	path   string
	config proxy.Config

	f builderFunc
}

func NewTransparentAdapter(config proxy.Config) proxy.ProxyAdapter {
	return &adapter{
		[]proxy.Handler{
			&handler{
				"run-query",
				"/api/v1/run-query",
				config,
				runQueryFunc,
			},
		},
	}
}

func (a *adapter) Handlers() []proxy.Handler {
	return a.handlers
}

func (h *handler) Name() string {
	return h.name
}

func (h *handler) Path() string {
	return h.path
}

func (h *handler) Handler() proxy.HandlerBuilderFunc {
	return func(input []byte) (message membuffers.Message, err *proxy.HttpErr) {
		return h.f(h, input)
	}
}

func runQueryFunc(h *handler, input []byte) (membuffers.Message, *proxy.HttpErr) {
	clientRequest := client.RunQueryRequestReader(input)
	if e := validate(clientRequest); e != nil {
		return nil, e
	}

	//h.logger.Info("http HttpServer received run-query", log.Stringable("request", clientRequest))

	res, resBody, err := sendHttpPost(h.config.Endpoints[0]+h.path, input)
	if err != nil {
		return nil, &proxy.HttpErr{http.StatusBadRequest, nil, err.Error()}
	}

	if res.StatusCode != http.StatusOK {
		return nil, &proxy.HttpErr{res.StatusCode, nil, res.Header.Get("X-ORBS-ERROR-DETAILS")}
	}

	return client.RunQueryResponseReader(resBody), nil
}

func validate(m membuffers.Message) *proxy.HttpErr {
	if !m.IsValid() {
		return &proxy.HttpErr{http.StatusBadRequest, log.Stringable("request", m), "http request is not a valid membuffer"}
	}
	return nil
}

func sendHttpPost(endpoint string, payload []byte) (*http.Response, []byte, error) {
	if len(payload) == 0 {
		return nil, nil, errors.New("payload sent by http is empty")
	}

	res, err := http.Post(endpoint, orbs.CONTENT_TYPE_MEMBUFFERS, bytes.NewReader(payload))
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed sending http post")
	}

	buf, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil, buf, errors.Wrap(err, "failed reading http response")
	}

	// check if we have the content type response we expect
	contentType := res.Header.Get("Content-Type")
	if contentType != orbs.CONTENT_TYPE_MEMBUFFERS {

		// handle real 404 (incorrect endpoint) gracefully
		if res.StatusCode == 404 {
			// TODO: streamline these errors
			return res, buf, errors.Wrap(orbs.NoConnectionError, "http 404 not found")
		}

		if contentType == "text/plain" || contentType == "application/json" {
			return nil, buf, errors.Errorf("http request failed: %s", string(buf))
		} else {
			return nil, buf, errors.Errorf("http request failed with Content-Type '%s': %x", contentType, buf)
		}
	}

	return res, buf, nil
}
