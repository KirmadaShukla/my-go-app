package router

import (
	"net/http"

	"my-go-app/internal/handler"
	"my-go-app/internal/middleware"
)

func registerTutor(mux *http.ServeMux, d Deps) {
	mux.HandleFunc("GET /tutor/subjects", d.Handler.ListTutorSubjects)

	mux.HandleFunc("POST /tutor/sessions", middleware.Protect(
		d.Tokens,
		middleware.ValidateJSON[handler.StartTutorSessionRequest](d.Handler.StartTutorSession),
	))

	mux.HandleFunc("POST /tutor/sessions/{id}/voice", middleware.Protect(
		d.Tokens,
		d.Handler.TutorVoice,
	))
}
