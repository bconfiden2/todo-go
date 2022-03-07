package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
)

var (
	rd *render.Render
	todoMap map[int]Todo
	lastID int = 0
)

type Success struct {
	Success		bool	`json:"success"`
}

type Todo struct {
	ID			int		`json:"id,omitempty"`
	Name		string	`json:"name"`
	Completed	bool	`json:"completed,omitempty"`
}

type Todos []Todo
func (t Todos) Len() int {
	return len(t)
}
func (t Todos) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}
func (t Todos) Less(i, j int) bool {
	return t[i].ID > t[j].ID
}

func GetTodoHandler(w http.ResponseWriter, r *http.Request) {
	list := make(Todos, 0)
	for _, todo := range todoMap {
		list = append(list, todo)
	}
	sort.Sort(list)
	rd.JSON(w, http.StatusOK, list)
}

func PostTodoHandler(w http.ResponseWriter, r *http.Request) {
	var todo Todo
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	lastID++
	todo.ID = lastID
	todoMap[lastID] = todo
	rd.JSON(w, http.StatusCreated, todo)
}

func PutTodoHandler(w http.ResponseWriter, r *http.Request) {
	var newTodo Todo
	err := json.NewDecoder(r.Body).Decode(&newTodo)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	if todo, ok := todoMap[id]; ok {
		todo.Name = newTodo.Name
		todo.Completed = newTodo.Completed
		rd.JSON(w, http.StatusOK, Success{true})
	} else {
		rd.JSON(w, http.StatusBadRequest, Success{false})
	}
}

func DeleteTodoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	if _, ok := todoMap[id]; ok {
		delete(todoMap, id)
		rd.JSON(w, http.StatusOK, Success{true})
	} else {
		rd.JSON(w, http.StatusNotFound, Success{false})
	}
}

func MakeWebHandler() http.Handler {
	todoMap = make(map[int]Todo)
	mux := mux.NewRouter()
	mux.Handle("/", http.FileServer(http.Dir("public")))
	mux.HandleFunc("/todo", GetTodoHandler).Methods("GET")
	mux.HandleFunc("/todo", PostTodoHandler).Methods("POST")
	mux.HandleFunc("/todo/{id:[0-9]+}", PutTodoHandler).Methods("PUT")
	mux.HandleFunc("/todo/{id:[0-9]+}", DeleteTodoHandler).Methods("DELETE")
	return mux
}

func main() {
	rd = render.New()
	m := MakeWebHandler()
	n := negroni.Classic()
	n.UseHandler(m)

	if err := http.ListenAndServe(":12345", n); err != nil {
		panic(err)
	}
}