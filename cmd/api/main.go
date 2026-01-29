package main

import (
	"fmt"
	"net/http"

	"github.com/anglesson/simple-web-server/internal/ebook"
	"github.com/anglesson/simple-web-server/internal/platform/uuid"
	"github.com/anglesson/simple-web-server/pkg/database"
	"github.com/go-chi/chi/v5"
)

// chiAdapter transforma o chi.Router na nossa interface router.Router
type chiAdapter struct {
	r chi.Router
}

func (a *chiAdapter) Handle(method, path string, h http.HandlerFunc) {
	a.r.Method(method, path, h)
}

func main() {
	mainRouter := chi.NewRouter()

	uuid := uuid.NewGoogleUUID()

	mainRouter.Route("/api/v1", func(r chi.Router) {
		adapter := &chiAdapter{r: r}
		ebook.Setup(database.DB, uuid, adapter)
	})

	fmt.Println("Server starting...")
	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		fmt.Printf("%-7s %s\n", method, route)
		return nil
	}

	if err := chi.Walk(mainRouter, walkFunc); err != nil {
		fmt.Printf("Logging err: %s\n", err.Error())
	}
	http.ListenAndServe(":3000", mainRouter)
}
