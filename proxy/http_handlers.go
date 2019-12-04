package proxy

import (
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
	"github.com/orbs-network/scribe/log"
	"net/http"
)

func (s *Service) sendTransactionHandler(w http.ResponseWriter, r *http.Request) {
	bytes, e := readInput(r)
	if e != nil {
		s.writeErrorResponseAndLog(w, e)
		return
	}

	clientRequest := client.SendTransactionRequestReader(bytes)
	if e := validate(clientRequest); e != nil {
		s.writeErrorResponseAndLog(w, e)
		return
	}

	s.logger.Info("http HttpServer received send-transaction", log.Stringable("request", clientRequest))

	res, resBody, err := sendHttpPost(s.config.Endpoints[0]+SEND_TRANSACTION, bytes)
	if err != nil {
		s.logger.Error(err.Error())
	}

	result := client.SendTransactionResponseReader(resBody)
	if result.IsValid() {
		s.writeMembuffResponse(w, result, res.StatusCode, err)
	} else {
		s.writeErrorResponseAndLog(w, &HttpErr{http.StatusInternalServerError, log.Error(err), err.Error()})
	}
}

func (s *Service) sendTransactionAsyncHandler(w http.ResponseWriter, r *http.Request) {
	bytes, e := readInput(r)
	if e != nil {
		s.writeErrorResponseAndLog(w, e)
		return
	}

	clientRequest := client.SendTransactionRequestReader(bytes)
	if e := validate(clientRequest); e != nil {
		s.writeErrorResponseAndLog(w, e)
		return
	}

	s.logger.Info("http HttpServer received send-transaction-async", log.Stringable("request", clientRequest))

	panic("not implemented")
}

func (s *Service) runQueryHandler(w http.ResponseWriter, r *http.Request) {
	bytes, e := readInput(r)
	if e != nil {
		s.writeErrorResponseAndLog(w, e)
		return
	}

	clientRequest := client.RunQueryRequestReader(bytes)
	if e := validate(clientRequest); e != nil {
		s.writeErrorResponseAndLog(w, e)
		return
	}

	s.logger.Info("http HttpServer received run-query", log.Stringable("request", clientRequest))

	res, resBody, err := sendHttpPost(s.config.Endpoints[0]+RUN_QUERY, bytes)
	if err != nil {
		s.logger.Error(err.Error())
	}

	result := client.RunQueryResponseReader(resBody)
	if result.IsValid() {
		s.writeMembuffResponse(w, result, res.StatusCode, err)
	} else {
		s.writeErrorResponseAndLog(w, &HttpErr{http.StatusInternalServerError, log.Error(err), err.Error()})
	}
}

func (s *Service) getTransactionStatusHandler(w http.ResponseWriter, r *http.Request) {
	bytes, e := readInput(r)
	if e != nil {
		s.writeErrorResponseAndLog(w, e)
		return
	}

	clientRequest := client.GetTransactionStatusRequestReader(bytes)
	if e := validate(clientRequest); e != nil {
		s.writeErrorResponseAndLog(w, e)
		return
	}

	s.logger.Info("http HttpServer received get-transaction-status", log.Stringable("request", clientRequest))

	panic("not implemented")
}

func (s *Service) getTransactionReceiptProofHandler(w http.ResponseWriter, r *http.Request) {
	bytes, e := readInput(r)
	if e != nil {
		s.writeErrorResponseAndLog(w, e)
		return
	}

	clientRequest := client.GetTransactionReceiptProofRequestReader(bytes)
	if e := validate(clientRequest); e != nil {
		s.writeErrorResponseAndLog(w, e)
		return
	}

	s.logger.Info("http HttpServer received get-transaction-receipt-proof", log.Stringable("request", clientRequest))

	panic("not implemented")
}

func (s *Service) getBlockHandler(w http.ResponseWriter, r *http.Request) {
	bytes, e := readInput(r)
	if e != nil {
		s.writeErrorResponseAndLog(w, e)
		return
	}

	clientRequest := client.GetBlockRequestReader(bytes)
	if e := validate(clientRequest); e != nil {
		s.writeErrorResponseAndLog(w, e)
		return
	}

	s.logger.Info("http HttpServer received get-block", log.Stringable("request", clientRequest))

	panic("not implemented")
}
