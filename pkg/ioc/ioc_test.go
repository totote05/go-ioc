package ioc_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/totote05/go-ioc/pkg/ioc"
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
	MyHandler        ioc.Handler
)

func init() {
	// bind singleton
	ioc.BindSingleton(NewSingletonService)

	// bind transient
	ioc.BindTransient(NewTransientService)

	// bind handler
	ioc.BindTransient(NewHandler)
}

func (t *MyService) GetName() string {
	return t.Name
}

func (t *MyService) Handle(w http.ResponseWriter, r *http.Request) {}

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

func NewHandler() MyHandler {
	return &MyService{}
}

func TestUnboundType(t *testing.T) {
	assert.PanicsWithValue(t, "Unbound type: ioc_test.TestService", func() {
		ioc.Resolve[TestService]()
	})
}

func TestAlreadyBoundType(t *testing.T) {
	assert.PanicsWithValue(t, "Type already bound: ioc_test.SingletonService", func() {
		ioc.BindSingleton(NewSingletonService)
		ioc.BindSingleton(NewSingletonService)
	})
}

func TestConstructorIsNotAFunction(t *testing.T) {
	assert.PanicsWithValue(t, "Constructor must be a function", func() {
		ioc.BindSingleton(&MyService{})
	})
}

func TestConstructorMustReturnSingleValue(t *testing.T) {
	assert.PanicsWithValue(t, "Constructor must return a single value", func() {
		ioc.BindSingleton(func() (TestService, error) {
			return nil, nil
		})
	})
}

func TestSingletonBinding(t *testing.T) {
	assert.NotPanics(t, func() {
		instance1 := ioc.Resolve[SingletonService]()
		instance2 := ioc.Resolve[SingletonService]()

		if instance1 != instance2 {
			t.Errorf("Expected instance1 and instance2 to be equal, but got %v and %v", instance1, instance2)
		}
		assert.Equal(t, instance1, instance2)
	})
}

func TestTransientBinding(t *testing.T) {
	assert.NotPanics(t, func() {
		instance1 := ioc.Resolve[TransientService]()
		instance2 := ioc.Resolve[TransientService]()

		if instance1 == instance2 {
			t.Errorf("Expected instance1 and instance2 to be different, but got %v and %v", instance1, instance2)
		}
		assert.NotEqual(t, instance1, instance2)
	})
}

func TestConcurrentAccess(t *testing.T) {
	assert.NotPanics(t, func() {
		instance := ioc.Resolve[SingletonService]()

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				newInstance := ioc.Resolve[SingletonService]()
				assert.Equal(t, instance, newInstance)
			}()
		}
		wg.Wait()
	})
}

func TestComplexService(t *testing.T) {
	assert.NotPanics(t, func() {
		ioc.BindSingleton(NewComplexService)

		instance := ioc.Resolve[*complexService]()

		assert.NotNil(t, instance)
		assert.NotNil(t, instance.SingletonService)
		assert.NotNil(t, instance.TransientService)
	})
}

func TestResolveHandler(t *testing.T) {
	assert.NotPanics(t, func() {
		req, err := http.NewRequest("GET", "/", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()

		ioc.ResolveHanlder[MyHandler](rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestREADMEExample(t *testing.T) {
	assert.NotPanics(t, func() {
		// Bind singleton service
		// ioc.BindSingleton(NewSingletonService)

		// Bind transient service
		ioc.BindTransient(NewMyService)

		// Bind handler
		// ioc.BindTransient(NewHandler)

		// Resolve the services
		singletonService := ioc.Resolve[SingletonService]()
		singletonService2 := ioc.Resolve[SingletonService]()
		transientService := ioc.Resolve[*MyService]()
		transientService2 := ioc.Resolve[*MyService]()

		// Assert that the singleton services are the same
		if singletonService != singletonService2 {
			t.Errorf("Expected singletonService and singletonService2 to be equal, but got %v and %v", singletonService, singletonService2)
		}

		// Assert that the transient services are different
		if transientService == transientService2 {
			t.Errorf("Expected transientService and transientService2 to be different, but got %v and %v", transientService, transientService2)
		}

		// Create a new request
		req, err := http.NewRequest("GET", "/", nil)
		assert.NoError(t, err)

		// Create a new response recorder
		rr := httptest.NewRecorder()

		// Resolve the handler and call the Handle method
		ioc.ResolveHanlder[MyHandler](rr, req)

		// Assert that the response status code is 200
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
