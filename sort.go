package gohm

import (
	"strings"
)

type sort struct {
	Query     query
	ByArg     string
	StoreArg  string
	CountArg  int64
	OffsetArg int64
	OrdersArg []string
}

func (q query) Sort() sort {
	return sort{
		Query: q,
	}
}

func (s sort) By(by string) sort {
	s.ByArg = by
	return s
}

func (s sort) Store(store string) sort {
	s.StoreArg = store
	return s
}

func (s sort) Limit(offset int64, count int64) sort {
	s.OffsetArg = offset
	s.CountArg = count
	return s
}

func (s sort) Order(order string) sort {
	s.OrdersArg = strings.Split(order, " ")
	return s
}

func (s sort) Fetch(v interface{}) error {
	set, err := s.Query.Model(v).Set()
	if err != nil {
		return err
	}
	args := make([]interface{}, 0)
	if len(s.ByArg) > 0 {
		args = append(args, "BY", s.ByArg)
	}
	if len(s.StoreArg) > 0 {
		args = append(args, "STORE", s.StoreArg)
	}
	if s.CountArg > 0 || s.OffsetArg > 0 {
		args = append(args, "LIMIT", s.OffsetArg, s.CountArg)
	}
	for i := range s.OrdersArg {
		args = append(args, s.OrdersArg[i])
	}
	return SetSort(set, v, args)
}
