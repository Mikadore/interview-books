package model

type Author = string

type Book struct {
	Author    Author `json:"author"`
	ISBN      string `json:"isbn"`
	Title     string `json:"title"`
	PageCount uint   `json:"pages"`
}

type Shelf struct {
	StartLetter rune
	EndLetter   rune
	Books       []Book
}

func (self *Shelf) push(book Book) {
	self.Books = append(self.Books, book)
}
