package main

import (
	"FinanceGolang/core/api"
	"FinanceGolang/core/dbcore"
	"FinanceGolang/core/settings"
	"fmt"
	"log"
)

func main() {
	// Загрузка конфигурации
	if err := settings.Init(); err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}
	cfg := settings.Get()

	// Инициализация базы данных
	db, err := dbcore.InitDB()
	if err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}
	defer dbcore.CloseDB()

	// Auto migrate models
	err = dbcore.CreateTables(db)
	if err != nil {
		log.Fatalf("Ошибка миграции базы данных: %v", err)
	}

	// Инициализация контроллеров напрямую через Router
	// Router создает все необходимые репозитории и сервисы внутри себя

	// Инициализация контроллеров
	router := api.NewRouter()

	// Настройка Gin и middleware
	r := router.InitRoutes()

	// Запуск сервера
	addr := fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort)
	log.Printf("Сервер запускается на %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
