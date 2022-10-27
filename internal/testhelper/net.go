package testhelper

func NoopLookupHost(host string) ([]string, error) {
	return []string{host}, nil
}
