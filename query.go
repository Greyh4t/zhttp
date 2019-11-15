package zhttp

type Value struct {
	Raw   string
	Pairs map[string]string
}

type Query Value

// RawQuery used to create Query struct with string
// like key=value1&key2=value2
// will overwrite other query part of url
func RawQuery(query string) Query {
	return Query{Raw: query}
}

// PairsQuery used to create Query struct with map
// will overwritten duplicate key of url
func PairsQuery(query map[string]string) Query {
	return Query{
		Pairs: query,
	}
}
