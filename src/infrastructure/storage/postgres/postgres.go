package postgres

import (
	"database/sql"
	"time"

	"github.com/spf13/viper"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func InitPostgres() *bun.DB {
	addr := viper.GetString("db.host") + viper.GetString("db.port")
	user := viper.GetString("db.user")
	password := viper.GetString("db.pass")
	dbName := viper.GetString("db.name")
	appName := viper.GetString("server_name")

	pgConn := pgdriver.NewConnector(
		pgdriver.WithNetwork("tcp"),
		// Disable SSL
		pgdriver.WithInsecure(true),
		pgdriver.WithAddr(addr),
		pgdriver.WithUser(user),
		pgdriver.WithPassword(password),
		pgdriver.WithDatabase(dbName),
		pgdriver.WithApplicationName(appName),
		pgdriver.WithTimeout(5*time.Second),
		pgdriver.WithDialTimeout(5*time.Second),
		pgdriver.WithReadTimeout(5*time.Second),
		pgdriver.WithWriteTimeout(5*time.Second),
		// Set to all connections server timezone "SET key TO val;"
		pgdriver.WithConnParams(map[string]interface{}{
			"timezone": viper.GetString("server_timezone"),
		}),
	)

	//// Pgx driver
	//config, err := pgx.ParseConfig(fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, password, addr, dbName))
	//if err != nil {
	//	panic(err)
	//}
	//config.PreferSimpleProtocol = true
	//
	//sqldb := stdlib.OpenDB(*config)

	// New bun
	// With Discard Unknown Columns
	db := bun.NewDB(sql.OpenDB(pgConn), pgdialect.New(), bun.WithDiscardUnknownColumns())
	if err := db.Ping(); err != nil {
		panic(err)
	}

	// Db logger
	// go get github.com/uptrace/bun/extra/bundebug
	//db.AddQueryHook(bundebug.NewQueryHook(
	//	bundebug.WithVerbose(true),
	//	bundebug.FromEnv("BUNDEBUG"),
	//))

	RegisterM2M(db)

	return db
}

func RegisterM2M(db *bun.DB) {
	// Register many to many model so bun can better recognize m2m relation.
	// This should be done before you use the model for the first time.

}
