// Copyright (C) 2018 Mikhail Masyagin

/*
Package server содержит основную функциональность сервера.
*/
package server

import (
	cp "SchoolServer/libtelco/config-parser"
	"SchoolServer/libtelco/log"

	// ss "SchoolServer/libtelco/sessions"

	api "SchoolServer/libtelco/rest-api"

	"net/http"
	"runtime"

	"github.com/gorilla/context"
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
						// Тесты.
						kek := ss.NewSession(&serv.config.Schools[0])

						err := kek.Login()
						if err != nil {
							fmt.Println(err)
						}
		// Вставь нужное кол-во значений.
						_, err = kek.GetLessonDescription("16.10.2017", "13075", "11198")
						if err != nil {
							fmt.Println(err)
						}

						err = kek.Logout()
						if err != nil {
							fmt.Println(err)
						}
	*/
	// Подключаем handler'ы из RestAPI.
	serv.api.BindHandlers()
	return http.ListenAndServe(serv.config.ServerAddr, context.ClearHandler(http.DefaultServeMux))
}
