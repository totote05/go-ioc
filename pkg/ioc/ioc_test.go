package ioc_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go-ioc.totote05.ar/pkg/ioc"
)

type (
	MyService struct {
		Name string
	}
	complexService struct {
		SingletonService SingletonService
		TransientService TransientService
	}
	TestService interface {
		GetName() string
	}
	SingletonService TestService
	TransientService TestService
)

func (t *MyService) GetName() string {
	return t.Name
}

func NewMyService() *MyService {
	return &MyService{}
}

func NewTestService() TestService {
	return &MyService{Name: time.Now().Format(time.RFC3339Nano)}
}

func NewSingletonService() SingletonService {
	return NewTestService()
}

func NewTransientService() TransientService {
	return NewTestService()
}

func NewComplexService(singletonService SingletonService, transientService TransientService) *complexService {
	return &complexService{
		SingletonService: singletonService,
		TransientService: transientService,
	}
}

func TestUnboundType(t *testing.T) {
	assert.PanicsWithValue(t, "Unbound type: ioc_test.TestService", func() {
		container := ioc.NewContainer()
		ioc.Resolve[TestService](container)
	})
}

func TestAlreadyBoundType(t *testing.T) {
	assert.PanicsWithValue(t, "Type already bound: ioc_test.SingletonService", func() {
		container := ioc.NewContainer()
		container.BindSingleton(NewSingletonService)
		container.BindSingleton(NewSingletonService)
	})
}

func TestConstructorIsNotAFunction(t *testing.T) {
	assert.PanicsWithValue(t, "Constructor must be a function", func() {
		container := ioc.NewContainer()
		container.BindSingleton(&MyService{})
	})
}

func TestConstructorMustReturnSingleValue(t *testing.T) {
	assert.PanicsWithValue(t, "Constructor must return a single value", func() {
		container := ioc.NewContainer()
		container.BindSingleton(func() (TestService, error) {
			return nil, nil
		})
	})
}

func TestSingletonBinding(t *testing.T) {
	assert.NotPanics(t, func() {
		container := ioc.NewContainer()
		container.BindSingleton(NewSingletonService)

		instance1 := ioc.Resolve[SingletonService](container)
		instance2 := ioc.Resolve[SingletonService](container)

		if instance1 != instance2 {
			t.Errorf("Expected instance1 and instance2 to be equal, but got %v and %v", instance1, instance2)
		}
		assert.Equal(t, instance1, instance2)
	})
}

func TestTransientBinding(t *testing.T) {
	assert.NotPanics(t, func() {
		container := ioc.NewContainer()
		container.BindTransient(NewTransientService)

		instance1 := ioc.Resolve[TransientService](container)
		instance2 := ioc.Resolve[TransientService](container)

		if instance1 == instance2 {
			t.Errorf("Expected instance1 and instance2 to be different, but got %v and %v", instance1, instance2)
		}
		assert.NotEqual(t, instance1, instance2)
	})
}

func TestConcurrentAccess(t *testing.T) {
	assert.NotPanics(t, func() {
		container := ioc.NewContainer()
		container.BindSingleton(NewSingletonService)
		instance := ioc.Resolve[SingletonService](container)

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				newInstance := ioc.Resolve[SingletonService](container)
				assert.Equal(t, instance, newInstance)
			}()
		}
		wg.Wait()
	})
}

func TestComplexService(t *testing.T) {
	assert.NotPanics(t, func() {
		container := ioc.NewContainer()

		container.BindSingleton(NewComplexService)
		container.BindSingleton(NewSingletonService)
		container.BindTransient(NewTransientService)

		instance := ioc.Resolve[*complexService](container)

		assert.NotNil(t, instance)
		assert.NotNil(t, instance.SingletonService)
		assert.NotNil(t, instance.TransientService)
	})
}

func TestREADMEExample(t *testing.T) {
	assert.NotPanics(t, func() {
		// Create a container
		container := ioc.NewContainer()

		// Bind singleton service
		container.BindSingleton(NewSingletonService)

		// Bind transient service
		container.BindTransient(NewMyService)

		// Resolve the services
		singletonService := ioc.Resolve[SingletonService](container)
		singletonService2 := ioc.Resolve[SingletonService](container)
		transientService := ioc.Resolve[*MyService](container)
		transientService2 := ioc.Resolve[*MyService](container)

		// Assert that the singleton services are the same
		if singletonService != singletonService2 {
			t.Errorf("Expected singletonService and singletonService2 to be equal, but got %v and %v", singletonService, singletonService2)
		}

		// Assert that the transient services are different
		if transientService == transientService2 {
			t.Errorf("Expected transientService and transientService2 to be different, but got %v and %v", transientService, transientService2)
		}
	})
}
