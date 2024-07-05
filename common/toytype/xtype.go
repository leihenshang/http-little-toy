package toytype

import "fmt"

type MyStrSlice []string

func (s *MyStrSlice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *MyStrSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}
