package handlers

import (
	"net/http"

	"github.com/thomastaylor312/printing-api/store"
	"github.com/thomastaylor312/printing-api/types"
)

type PaperHandlers struct {
	db store.DataStore
}

func NewPaperHandlers(db store.DataStore) *PaperHandlers {
	return &PaperHandlers{db: db}
}

// GetPapers gets all papers from the database
func (p *PaperHandlers) GetPapers(w http.ResponseWriter, r *http.Request) {
	get[*types.PaperType](p.db, "papers", w, r)
}

// AddPaper adds a paper to the database
func (p *PaperHandlers) AddPaper(w http.ResponseWriter, r *http.Request) {
	add[*types.PaperType](p.db, "papers", w, r, nil, nil)
}

// UpdatePaper updates a paper in the database
func (p *PaperHandlers) UpdatePaper(w http.ResponseWriter, r *http.Request) {
	update[*types.PaperType](p.db, "papers", w, r, nil, nil)
}

// DeletePaper deletes a paper from the database
func (p *PaperHandlers) DeletePaper(w http.ResponseWriter, r *http.Request) {
	delete[*types.PaperType](p.db, "papers", w, r, nil)
}
