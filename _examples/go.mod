module github.com/cucumber/godog/_examples

go 1.16

replace github.com/cucumber/godog => ../

require (
	github.com/DATA-DOG/go-txdb v0.1.8
	github.com/cucumber/godog v0.14.0
	github.com/go-sql-driver/mysql v1.7.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.2
)

require github.com/kr/pretty v0.3.0 // indirect
