# Gin Cors Middleware

The gin-cors package can be used in every [Gin](https://github.com/gin-gonic/gin) project and configure the cors behaviour of your application.

## Getting started
After installing and setting up Go you can create yor first project with gin.

In your main go file create a basic gin http server with the gin-cors middleware package
```go
package main

import (
  "net/http"

  "github.com/gin-gonic/gin"
  cors "github.com/OnlyNico43/gin-cors/v2"
)

func main() {
  r := gin.Default()

  r.Use(cors.Middleware(cors.DefaultConfig()))

  r.GET("/ping", func(c *gin.Context) {
    c.String(http.StatusOK, "pong")
  })

  r.Run(":8080")
}
```

Done, now your application is equipped with a basic cors configuration


## How to configure
You can either use the default configuration or decide to use your own values in the Config object.

For more details please see the [documentation](https://pkg.go.dev/github.com/bytewise43/gin-cors)

## Contributing
If you want to contribute to this project, please read the [contribution guidelines](CONTRIBUTION.md) first. I welcome any contributions, whether it's fixing bugs, adding new features, or improving documentation.

## Having issues?
If you encounter any issues while using this package, please feel free to open an issue in the GitHub repository. I will do my best to address it as soon as possible.
