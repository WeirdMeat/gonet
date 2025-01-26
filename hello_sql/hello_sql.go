package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"strconv"
)

var db *sql.DB

type Book struct {
	year   int
	author string
	title  string
	//rating Star
}

func (b *Book) String() string {
	return fmt.Sprintf("%s by %s(%d)", b.title, b.author, b.year)
}

func AllBooks(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("select year, author, title from books")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var year int
		var author string
		var title string
		err = rows.Scan(&year, &author, &title)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(w, "%s\n", &Book{year: year, author: author, title: title})
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}

func CreateBook(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	title := mux.Vars(r)["title"]
	year, err := strconv.Atoi(r.FormValue("year"))
	if err != nil {
		panic(err)
	}
	author := r.FormValue("author")
	_, err = db.Exec("insert into books(year, author, title) values(?, ?, ?)", year, author, title)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "you created a book: %s", &Book{year: year, author: author, title: title})
}

func ReadBook(w http.ResponseWriter, r *http.Request) {
	title := mux.Vars(r)["title"]
	stmt, err := db.Prepare("select year, title, author from books where title = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	var year int
	var author string
	err = stmt.QueryRow(title).Scan(&year, &title, &author)
	if err != nil {
		fmt.Fprintf(w, "there are no such book")
	} else {
		fmt.Fprintf(w, "you've chosen a %s", &Book{year: year, author: author, title: title})
	}
}

func UpdateBook(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	title := mux.Vars(r)["title"]
	year, err := strconv.Atoi(r.FormValue("year"))
	if err != nil {
		panic(err)
	}
	author := r.FormValue("author")
	_, err = db.Exec(`update books set author = ?, year = ? WHERE title = ?`, author, year, title)
	if err != nil {
		fmt.Fprintf(w, "no such book")
	} else {
		fmt.Fprintf(w, "you updated a %s book", title)
	}
}

func DeleteBook(w http.ResponseWriter, r *http.Request) {
	title := mux.Vars(r)["title"]
	_, err := db.Exec(`DELETE FROM books WHERE title = ?`, title)
	if err != nil {
		fmt.Fprintf(w, "no such book")
	} else {
		fmt.Fprintf(w, "you deleted a %s book", title)
	}
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./library.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	sqlStmt := `
	drop table IF EXISTS books;
	create table books (year integer, author nvarchar(255), title nvarchar(255));
	delete from books;
	insert into books(year, author, title) values
	( 1813, 'Jane Austen', 'Pride and Prejudice'),
	( 1960, 'Harper Lee', 'To Kill A Mockingbird'),
	( 1925, 'F. Scott Fitzgerald', 'The Great Gatsby'),
	( 1967, 'Gabriel García', 'One Hundred Years of Solitude'),
	( 1901, 'Thomas Mann', 'Buddenbrooks');
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
		return
	}

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
	})
	bookrouter := r.PathPrefix("/books").Subrouter()
	bookrouter.HandleFunc("/", AllBooks)
	bookrouter.HandleFunc("/{title}", CreateBook).Methods("POST")
	bookrouter.HandleFunc("/{title}", ReadBook).Methods("GET")
	bookrouter.HandleFunc("/{title}", UpdateBook).Methods("PUT")
	bookrouter.HandleFunc("/{title}", DeleteBook).Methods("DELETE")

	r.HandleFunc("/books/{title}/page/{page}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		title := vars["title"]
		page := vars["page"]
		fmt.Fprintf(w, "You've requested the book: %s on page %s\n", title, page)
	})

	fs := http.FileServer(http.Dir("static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	err = http.ListenAndServe(":80", r)
	if err != nil {
		panic(err)
	}
}
