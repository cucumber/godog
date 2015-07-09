# An example of API with DB

The following example demonstrates steps how we describe and test our API with DB using **godog**.
To start with, see [API example](https://github.com/DATA-DOG/godog/tree/master/examples/api) before.
We have extended it to be used with database.

The interesting point is, that we have [go-txdb](https://github.com/DATA-DOG/go-txdb) library,
which has an implementation of custom sql.driver to allow execute every and each scenario
within a **transaction**. After it completes, transaction is rolled back so the state could
be clean for the next scenario.

To run **users.feature** you need MySQL installed on your system with an anonymous root password.
Then run:

    make test

The json comparisom function should be improved and we should also have placeholders for primary
keys when comparing a json result.
