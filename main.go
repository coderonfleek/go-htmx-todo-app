package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

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

	//Fetch Add Task Form
	gRouter.HandleFunc("/newtaskform", getTaskForm)

	//Add Task
	gRouter.HandleFunc("/tasks", addTask).Methods("POST")

	//Fetch Update Form
	gRouter.HandleFunc("/gettaskupdateform/{id}", getTaskUpdateForm).Methods("GET")

	//Update Task
	gRouter.HandleFunc("/tasks/{id}", updateTask).Methods("PUT", "POST")

	//Delete Task
	gRouter.HandleFunc("/tasks/{id}", deleteTask).Methods("DELETE")

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

func getTaskForm(w http.ResponseWriter, r *http.Request) {

	tmpl.ExecuteTemplate(w, "addTaskForm", nil)
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

func getTaskUpdateForm(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	//Convert string id from URL to integer
	taskId, _ := strconv.Atoi(vars["id"])

	task, err := getTaskByID(db, taskId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	tmpl.ExecuteTemplate(w, "updateTaskForm", task)

}

func updateTask(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	taskItem := r.FormValue("task")
	//taskStatus, _ := strconv.ParseBool(r.FormValue("done"))
	var taskStatus bool

	fmt.Println(r.FormValue("done"))

	//Check the string value of the checkbox
	switch strings.ToLower(r.FormValue("done")) {
	case "yes", "on":
		taskStatus = true
	case "no", "off":
		taskStatus = false
	default:
		taskStatus = false
	}

	taskId, _ := strconv.Atoi(vars["id"])

	task := Task{
		taskId, taskItem, taskStatus,
	}

	updateErr := updateTaskById(db, task)

	if updateErr != nil {
		log.Fatal(updateErr)
	}

	//Refresh all Tasks
	todos, _ := getTasks(db)

	tmpl.ExecuteTemplate(w, "todoList", todos)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	taskId, _ := strconv.Atoi(vars["id"])

	err := deleTaskWithID(db, taskId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	//Return list
	todos, _ := getTasks(db)

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

func getTaskByID(dbPointer *sql.DB, id int) (*Task, error) {

	query := "SELECT id, task, done FROM tasks WHERE id = ?"

	var task Task

	row := dbPointer.QueryRow(query, id)
	err := row.Scan(&task.Id, &task.Task, &task.Done)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("No task was found with task %d", id)
		}
		return nil, err
	}

	return &task, nil

}

func updateTaskById(dbPointer *sql.DB, task Task) error {

	query := "UPDATE tasks SET task = ?, done = ? WHERE id = ?"

	result, err := dbPointer.Exec(query, task.Task, task.Done, task.Id)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		fmt.Println("No rows updated")
	} else {
		fmt.Printf("%d row(s) updated\n", rowsAffected)
	}

	return nil

}

func deleTaskWithID(dbPointer *sql.DB, id int) error {

	query := "DELETE FROM tasks WHERE id = ?"

	stmt, err := dbPointer.Prepare(query)

	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no task found with id %d", id)
	}

	fmt.Printf("Deleted %d task(s)\n", rowsAffected)
	return nil

}
