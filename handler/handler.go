package handler

import (
	"fmt"
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"io"
	"net/http"
	internalconfig "redis-postgres-service/config"
	"redis-postgres-service/controller/incremental"
	"redis-postgres-service/controller/sign"
	"redis-postgres-service/controller/users"
	"redis-postgres-service/entity"
	mapper "redis-postgres-service/mapper/common"
)

const configKey = "handler"

type Handler interface {
	Incremental(w http.ResponseWriter, req *http.Request)
	Signature(w http.ResponseWriter, req *http.Request)
	AddUser(w http.ResponseWriter, req *http.Request)
}

// Compile time check that handler implements Handler interface
var _ Handler = (*handler)(nil)

type handler struct {
	logger          *zap.Logger
	usersCtrl       users.Controller
	incrementalCtrl incremental.Controller
	signCtrl        sign.Controller
	config          internalconfig.HandlerConfig
}

// Params is an fx container for all Controller Handler
type Params struct {
	fx.In

	Logger          *zap.Logger
	ConfigProvider  config.Provider
	UsersCtrl       users.Controller
	IncrementalCtrl incremental.Controller
	SignController  sign.Controller
}

// New is a constructor of Handler interface that is provided to the fx
func New(p Params) (Handler, error) {
	var cfg internalconfig.HandlerConfig
	err := p.ConfigProvider.Get(configKey).Populate(&cfg) // unreachable in tests, cause provider is populating from valid yaml.
	if err != nil {
		return nil, err
	}
	return &handler{
		logger:          p.Logger,
		usersCtrl:       p.UsersCtrl,
		incrementalCtrl: p.IncrementalCtrl,
		signCtrl:        p.SignController,
		config:          cfg,
	}, nil
}

// Incremental is a POST endpoint to that increments provided key in the request body in redis by the value
// provided in the same request and returns the result of the incrementation
// expected JSON request is defined by entity.IncrementRequest
// expected JSON response is defined by entity.IncrementResponse
func (h *handler) Incremental(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.With(
		zap.String("scope", "handler"),
		zap.String("function", "Incremental"),
	).Sugar()
	logger.Info("Incoming request")
	defer req.Body.Close()
	if req.ContentLength > h.config.RequestBodyLimit {
		http.Error(
			w,
			fmt.Sprintf(entity.BadRequest, "request body is too big"),
			http.StatusBadRequest,
		)
		logger.Error(entity.UnableToReadTheBody)
		return
	}
	data, err := io.ReadAll(io.LimitReader(req.Body, h.config.RequestBodyLimit))
	if err != nil {
		http.Error(w, entity.UnableToReadTheBody, http.StatusBadRequest)
		logger.Error(entity.UnableToReadTheBody)
		return
	}
	request, err := mapper.BytesToType[entity.IncrementRequest](data)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf(entity.BadRequest, err),
			http.StatusBadRequest,
		)
		logger.Errorf(entity.BadRequest, err)
		return
	}
	logger.With("request", request).Info("Request received")
	response, err := h.incrementalCtrl.Inc(req.Context(), request)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf(entity.FailedToProcessTheRequest, err),
			http.StatusBadGateway,
		)
		logger.Errorf(entity.FailedToProcessTheRequest, err)
		return
	}
	incrementResponse, err := mapper.TypeToBytes[entity.IncrementResponse](response)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf(entity.FailedToProcessTheResponse, err),
			http.StatusInternalServerError,
		)
		logger.Errorf(entity.FailedToProcessTheResponse, err)
		return // unreachable in tests cause response struct can always be represented as json
	}
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(incrementResponse)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf(entity.FailedToWriteTheResponse, err),
			http.StatusInternalServerError,
		)
		logger.Errorf(entity.FailedToWriteTheResponse, err)
		return // unreachable in tests
	}
	logger.With("response", response).Info("Request completed")
	return
}

// Signature is a POST endpoint to that signs the provided text in the request body by the key using SHA512
// algorythm and returns a respective Hex signature
// expected JSON request is defined by entity.SignRequest
// expected JSON response is defined by entity.SignResponse
func (h *handler) Signature(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.With(
		zap.String("scope", "handler"),
		zap.String("function", "Signature"),
	).Sugar()
	logger.Info("Request received")
	defer req.Body.Close()
	if req.ContentLength > h.config.RequestBodyLimit {
		http.Error(
			w,
			fmt.Sprintf(entity.BadRequest, "request body is too big"),
			http.StatusBadRequest,
		)
		logger.Error(entity.UnableToReadTheBody)
		return
	}
	data, err := io.ReadAll(io.LimitReader(req.Body, h.config.RequestBodyLimit))
	if err != nil {
		http.Error(w, entity.UnableToReadTheBody, http.StatusBadRequest)
		logger.Error(entity.UnableToReadTheBody)
		return
	}
	request, err := mapper.BytesToType[entity.SignRequest](data)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf(entity.BadRequest, err),
			http.StatusBadRequest,
		)
		logger.Errorf(entity.BadRequest, err)
		return
	}
	response, err := h.signCtrl.Sign(req.Context(), request)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf(entity.FailedToProcessTheRequest, err),
			http.StatusBadGateway,
		)
		logger.Errorf(entity.FailedToProcessTheRequest, err)
		return
	}
	signResponse, err := mapper.TypeToBytes[entity.SignResponse](response)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf(entity.FailedToProcessTheResponse, err),
			http.StatusInternalServerError,
		)
		logger.Errorf(entity.FailedToProcessTheResponse, err)
		return // unreachable in tests cause response struct can always be represented as json
	}
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(signResponse)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf(entity.FailedToWriteTheResponse, err),
			http.StatusInternalServerError,
		)
		logger.Errorf(entity.FailedToWriteTheResponse, err)
		return // unreachable in tests
	}
	logger.With("response", response).Info("Request completed")
	return
}

// AddUser is a POST endpoint to that appends a row to `users` table in postgres
// using the user/age keys provided in the request and returns the id of the new row in the table
// expected JSON request is defined by entity.AddUserRequest
// expected JSON response is defined by entity.AddUserResponse
func (h *handler) AddUser(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.With(
		zap.String("scope", "handler"),
		zap.String("function", "AddUser"),
	).Sugar()
	logger.Info("Request received")
	defer req.Body.Close()
	if req.ContentLength > h.config.RequestBodyLimit {
		http.Error(
			w,
			fmt.Sprintf(entity.BadRequest, "request body is too big"),
			http.StatusBadRequest,
		)
		logger.Error(entity.RequestBodyIsTooBig)
		return
	}
	data, err := io.ReadAll(io.LimitReader(req.Body, h.config.RequestBodyLimit))
	if err != nil {
		http.Error(w, entity.UnableToReadTheBody, http.StatusBadRequest)
		logger.Error(entity.UnableToReadTheBody)
		return
	}
	request, err := mapper.BytesToType[entity.AddUserRequest](data)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf(entity.BadRequest, err),
			http.StatusBadRequest,
		)
		logger.Errorf(entity.BadRequest, err)
		return
	}
	response, err := h.usersCtrl.Add(req.Context(), request)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf(entity.FailedToProcessTheRequest, err),
			http.StatusBadGateway,
		)
		logger.Errorf(entity.FailedToProcessTheRequest, err)
		return
	}
	addUserResponse, err := mapper.TypeToBytes[entity.AddUserResponse](response)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf(entity.FailedToProcessTheResponse, err),
			http.StatusInternalServerError,
		)
		logger.Errorf(entity.FailedToProcessTheResponse, err)
		return // unreachable in tests cause response struct can always be represented as json
	}
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(addUserResponse)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf(entity.FailedToWriteTheResponse, err),
			http.StatusInternalServerError,
		)
		logger.Errorf(entity.FailedToWriteTheResponse, err)
		return // unreachable in tests
	}
	logger.With("response", response).Info("Request completed")
	return
}
