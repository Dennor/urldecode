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
	StringValue   string                   `url:"string_value"`
	UnsignedValue uint                     `url:"unsigned_value"`
	JSONValue     encoding.TextUnmarshaler `url:"json_value"`
}

type benchMockUnmarshalText struct{}

func (b *benchMockUnmarshalText) UnmarshalText(data []byte) error {
	// Do nothing
	return nil
}

var (
	jsonValueMock = new(mockUnmarshalText)
	urlBytes      = []byte("bool_value%3Dtrue%26int_value%3D5%26float_value%3D1.234%26string_value%3DBj%C3%B6rk%20Gu%C3%B0mundsd%C3%B3ttir%26unsigned_value%3D5%26json_value%3D%7B%22field1%22%3A%201%2C%20%22field2%22%3A%202%7D")
	urlDecoded    = TestType{
		BoolValue:     true,
		IntValue:      5,
		FloatValue:    1.234,
		StringValue:   "Björk Guðmundsdóttir",
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
