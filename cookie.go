package zhttp

type Cookie Value

// RawCookie used to create Cookie struct with string
// like key1=value1;key2=value2
// will append to request cookie
func RawCookie(cookie string) Cookie {
	return Cookie{Raw: cookie}
}

// PairsCookie used to create Cookie struct with map
// will append to request cookie
func PairsCookie(cookie map[string]string) Cookie {
	return Cookie{
		Pairs: cookie,
	}
}
