package handlers

import (
	"api/internal/logger"
	"api/internal/messages"
	"api/internal/response"
	"html/template"
	"net/http"
)

// serveHTML обрабатывает запрос на отдачу HTML страницы
func serveHTML(w http.ResponseWriter, r *http.Request, filename string) {
	tmpl, err := template.ParseFiles("assets/html/" + filename)
	if err != nil {
		logger.Error(messages.ServiceStatic, messages.LogErrLoadTemplate, map[string]string{
			messages.LogDetails:  err.Error(),
			messages.LogReqPath:  r.URL.Path,
			messages.LogFilename: filename,
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrPageLoad, nil)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		logger.Error(messages.ServiceStatic, messages.LogErrRenderTemplate, map[string]string{
			messages.LogDetails:  err.Error(),
			messages.LogReqPath:  r.URL.Path,
			messages.LogFilename: filename,
		})
		response.WriteAPIResponse(w, http.StatusInternalServerError, false, messages.ClientErrPageLoad, nil)
		return
	}

	logger.Info(messages.ServiceStatic, messages.LogStatusPageServed, map[string]string{
		messages.LogReqPath:  r.URL.Path,
		messages.LogFilename: filename,
	})
}

// OutIndex отдает главную страницу
func OutIndex(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, r, "index.html")
}

// OutRegister отдает страницу регистрации
func OutRegister(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, r, "register.html")
}

// OutLogin отдает страницу входа
func OutLogin(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, r, "login.html")
}

// OutListing отдает страницу с созданием объявления
func OutListing(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, r, "listing.html")
}

// OutEdit отдает страницу редактирования объявления
func OutEdit(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, r, "edit.html")
}
