// Copyright 2019 the orbs-network-go authors
// This file is part of the orbs-network-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package httpserver

import (
	"encoding/json"
	"github.com/orbs-network/scribe/log"
	"github.com/orbs-network/trash-panda/config"
	"net/http"
)

type IndexResponse struct {
	Status      string
	Description string
	Version     config.Version
}

// Serves both index and 404 because router is built that way
func (s *HttpServer) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	data, _ := json.MarshalIndent(IndexResponse{
		Status:      "OK",
		Description: "ORBS blockchain proxy API",
		Version:     config.GetVersion(),
	}, "", "  ")

	_, err := w.Write(data)
	if err != nil {
		s.logger.Info("error writing index.json response", log.Error(err))
	}
}

func (s *HttpServer) robots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, err := w.Write([]byte("User-agent: *\nDisallow: /\n"))
	if err != nil {
		s.logger.Info("error writing robots.txt response", log.Error(err))
	}
}
