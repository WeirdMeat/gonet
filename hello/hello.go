package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"slices"
	"strconv"
)

type Book struct {
	year   int
	author string
	title  string
	//rating Star
}

var library = []Book{
	{year: 1813, author: "Jane Austen", title: "Pride and Prejudice"},
	{year: 1960, author: "Harper Lee", title: "To Kill A Mockingbird"},
	{year: 1925, author: "F. Scott Fitzgerald", title: "The Great Gatsby"},
	{year: 1967, author: "Gabriel García", title: "One Hundred Years of Solitude"},
	{year: 1901, author: "Thomas Mann", title: "Buddenbrooks"},
}

func (b *Book) String() string {
	return fmt.Sprintf("%s by %s(%d)", b.title, b.author, b.year)
}

func AllBooks(w http.ResponseWriter, r *http.Request) {
	for _, b := range library {
		fmt.Fprintf(w, "%s\n", &b)
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
	book := Book{year: year, author: author, title: title}
	library = append(library, book)
	fmt.Fprintf(w, "you created a book: %s", &book)
}

func ReadBook(w http.ResponseWriter, r *http.Request) {
	title := mux.Vars(r)["title"]
	idx := slices.IndexFunc(library, func(b Book) bool { return b.title == title })
	if idx == -1 {
		fmt.Fprintf(w, "there are no such book")
	} else {
		fmt.Fprintf(w, "you've chosen a %s", &library[idx])
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
	idx := slices.IndexFunc(library, func(b Book) bool { return b.title == title })
	if idx == -1 {
		fmt.Fprintf(w, "there are no such book")
	} else {
		library[idx] = Book{year: year, author: author, title: title}
		fmt.Fprintf(w, "you updated a %s book", title)
	}
}

func DeleteBook(w http.ResponseWriter, r *http.Request) {
	title := mux.Vars(r)["title"]
	library_len := len(library)
	library = slices.DeleteFunc(library, func(b Book) bool { return b.title == title })
	if library_len != len(library) {
		fmt.Fprintf(w, "you deleted a %s book", title)
	} else {
		fmt.Fprintf(w, "no such book")
	}
}

func main() {
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

	err := http.ListenAndServe(":80", r)
	if err != nil {
		panic(err)
	}
}
