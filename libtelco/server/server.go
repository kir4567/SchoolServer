// Copyright (C) 2018 Mikhail Masyagin

/*
Package server содержит основную функциональность сервера.
*/
package server

import (
	cp "SchoolServer/libtelco/config-parser"
	"SchoolServer/libtelco/log"
	api "SchoolServer/libtelco/rest-api"
	"net/http"
	//ss "SchoolServer/libtelco/sessions"

	"runtime"
)

// Server struct содержит конфигурацию сервера.
type Server struct {
	config *cp.Config
	api    *api.RestAPI
	logger *log.Logger
}

// NewServer создает новый сервер.
func NewServer(config *cp.Config, logger *log.Logger) *Server {
	serv := &Server{
		config: config,
		api:    api.NewRestAPI(logger, config),
	}
	return serv
}

// Run запускает сервер.
func (serv *Server) Run() error {
	// Задаем максимальное количество потоков.
	runtime.GOMAXPROCS(serv.config.MaxProcs)

	/*
		// TODO: протестировать все Get'ы.

		s := ss.NewSession(&serv.config.Schools[1])
		err := s.Login()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		data, err := s.GetLessonsMap("10684")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(data)
	*/

	// Подключаем handler'ы из RestAPI.
	serv.api.BindHandlers()
	return http.ListenAndServe(serv.config.ServerAddr, nil)
}
