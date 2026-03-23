package main

import (
	"log"
	"net/http"
	"os"

	"inventory_ms/internal/infrastructure/database"
	"inventory_ms/internal/usecase"

	httpdelivery "inventory_ms/internal/delivery/http"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")

	db, err := database.NewPostgresDB(dsn)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Close()

	// Repositories
	invRepo := database.NewPostgresInventoryRepo(db)
	resRepo := database.NewPostgresReservationRepo(db)

	// Use cases
	reserveUC := usecase.NewReserveUseCase(invRepo, resRepo)
	returnUC := usecase.NewReturnUseCase(invRepo, resRepo)
	deliverUC := usecase.NewDeliverUseCase(invRepo, resRepo)

	// Delivery
	handler := httpdelivery.NewHandler(reserveUC, returnUC, deliverUC)
	router := httpdelivery.NewRouter(handler)

	addr := ":8080"
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
