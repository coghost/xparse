package py3

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type Py3Suite struct {
	suite.Suite
}

func TestPy3(t *testing.T) {
	suite.Run(t, new(Py3Suite))
}

func (s *Py3Suite) SetupSuite() {
}

func (s *Py3Suite) TearDownSuite() {
}

func (s *Py3Suite) TestPy3() {
	raw := `
import sys
arr = sys.argv[1].split(",")
print(arr[0])`

	res, err := Eval(raw, "python_is_great, do you know?")
	s.Nil(err)

	s.Equal("python_is_great", res.RefinedString)
}
