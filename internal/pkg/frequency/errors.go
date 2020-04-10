package frequency

import "fmt"

type errInvalidExpr string

func (expr errInvalidExpr) Error() string {
	return fmt.Sprintf("invalid frequency expression '%s'", string(expr))
}

type errWrongOrder string

func (expr errWrongOrder) Error() string {
	return fmt.Sprintf("Wrong order of elements in frequency expression %s, should be in descending order", string(expr))
}
