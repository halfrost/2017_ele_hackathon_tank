package cc

import "time"

// Configer is a abstraction for config.
type Configer interface {
	KV() map[string]interface{}
	Has(name string) bool
	Must(name string)

	Raw(name string) interface{}
	Value(name string) Valuer
	Config(name string) Configer
	Pattern(name string) Patterner

	SetDefault(name string, value interface{})
	Set(name string, value interface{})

	String(name string) string
	StringOr(name string, deflt string) string
	StringAnd(name string, pattern string) (string, bool)
	StringAndOr(name string, pattern string, deflt string) string

	Bool(name string) bool
	BoolOr(name string, deflt bool) bool

	Int(name string) int
	IntOr(name string, deflt int) int
	IntAnd(name string, pattern string) (int, bool)
	IntAndOr(name string, pattern string, deflt int) int

	Int64(name string) int64
	Int64Or(name string, deflt int64) int64
	Int64And(name string, pattern string) (int64, bool)
	Int64AndOr(name string, pattern string, deflt int64) int64

	Float(name string) float64
	FloatOr(name string, deflt float64) float64
	FloatAnd(name string, pattern string) (float64, bool)
	FloatAndOr(name string, pattern string, deflt float64) float64

	Duration(name string) time.Duration
	DurationOr(name string, deflt int64) time.Duration
	DurationAnd(name string, pattern string) (time.Duration, bool)
	DurationAndOr(name string, pattern string, deflt int64) time.Duration
}

// Valuer is a abstraction for config value, which can convert into multiple types.
type Valuer interface {
	Exist() bool

	Raw() interface{}
	Config() Configer
	Pattern() Patterner
	Map() map[string]Valuer
	List() []Valuer

	String() string
	StringOr(deflt string) string
	StringAnd(pattern string) (string, bool)
	StringAndOr(pattern string, deflt string) string

	Bool() bool
	BoolOr(deflt bool) bool

	Int() int
	IntOr(deflt int) int
	IntAnd(pattern string) (int, bool)
	IntAndOr(pattern string, deflt int) int

	Int64() int64
	Int64Or(deflt int64) int64
	Int64And(pattern string) (int64, bool)
	Int64AndOr(pattern string, deflt int64) int64

	Float() float64
	FloatOr(deflt float64) float64
	FloatAnd(pattern string) (float64, bool)
	FloatAndOr(pattern string, deflt float64) float64

	Duration() time.Duration
	DurationOr(deflt int64) time.Duration
	DurationAnd(pattern string) (time.Duration, bool)
	DurationAndOr(pattern string, deflt int64) time.Duration
}

// Patterner is abstraction which do validation work.
// if pattern is not valid, then the following methods will always return false.
//
// string pattern use the native regular expression to validate the value.
//
// int(time.Duration) and float64 pattern use the basic if-like conditions to calculate and validate
// the value, use 'N' as placeholder for number, bit operation is not supported,
// for example:
//     "N>2"
//     "N>1&&N<=5"
//     "N<1||N>3"
//     "(N%2==0)&&(N<=4||N>=8)"
type Patterner interface {
	Err() error
	ValidateInt(n int) bool
	ValidateFloat(n float64) bool
	ValidateString(s string) bool
}
