package web

import (
	"github.com/wureny/FluxGo/config"
	"github.com/wureny/FluxGo/internal/service"
)

type Tx struct {
}

type APIGuard struct {
	Params *config.Params
	Serv   *service.Serv
}

func NewAPIGuard(p *config.Params, s *service.Serv) *APIGuard {
	return &APIGuard{
		Params: p,
		Serv:   s,
	}
}
