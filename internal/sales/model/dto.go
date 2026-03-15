package model

type Pagination struct {
	Page       int
	Limit      int
	Total      int64
	Start      int
	End        int
	PrevPage   int
	NextPage   int
	PageSize   int
	TotalPages int
	HasPrev    bool
	HasNext    bool
	Pages      []int
	SearchTerm string
	FileType   string
}

// NewPagination creates a new pagination with calculated fields
func NewPagination(page, limit int) *Pagination {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	start := (page-1)*limit + 1
	end := page * limit

	prevPage := page - 1
	if prevPage < 1 {
		prevPage = 1
	}

	nextPage := page + 1

	return &Pagination{
		Page:       page,
		Limit:      limit,
		PageSize:   limit,
		Start:      start,
		End:        end,
		PrevPage:   prevPage,
		NextPage:   nextPage,
		Total:      0,
		TotalPages: 0,
	}
}

// SetTotal updates the pagination with total count and recalculates fields
func (p *Pagination) SetTotal(total int64) {
	p.Total = total
	p.TotalPages = int((total + int64(p.Limit) - 1) / int64(p.Limit))

	if p.End > int(total) {
		p.End = int(total)
	}

	if p.NextPage > p.TotalPages {
		p.NextPage = p.TotalPages
	}

	p.HasPrev = p.Page > 1
	p.HasNext = p.Page < p.TotalPages

	p.Pages = p.generatePageNumbers()
}

func (p *Pagination) generatePageNumbers() []int {
	var pages []int

	if p.TotalPages <= 7 {
		for i := 1; i <= p.TotalPages; i++ {
			pages = append(pages, i)
		}
	} else {
		if p.Page <= 4 {
			for i := 1; i <= 5; i++ {
				pages = append(pages, i)
			}
			pages = append(pages, -1)
			pages = append(pages, p.TotalPages)
		} else if p.Page >= p.TotalPages-3 {
			pages = append(pages, 1)
			pages = append(pages, -1)
			for i := p.TotalPages - 4; i <= p.TotalPages; i++ {
				pages = append(pages, i)
			}
		} else {
			pages = append(pages, 1)
			pages = append(pages, -1)
			for i := p.Page - 1; i <= p.Page+1; i++ {
				pages = append(pages, i)
			}
			pages = append(pages, -1)
			pages = append(pages, p.TotalPages)
		}
	}

	return pages
}

type ClientFilter struct {
	Term       string
	EbookID    uint
	Pagination *Pagination
}

type ClientRequest struct {
	ID        uint   `json:"id"`
	Name      string `validate:"required,min=5,max=120" json:"name"`
	CPF       string `validate:"required,max=120" json:"cpf"`
	Birthdate string `validate:"required"`
	Email     string `validate:"required,email" json:"email"`
	Phone     string `validate:"max=14" json:"phone"`
}

type UpdateClientInput struct {
	ID           uint
	Email        string
	Phone        string
	EmailCreator string
}
