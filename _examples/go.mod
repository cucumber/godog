module github.com/cucumber/godog/_examples

go 1.24

toolchain go1.24.2

replace github.com/cucumber/godog => ../

require (
	github.com/DATA-DOG/go-txdb v0.2.1
	github.com/cucumber/godog v0.15.0
	github.com/go-sql-driver/mysql v1.9.3
	github.com/spf13/pflag v1.0.6
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/kr/pretty v0.3.0 // indirect
)
