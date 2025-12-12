package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/Daankoo/lab5_PT/db/sqlc"
	"github.com/Daankoo/lab5_PT/internal/server"
)

func main() {
	connPool, err := pgxpool.New(context.Background(), "postgresql://postgres:postgres@localhost:5432/api_db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer connPool.Close()

	store := db.NewStore(connPool)

	server := server.NewServer(store)
	server.Run(":3000")
}
