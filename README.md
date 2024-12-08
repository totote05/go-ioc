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
  SingletonService interface {}
  MyService struct {}
)

// Define the services constructor
func NewSingletonService() SingletonService {
  return &MyService{}
}

func NewMyService() *MyService {
  return &MyService{}
}

func main() {
  // Create a container
  container := ioc.NewContainer()

  // Bind singleton service
  container.BindSingleton(NewSingletonService)

  // Bind transient service
  container.BindTransient(NewMyService)

  // Resolve the services
  singletonService := ioc.Resolve[SingletonService](container)
  transientService := ioc.Resolve[*MyService](container)
}
```