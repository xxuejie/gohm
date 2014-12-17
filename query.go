package gohm

type query struct {
	G       *Gohm
	Queries map[string]interface{}
}

func (g *Gohm) All() query {
	return query{
		G:       g,
		Queries: make(map[string]interface{}),
	}
}

// func (q query) Find(k string, v interface{}) query {
// 	q.Queries[k] = v
// 	return q
// }

func (q query) Fetch(v interface{}) error {
	modelName := fetchTypeNameFromReturnInterface(v)
	// TODO: filter case
	return NewSet(q.G, connectKeys(modelName, "all"), modelName).Fetch(v)
}
