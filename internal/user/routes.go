package user

import "net/http"

func SetupUserRoutes(mux *http.ServeMux, userHandler *UserHandler) {

	mux.HandleFunc("POST /api/users/register", userHandler.Register)

}
