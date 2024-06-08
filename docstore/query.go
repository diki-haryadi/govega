package docstore

// FilterOpt filter option
type FilterOpt struct {
	Field string
	Value interface{}
	Ops   string
}

// QueryOpt query option
type QueryOpt struct {
	Limit    int
	Skip     int
	Page     int
	OrderBy  string
	IsAscend bool
	Filter   []FilterOpt
}

func (q *QueryOpt) AddFilter(filter FilterOpt) *QueryOpt {
	for _, f := range q.Filter {
		if f.Field == filter.Field {
			return q
		}
	}

	q.Filter = append(q.Filter, filter)
	return q
}
