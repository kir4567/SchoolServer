// posts
package restapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// File struct используется в Post
type File struct {
	File     string `json:"file"`
	FileName string `json:"fileName"`
}

// type Post struct используется в postsResponse
type Post struct {
	Unread  bool   `json:"unread"`
	Author  string `json:"author"`
	Title   string `json:"title"`
	Date    string `json:"date"`
	Message string `json:"message"`
	Files   []File `json:"files"`
}

// postsResponse struct используется в GetPostsHandler
type postsResponse struct {
	Posts []Post `json:"posts"`
}

// GetPostsHandler обрабатывает запрос на получение объявлений
func (rest *RestAPI) GetPostsHandler(respwr http.ResponseWriter, req *http.Request) {
	rest.logger.Info("REST: GetPostsHandler called", "IP", req.RemoteAddr)
	// Проверка метода запроса
	if req.Method != "GET" {
		rest.logger.Info("REST: Wrong method", "Method", req.Method, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Получить существующие имя и объект локальной сессии
	sessionName, session := rest.getLocalSession(respwr, req)
	if session == nil {
		return
	}
	// Получим удаленную сессию
	remoteSession, ok := rest.sessionsMap[sessionName]
	if !ok {
		// Если нет удаленной сессии
		rest.logger.Info("REST: No remote session", "IP", req.RemoteAddr)
		// Создать новую
		remoteSession = rest.remoteRelogin(respwr, req, session)
		if remoteSession == nil {
			return
		}
	}
	userName := session.Values["userName"].(string)
	schoolID := session.Values["schoolID"].(int)
	// Получить объявления
	posts, err := remoteSession.GetAnnouncements(strconv.Itoa(schoolID), rest.config.ServerName)
	if err != nil {
		if strings.Contains(err.Error(), "You was logged out from server") {
			// Если удаленная сессия есть, но не активна
			rest.logger.Info("REST: Remote connection timed out", "IP", req.RemoteAddr)
			// Создать новую
			remoteSession = rest.remoteRelogin(respwr, req, session)
			if remoteSession == nil {
				return
			}
			// Повторно получить с сайта школы
			posts, err = remoteSession.GetAnnouncements(strconv.Itoa(schoolID), rest.config.ServerName)
			if err != nil {
				// Ошибка
				rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
				respwr.WriteHeader(http.StatusBadGateway)
				return
			}
		} else {
			// Другая ошибка
			rest.logger.Error("REST: Error occured when getting data from site", "Error", err, "IP", req.RemoteAddr)
			respwr.WriteHeader(http.StatusBadGateway)
			return
		}
	}
	// Обновить БД
	err = rest.Db.UpdatePosts(userName, schoolID, posts)
	if err != nil {
		rest.logger.Error("REST: Error occured when updating posts in DB", "Error", err, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Сформировать ответ по протоколу
	resp := postsResponse{}
	for _, p := range posts.Posts {
		file := File{p.FileLink, p.FileName}
		newPost := Post{
			Unread:  p.Unread,
			Author:  p.Author,
			Title:   p.Title,
			Date:    p.Date,
			Message: p.Message,
			Files:   []File{file},
		}
		resp.Posts = append(resp.Posts, newPost)
	}
	// Закодировать ответ в JSON
	bytes, err := json.Marshal(resp)
	if err != nil {
		rest.logger.Error("REST: Error occured when marshalling response", "Error", err, "Response", resp, "IP", req.RemoteAddr)
		respwr.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Отправить ответ клиенту
	status, err := respwr.Write(bytes)
	if err != nil {
		rest.logger.Error("REST: Error occured when sending response", "Error", err, "Response", resp, "Status", status, "IP", req.RemoteAddr)
	} else {
		rest.logger.Info("REST: Successfully sent response", "Response", resp, "IP", req.RemoteAddr)
	}
	// Отправить пуш на удаление пушей с объявлениями
	err = rest.pushDelete(userName, schoolID, "new_post")
	if err != nil {
		rest.logger.Error("REST: Error occured when sending deleting push", "Error", err, "Category", "new_post", "IP", req.RemoteAddr)
	}
}
