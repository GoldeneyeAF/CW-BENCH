package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// struct based on books.json file. Please refer
type Book struct {
	Id       string `json:"id"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	Price    string `json:"price"`
	Imageurl string `json:"image_url"`
}

// define port
const PORT string = ":8080"

// message to send as json response
type Message struct {
	Msg string
}

// response as json format
func jsonMessageByte(msg string) []byte {
	errrMessage := Message{msg}
	byteContent, _ := json.Marshal(errrMessage)
	return byteContent
}

// print logs in console
func checkError(err error) {
	if err != nil {
		log.Printf("Error - %v", err)
	}

}

// List all the books handler
func handleGetBooks(w http.ResponseWriter, r *http.Request) {
	books, err := getBooks()

	// send server error as response
	if err != nil {
		log.Printf("Server Error %v\n", err)
		w.WriteHeader(500)
		w.Write(jsonMessageByte("Internal server error"))
	} else {
		booksByte, _ := json.Marshal(books)
		w.Write(booksByte)
	}

}

// get book by id handler
func handleGetBookById(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()
	// get book id from URL
	bookId := query.Get("id")
	book, _, err := getBookById(bookId)
	// send server error as response
	if err != nil {
		log.Printf("Server Error %v\n", err)
		w.WriteHeader(500)
		w.Write(jsonMessageByte("Internal server error"))
	} else {
		// check requested book exists or not
		if (Book{}) == book {
			w.Write(jsonMessageByte("Book Not found"))
		} else {
			bookByte, _ := json.Marshal(book)
			w.Write(bookByte)
		}
	}
}

// add book handler
func handleAddBook(w http.ResponseWriter, r *http.Request) {
	// check for post method
	if r.Method != "POST" {
		w.WriteHeader(405)
		w.Write(jsonMessageByte(r.Method + " - Method not allowed"))
	} else {
		// read the body
		newBookByte, err := ioutil.ReadAll(r.Body)
		// check for valid data from client
		if err != nil {
			log.Printf("Client Error %v\n", err)
			w.WriteHeader(400)
			w.Write(jsonMessageByte("Bad Request"))
		} else {
			books, _ := getBooks() // get all books
			var newBooks []Book    // to add new book

			json.Unmarshal(newBookByte, &newBooks) // new book added
			books = append(books, newBooks...)     // add both
			// Write all the books in books.json file
			err = saveBooks(books)
			// send server error as response
			if err != nil {
				log.Printf("Server Error %v\n", err)
				w.WriteHeader(500)
				w.Write(jsonMessageByte("Internal server error"))
			} else {
				w.Write(jsonMessageByte("New book added successfully"))
			}

		}
	}
}

// update book handler
func handleUpdateBook(w http.ResponseWriter, r *http.Request) {
	// check for post method
	if r.Method != "POST" {
		w.WriteHeader(405)
		w.Write(jsonMessageByte(r.Method + " - Method not allowed"))
	} else {
		// read the body
		updateBookByte, err := ioutil.ReadAll(r.Body)
		// check for valid data from client
		if err != nil {
			log.Printf("Client Error %v\n", err)
			w.WriteHeader(400)
			w.Write(jsonMessageByte("Bad Request"))
		} else {
			var updateBook Book // to update a book

			err = json.Unmarshal(updateBookByte, &updateBook) // new book added
			checkError(err)
			id := updateBook.Id

			book, _, _ := getBookById(id)
			// check requested book exists or not
			if (Book{}) == book {
				w.Write(jsonMessageByte("Book Not found"))
			} else {
				books, _ := getBooks()

				for i, book := range books {
					if book.Id == updateBook.Id {
						books[i] = updateBook
					}
				}
				// write books in books.json
				err = saveBooks(books)
				// send server error as response
				if err != nil {
					log.Printf("Server Error %v\n", err)
					w.WriteHeader(500)
					w.Write(jsonMessageByte("Internal server error"))
				} else {
					w.Write(jsonMessageByte("Book updated successfully"))
				}
			}
		}
	}
}

// delete book by id handler
func handleDeleteBookById(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()
	// get book id from URL
	bookId := query.Get("id")
	book, book_index, err := getBookById(bookId)
	// send server error as response
	if err != nil {
		log.Printf("Server Error %v\n", err)
		w.WriteHeader(500)
		w.Write(jsonMessageByte("Internal server error"))
	} else {
		// check requested book exists or not
		if (Book{}) == book {
			w.Write(jsonMessageByte("Book Not found"))
		} else {
			books, _ := getBooks()
			// remove books from slice
			books = append(books[:book_index], books[book_index+1:]...)
			saveBooks(books)
			w.Write(jsonMessageByte("Book deleted successfully"))
		}
	}
}

// Get books - returns books and error
func getBooks() ([]Book, error) {
	books := []Book{}
	booksByte, err := ioutil.ReadFile("./books.json")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(booksByte, &books)
	if err != nil {
		return nil, err
	}
	return books, nil
}

// Get books - returns book, book index and error
func getBookById(id string) (Book, int, error) {
	books, err := getBooks()
	var requestedBook Book
	var requestedBookIndex int

	if err != nil {
		return Book{}, 0, err
	}

	for i, book := range books {
		if book.Id == id {
			requestedBook = book
			requestedBookIndex = i
		}
	}

	return requestedBook, requestedBookIndex, nil
}

// save books to books.json file
func saveBooks(books []Book) error {

	// converting into bytes for writing into a file
	booksBytes, err := json.Marshal(books)

	checkError(err)

	err = ioutil.WriteFile("./books.json", booksBytes, 0644)

	return err
}
