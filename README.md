# go-ioc

Es una librería para pode entender y trabajar con la inversión de control en Go.

## Instalación

```bash
go get github.com/totote05/go-ioc
```

## Uso

```go
package main

import (
  "fmt"
	"go-ioc.totote05.ar/pkg/ioc"
)

// Define the services
type (
  MyService        struct {}
  SingletonService any
  MyHandler        ioc.Handler
)

// Implement the handler interface
func (t *MyService) Handle(w http.ResponseWriter, r *http.Request) {}

// Define the services constructor
func NewMyService() *MyService {
  return &MyService{}
}

func NewSingletonService() SingletonService {
  return &MyService{}
}

func NewHandler() MyHandler {
	return &MyService{}
}

func main() {
  // Bind singleton service
  ioc.BindSingleton(NewSingletonService)

  // Bind transient service
  ioc.BindTransient(NewMyService)

  // Bind handler
  ioc.BindTransient(NewHandler)

  // Resolve the services
  singletonService := ioc.Resolve[SingletonService]()
  transientService := ioc.Resolve[*MyService]()

  // Resolve the handler with HandleFunc
  http.HandleFunc("GET /", ioc.ResolveHanlder[MyHandler])
}
```