package urldecode

import (
	"encoding"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUnmarshalText struct {
	mock.Mock
}

func (m *mockUnmarshalText) UnmarshalText(data []byte) error {
	return m.Called(data).Error(0)
}

type TestType struct {
	BoolValue     bool                     `url:"bool_value"`
	IntValue      int                      `url:"int_value"`
	FloatValue    float64                  `url:"float_value"`
	UnsignedValue uint                     `url:"unsigned_value"`
	StringValue   string                   `url:"string_value"`
	JSONValue     encoding.TextUnmarshaler `url:"json_value"`
}

type benchMockUnmarshalText struct{}

func (b *benchMockUnmarshalText) UnmarshalText(data []byte) error {
	// Do nothing
	return nil
}

var (
	jsonValueMock = new(mockUnmarshalText)
	urlBytes      = []byte("bool_value=true&int_value=5&float_value=1.234&unsigned_value=5&json_value=%7B%22field1%22%3A%201%2C%20%22field2%22%3A%202%7D&string_value=http%3A%2F%2Fexample.com%2Fpath%3Fa%3Da%26b%3Db%26name%3DBj%C3%B6rk%20Gu%C3%B0mundsd%C3%B3ttir")
	urlDecoded    = TestType{
		BoolValue:     true,
		IntValue:      5,
		FloatValue:    1.234,
		StringValue:   "http://example.com/path?a=a&b=b&name=Björk Guðmundsdóttir",
		UnsignedValue: 5,
		JSONValue:     jsonValueMock,
	}
)

func TestUrlUnmarshal(t *testing.T) {
	assert := assert.New(t)
	jsonValueMock.On("UnmarshalText", []byte(`{"field1": 1, "field2": 2}`)).Return(nil)
	actual := TestType{JSONValue: jsonValueMock}
	assert.NoError(Unmarshal(urlBytes, &actual))
	assert.Equal(urlDecoded, actual)
}

func BenchmarkUrlUnmarshal(b *testing.B) {
	b.ReportAllocs()
	actual := TestType{JSONValue: new(benchMockUnmarshalText)}
	for n := 0; n < b.N; n++ {
		Unmarshal(urlBytes, &actual)
	}
}
