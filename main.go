package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var tmpl *template.Template
var db *sql.DB

type Task struct {
	Id   int
	Task string
	Done bool
}

func init() {
	tmpl, _ = template.ParseGlob("templates/*.html")

}

func initDB() {
	var err error
	// Initialize the db variable
	db, err = sql.Open("mysql", "root:root@(127.0.0.1:3333)/testdb?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}

	// Check the database connection
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
}

func main() {

	gRouter := mux.NewRouter()

	//Setup MySQL
	initDB()
	defer db.Close()

	gRouter.HandleFunc("/", Homepage)

	//Get Tasks
	gRouter.HandleFunc("/tasks", fetchTasks).Methods("GET")

	//Add Task
	gRouter.HandleFunc("/tasks", addTask).Methods("POST")

	http.ListenAndServe(":4000", gRouter)

}

func Homepage(w http.ResponseWriter, r *http.Request) {

	tmpl.ExecuteTemplate(w, "home.html", nil)

}

func fetchTasks(w http.ResponseWriter, r *http.Request) {
	todos, _ := getTasks(db)
	//fmt.Println(todos)

	//If you used "define" to define the template, use the name you gave it here, not the filename
	tmpl.ExecuteTemplate(w, "todoList", todos)
}

func addTask(w http.ResponseWriter, r *http.Request) {

	task := r.FormValue("task")

	fmt.Println(task)

	query := "INSERT INTO tasks (task, done) VALUES (?, ?)"

	stmt, err := db.Prepare(query)

	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, executeErr := stmt.Exec(task, 0)

	if executeErr != nil {
		log.Fatal(executeErr)
	}

	// Return a new list of Todos
	todos, _ := getTasks(db)

	//You can also just send back the single task and append it
	//I like returning the whole list just to get everything fresh, but this might not be the best strategy
	tmpl.ExecuteTemplate(w, "todoList", todos)

}

func getTasks(dbPointer *sql.DB) ([]Task, error) {

	query := "SELECT id, task, done FROM tasks"

	rows, err := dbPointer.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var tasks []Task

	for rows.Next() {
		var todo Task

		rowErr := rows.Scan(&todo.Id, &todo.Task, &todo.Done)

		if rowErr != nil {
			return nil, err
		}

		tasks = append(tasks, todo)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil

}
