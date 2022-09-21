/*
	# Problem
	Today you have been asked to create an API using the framework Gin (https://pkg.go.dev/github.com/gin-gonic/gin).
	This project will test your skills in creating functional backends using Golang,
	will also test your ability to make development code and test your understanding of Golang.

	# Requirements
	- Create an API to catalog books, using the models of `shelf`, `book` , `author`.
	- Shelf contains books in a certain books starting with specific letters section, e.g. A-G, H-J, K-T, T-Z.
	- Book model has an author, pages, ISBN and title.

	# Task
	- Be able to create, read, update and delete books from shelfs. You should be able to see all shelfs, and query a specific shelfs' books contained in it.
*/

package main

import (
	"errors"
	"fmt"
	"github.com/Mikadore/interview-books/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
)

type State struct {
	mu      sync.RWMutex
	shelves []model.Shelf
}

func (self *State) FindById(id string) (*model.Shelf, error) {
	if len(id) != 3 {
		return nil, errors.New("Invalid id (length != 3)")
	}

	// Assuming ASCII, which is ok here, as
	// non-ASCII will just fail
	start := rune(id[0])
	sep := id[1]
	end := rune(id[2])

	if sep != '-' {
		return nil, errors.New("Invalid id (separator != '-')")
	}

	self.mu.RLock()
	defer self.mu.RUnlock()
	for i := range self.shelves {
		// Due to weird pointer behavior in range
		shelf := &self.shelves[i]
		if shelf.StartLetter == start && shelf.EndLetter == end {
			return shelf, nil
		}
	}

	return nil, errors.New("Shelf not found")
}

func main() {

	state := State{shelves: []model.Shelf{
		{
			StartLetter: 'A',
			EndLetter:   'G',
			Books:       make([]model.Book, 0),
		},
		{
			StartLetter: 'H',
			EndLetter:   'J',
			Books:       make([]model.Book, 0),
		},
		{
			StartLetter: 'K',
			EndLetter:   'T',
			Books:       make([]model.Book, 0),
		},
		{
			StartLetter: 'U',
			EndLetter:   'Z',
			Books:       make([]model.Book, 0),
		},
	}}
	server := gin.Default()

	// Returns a list of shelf ids
	server.GET("/shelves", func(c *gin.Context) {
		// Constrain scope so we don't hold the mutex longer than needed
		ids := func() []string {
			state.mu.RLock()
			defer state.mu.RUnlock()

			ids := make([]string, len(state.shelves))

			for i := range state.shelves {
				ids[i] = fmt.Sprintf("%c-%c", state.shelves[i].StartLetter, state.shelves[i].EndLetter)
			}

			return ids
		}()

		c.JSON(http.StatusOK, ids)
	})

	// GET shelf by id, returns list of Books in shelf
	server.GET("/shelves/:id", func(c *gin.Context) {
		id := c.Param("id")

		if shelf, err := state.FindById(id); err == nil {
			// Need to re-acquire the mutex before reading shelf
			state.mu.RLock()
			defer state.mu.RUnlock()

			c.JSON(http.StatusOK, &shelf.Books)
		} else {
			c.String(http.StatusNotFound, err.Error())
		}
	})

	// Get book in shelf {id} by it's isbn
	server.GET("/shelves/:id/:isbn", func(c *gin.Context) {
		id := c.Param("id")
		isbn := c.Param("isbn")

		if shelf, err := state.FindById(id); err == nil {
			state.mu.RLock()
			defer state.mu.RUnlock()

			for i := range shelf.Books {
				if shelf.Books[i].ISBN == isbn {
					c.JSON(http.StatusOK, shelf.Books[i])
					return
				}
			}
			c.String(http.StatusNotFound, fmt.Sprintf("Couldn't find isbn '%s'", isbn))
		} else {
			c.String(http.StatusNotFound, fmt.Sprintf("Shelf not found: %s", err.Error()))
		}
	})

	// Create a book in shelf {id} using {isbn} as the key
	server.POST("/shelves/:id/:isbn", func(c *gin.Context) {
		id := c.Param("id")
		isbn := c.Param("isbn")

		if shelf, err := state.FindById(id); err == nil {
			var book model.Book

			if err := c.BindJSON(&book); err == nil {
				if book.ISBN != isbn {
					c.String(http.StatusBadRequest, "ISBN doesn't equal path param")
				}

				state.mu.Lock()
				defer state.mu.Unlock()

				shelf.Books = append(shelf.Books, book)

				c.String(http.StatusCreated, "Success!")
			} else {
				c.String(http.StatusBadRequest, err.Error())
			}
		} else {
			c.String(http.StatusNotFound, err.Error())
		}
	})

	// Update an existing book in a shelf by it's ISBN
	server.PATCH("/shelves/:id/:isbn", func(c *gin.Context) {
		id := c.Param("id")
		isbn := c.Param("isbn")

		if shelf, err := state.FindById(id); err == nil {
			var book model.Book
			if err := c.BindJSON(&book); err == nil {
				state.mu.Lock()
				defer state.mu.Unlock()

				for i := range shelf.Books {
					if shelf.Books[i].ISBN == isbn {
						shelf.Books[i] = book
						c.String(http.StatusOK, "Success!")
						return
					}
				}
				c.String(http.StatusNotFound, fmt.Sprintf("Book not found: %s", isbn))
			} else {
				c.String(http.StatusBadRequest, err.Error())
			}
		} else {
			c.String(http.StatusNotFound, fmt.Sprintf("Shelf not found: %s", err.Error()))
		}
	})

	server.DELETE("/shelves/:id/:isbn", func(c *gin.Context) {
		id := c.Param("id")
		isbn := c.Param("isbn")

		if shelf, err := state.FindById(id); err == nil {
			state.mu.Lock()
			defer state.mu.Unlock()
			for i := range shelf.Books {
				if shelf.Books[i].ISBN == isbn {
					shelf.Books[i] = shelf.Books[len(shelf.Books)-1]
    				shelf.Books = shelf.Books[:len(shelf.Books)-1]

					c.String(http.StatusOK, "Deleted!")
					return
				}
			}
		} else {
			c.String(http.StatusNotFound, fmt.Sprintf("Shelf not found: %s", err.Error()))
		}
	})

	server.Run()
}
