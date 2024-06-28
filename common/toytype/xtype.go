package toytype

import "fmt"

type MyStrSlice []string

func (s MyStrSlice) String() string {
	var newS []string
	copy(newS, s)
	return fmt.Sprintf("%v", newS)
}

func (s MyStrSlice) Set(value string) error {
	s = append(s, value)
	return nil
}
