package frequency

import "fmt"

type errInvalidExpr string

func (expr errInvalidExpr) Error() string {
	return fmt.Sprintf("invalid frequency expression '%s'", string(expr))
}
