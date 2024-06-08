package constant

type dbConstKey string

const (
	TxKey         = dbConstKey("TxKey")
	MYSQL         = "mysql"
	POSTGRES      = "postgres"
	POSTGRES_CONN = `host=%s port=%d user=%s password=%s dbname=%s sslmode=disable`
)
