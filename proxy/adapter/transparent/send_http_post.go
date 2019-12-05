package transparent

import (
	"bytes"
	"context"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"time"
)

func sendHttpPost(endpoint string, payload []byte) (*http.Response, []byte, error) {
	if len(payload) == 0 {
		return nil, nil, errors.New("payload sent by http is empty")
	}

	// FIXME propagate timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(payload))
	req.Header.Set("Content-Type", orbs.CONTENT_TYPE_MEMBUFFERS)

	res, err := http.DefaultClient.Do(req.WithContext(ctx))
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
