// Copyright (C) 2018 Mikhail Masyagin

package main

import (
	"fmt"
	"net/http"
	"os"

	cp "github.com/masyagin1998/SchoolServer/libtelco/config-parser"
	"github.com/masyagin1998/SchoolServer/libtelco/log"
	"github.com/masyagin1998/SchoolServer/libtelco/server"
)

var (
	// Конфиг сервера.
	config *cp.Config
	// Логгер.
	logger *log.Logger
	// Стандартная ошибка.
	err error
)

// init производит:
// - чтение конфигурационных файлов;
// - создание логгера;
func init() {
	if config, err = cp.ReadConfig(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if logger, err = log.NewLogger(config.LogFile); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

}

func main() {
	// Вся информация о конфиге.
	logger.Info("SchoolServer V0.1 is running",
		"Server address", config.ServerAddr,
		"Postgres info", config.Postgres,
		"Max allowed threads", config.MaxProcs,
		"LogFile", config.LogFile,
	)
	// Вся информация о списке серверов.
	logger.Info("List of schools")
	for _, school := range config.Schools {
		logger.Info("School",
			"Name", school.Name,
			"Type", school.Type,
			"Link", school.Link,
			"Permission", school.Permission,
		)
	}

	// Запуск сервера.
	server := server.NewServer(config, logger)
	if err := server.Run(); err != http.ErrServerClosed {
		logger.Error("Fatal error occured, while running server", "error", err)
	} else {
		logger.Info("Server was successfully shutdowned")
	}
}
