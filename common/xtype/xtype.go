package xtype

import "fmt"

type StringSliceX []string

func (s *StringSliceX) String() string {
	return fmt.Sprintf("%v", []string(*s))
}

func (s *StringSliceX) Set(value string) error {
	*s = append(*s, value)
	return nil
}
