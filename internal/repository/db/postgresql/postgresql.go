package postgresql

import (
	_ "github.com/jackc/pgx/v5/stdlib"
)

/*
func CheckConn() error {
	ps := fmt.Sprint("host=localhost port=5432 user=videos password=userpassword dbname=videos sslmode=disable")

	db, err := sql.Open("pgx", ps)
	if err != nil {
		return err
	}
	defer db.Close()
	// ...

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		panic(err)
	}
	return nil
}

*/
