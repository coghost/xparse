package js

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type JSSuite struct {
	suite.Suite
}

func TestJS(t *testing.T) {
	suite.Run(t, new(JSSuite))
}

func (s *JSSuite) SetupSuite() {
}

func (s *JSSuite) TearDownSuite() {
}

func (s *JSSuite) TestSimple() {
	code := `
arr = input.split(",")
output = arr[0]
	`

	res, err := Eval(code, "js_is_great, do you know?")
	s.Nil(err)

	s.Equal("js_is_great", res.RefinedString)
}

func (s *JSSuite) TestUnderscore() {
	code := `
arr = input.split(",")
arr = _.uniq(arr)
output = arr.join()
	`

	res, err := Eval(code, "apple,banana,apple,kiwi,lemon,kiwi")
	s.Nil(err)

	s.Equal("apple,banana,kiwi,lemon", res.RefinedString)
}
