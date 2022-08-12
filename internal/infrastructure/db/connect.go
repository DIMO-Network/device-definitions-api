package db

import (
	"database/sql"
	"fmt"
	"log"

	intshared "github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/shared"
	_ "github.com/lib/pq"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func Connection(s intshared.Settings) *sql.DB {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		s.DBUser,
		s.DBPassword,
		s.DBName,
		s.DBHost,
		s.DBPort,
	))
	if err != nil {
		log.Fatal("Database Connection Error $s", err)
	}

	boil.SetDB(db)

	return db
}
