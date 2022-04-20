module github.com/cucumber/godog/_examples

go 1.16

replace github.com/cucumber/godog => ../

require (
	github.com/DATA-DOG/go-txdb v0.1.4
	github.com/cucumber/godog v0.0.0-00010101000000-000000000000
	github.com/go-sql-driver/mysql v1.6.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.1
)

require (
	github.com/kr/pretty v0.3.0 // indirect
	github.com/lib/pq v1.10.3 // indirect
)
