package app

import (
	ticketHttp "github.com/NamespaceManager/internal/ticket/delivery/http"
	ticketRepository "github.com/NamespaceManager/internal/ticket/repository"
	ticketUsecase "github.com/NamespaceManager/internal/ticket/usecase"

	usersHttp "github.com/NamespaceManager/internal/users/delivery/http"
	usersRepository "github.com/NamespaceManager/internal/users/repository"
	usersUsecase "github.com/NamespaceManager/internal/users/usecase"

	nsHTTP "github.com/NamespaceManager/internal/namespace/delivery/http"
	nsRepository "github.com/NamespaceManager/internal/namespace/repository"
	nsUsecase "github.com/NamespaceManager/internal/namespace/usecase"
)

func (a *App) MapHandlers() error {
	ticketGroup := a.gin.Group("/ticket")
	usersGroup := a.gin.Group("/users")
	nsGroup := a.gin.Group("/ns")

	ticketRepository := ticketRepository.NewTicketRepository(a.postgresDB)
	usersRepository := usersRepository.NewUsersRepository(a.postgresDB)
	nsRepository := nsRepository.NewUsersRepository(a.postgresDB)

	ticketUsecase := ticketUsecase.NewTicketUsecase(ticketRepository)
	usersUsecase := usersUsecase.NewUsersUsecase(usersRepository)
	nsUsecase := nsUsecase.NewNSUsecase(nsRepository)

	ticketHandlers := ticketHttp.NewTicketHandler(ticketUsecase)
	usersHandlers := usersHttp.NewUsersHandler(usersUsecase)
	nsHandlers := nsHTTP.NewNSHandler(nsUsecase)

	ticketHttp.MapTicketRoutes(ticketGroup, ticketHandlers)
	usersHttp.MapUsersRoutes(usersGroup, usersHandlers)
	nsHTTP.MapNSRoutes(nsGroup, nsHandlers)

	return nil
}
