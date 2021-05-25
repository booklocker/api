package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

var db *sql.DB

type Book struct {
	Id           uint   `json:"id"`
	Name         string `json:"name"`
	Author       string `json:"author"`
	ChapterCount uint   `json:"chapter_count"`
}

type Chapter struct {
	Number   uint   `json:"number"`
	Name     string `json:"name"`
	Contents string `json:"contents,omitempty"`
}

func getBooks(w http.ResponseWriter, _ *http.Request) {
	res, err := db.Query("SELECT b.book_id, b.book_name, b.author_name, COUNT(c.chapter_no) AS chapter_count FROM books b INNER JOIN chapters c ON b.book_id = c.book_id GROUP BY b.book_id")
	if err != nil {
		fmt.Fprintf(w, "{\"error\": \"Internal MySQL Error\"}")
		fmt.Println("MySQL Error when fetching book list: " + err.Error())
		return
	}

	var books []Book

	for res.Next() {
		var book Book
		err = res.Scan(&book.Id, &book.Name, &book.Author, &book.ChapterCount)
		if err != nil {
			fmt.Fprintf(w, "{\"error\": \"Internal MySQL Error\"}")
			fmt.Println("MySQL Error when parsing book: " + err.Error())
			continue
		}

		books = append(books, book)
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	err = encoder.Encode(books)
	if err != nil {
		fmt.Fprintf(w, "{\"error\": \"Internal JSON Error\"}")
		fmt.Println("Failed to encode chapter list: " + err.Error())
	}
}

func getChapters(w http.ResponseWriter, r *http.Request) {
	bookId, err := strconv.Atoi(r.URL.Query().Get("book"))
	if err != nil {
		fmt.Fprintf(w, "{\"error\": \"Invalid Book ID\"}")
		return
	}

	res, err := db.Query("SELECT chapter_no, chapter_name FROM chapters WHERE book_id = ?", bookId)
	if err != nil {
		fmt.Fprintf(w, "{\"error\": \"Internal MySQL Error\"}")
		fmt.Println("MySQL Error when fetching book list: " + err.Error())
		return
	}

	var chapters []Chapter

	for res.Next() {
		var chapter Chapter
		err = res.Scan(&chapter.Number, &chapter.Name)
		if err != nil {
			fmt.Fprintf(w, "{\"error\": \"Internal MySQL Error\"}")
			fmt.Println("MySQL Error when parsing chapter for contents: " + err.Error())
			continue
		}

		chapters = append(chapters, chapter)
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	err = encoder.Encode(chapters)
	if err != nil {
		fmt.Fprintf(w, "{\"error\": \"Internal JSON Error\"}")
		fmt.Println("Failed to encode chapter list: " + err.Error())
	}
}

func getChapter(w http.ResponseWriter, r *http.Request) {
	bookId, err := strconv.Atoi(r.URL.Query().Get("book"))
	if err != nil {
		fmt.Fprintf(w, "{\"error\": \"Invalid Book ID\"}")
		return
	}

	chapterNo, err := strconv.Atoi(r.URL.Query().Get("chapter"))
	if err != nil {
		fmt.Fprintf(w, "{\"error\": \"Invalid Chapter Number\"}")
		return
	}

	chapter := Chapter{Number: uint(chapterNo)}
	err = db.QueryRow("SELECT chapter_name, chapter_contents FROM chapters WHERE book_id = ? AND chapter_no = ?", bookId, chapterNo).Scan(&chapter.Name, &chapter.Contents)
	if err == sql.ErrNoRows {
		fmt.Fprintf(w, "{\"error\": \"Chapter Not Found\"}")
		return
	} else if err != nil {
		fmt.Fprintf(w, "{\"error\": \"Internal MySQL Error\"}")
		fmt.Println("MySQL Error when fetching chapter contents: " + err.Error())
		return
	}

	fmt.Fprintf(
		w, `{"number": %s, "name": "%s", "contents": %s}`,
		strconv.FormatUint(uint64(chapter.Number), 10), chapter.Name, chapter.Contents,
	)
}

func homePage(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "Welcome to the homepage!")
	fmt.Println("Endpoint: homePage")
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", homePage)
	router.HandleFunc("/chapter", getChapter)
	router.HandleFunc("/books", getBooks)
	router.HandleFunc("/chapters", getChapters)

	log.Fatal(http.ListenAndServe(":10000", router))
}

func main() {
	var err error
	db, err = sql.Open("mysql", "booklocker:password@tcp(localhost:3306)/booklocker")
	if err != nil {
		panic(err.Error)
	}

	defer db.Close()

	handleRequests()
}
