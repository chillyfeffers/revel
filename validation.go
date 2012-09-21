package rev

import (
	"fmt"
	"regexp"
	"time"
)

type ValidationError struct {
	Message, Key string
}

// Returns the Message.
func (e *ValidationError) String() string {
	if e == nil {
		return ""
	}
	return e.Message
}

// A Validation context manages data validation and error messages.
type Validation struct {
	Errors []*ValidationError
	keep   bool
}

func (v *Validation) Keep() {
	v.keep = true
}

func (v *Validation) Clear() {
	v.Errors = []*ValidationError{}
}

func (v *Validation) HasErrors() bool {
	return len(v.Errors) > 0
}

// Return the errors mapped by key.
// If there are multiple validation errors associated with a single key, the
// first one "wins".  (Typically the first validation will be the more basic).
func (v *Validation) ErrorMap() map[string]*ValidationError {
	m := map[string]*ValidationError{}
	for _, e := range v.Errors {
		if _, ok := m[e.Key]; !ok {
			m[e.Key] = e
		}
	}
	return m
}

// A ValidationResult is returned from every validation method.
// It provides an indication of success, and a pointer to the Error (if any).
type ValidationResult struct {
	Error *ValidationError
	Ok    bool
}

func (r *ValidationResult) Key(key string) *ValidationResult {
	if r.Error != nil {
		r.Error.Key = key
	}
	return r
}

func (r *ValidationResult) Message(message string) *ValidationResult {
	if r.Error != nil {
		r.Error.Message = message
	}
	return r
}

type Check interface {
	IsSatisfied(interface{}) bool
	DefaultMessage() string
}

/*
	Required validator. Use to ensure that a parameter is present in the request parameters and
	is not empty. Empty strings, slices, and zero dates are considered empty.
*/
type Required struct{}

func (r Required) IsSatisfied(obj interface{}) bool {
	if obj == nil {
		return false
	}

	if str, ok := obj.(string); ok {
		return len(str) > 0
	}
	if list, ok := obj.([]interface{}); ok {
		return len(list) > 0
	}
	if t, ok := obj.(time.Time); ok {
		return !t.IsZero()
	}
	return true
}

func (r Required) DefaultMessage() string {
	return "Required"
}

func (v *Validation) Required(obj interface{}) *ValidationResult {
	return v.check(Required{}, obj)
}

/*
	Min validator. Use to ensure that a parameter is an integer not less than a certain number.
*/
type Min struct {
	Min int
}

func (m Min) IsSatisfied(obj interface{}) bool {
	num, ok := obj.(int)
	if ok {
		return num >= m.Min
	}
	return false
}

func (m Min) DefaultMessage() string {
	return fmt.Sprintln("Minimum is", m.Min)
}

func (v *Validation) Min(n int, min int) *ValidationResult {
	return v.check(Min{min}, n)
}

/*
	Max validator. Use to ensure that a parameter is an integer not greater than a certain number.
*/
type Max struct {
	Max int
}

func (m Max) IsSatisfied(obj interface{}) bool {
	num, ok := obj.(int)
	if ok {
		return num <= m.Max
	}
	return false
}

func (m Max) DefaultMessage() string {
	return fmt.Sprintln("Maximum is", m.Max)
}

func (v *Validation) Max(n int, max int) *ValidationResult {
	return v.check(Max{max}, n)
}

/*
	Range validator. Use to ensure that a parameter is an int within an inclusive integer interval.
*/
type Range struct {
	Min int
	Max int
}

func (r Range) IsSatisfied(obj interface{}) bool {
	num, ok := obj.(int)
	if ok {
		return r.Min <= num && num <= r.Max
	}
}

func (r Range) DefaultMessage() string {
	return fmt.Sprintf("Valid range is %d to %d, inclusive.", r.Min, r.Max)
}

func (v *Validation) Range(n int, min, max int) *ValidationResult {
	return v.check(Range{min, max}, n)
}

// Requires an array or string to be at least a given length.
type MinSize struct {
	Min int
}

func (m MinSize) IsSatisfied(obj interface{}) bool {
	if arr, ok := obj.([]interface{}); ok {
		return len(arr) >= m.Min
	}
	if str, ok := obj.(string); ok {
		return len(str) >= m.Min
	}
	return false
}

func (m MinSize) DefaultMessage() string {
	return fmt.Sprintln("Minimum size is", m.Min)
}

func (v *Validation) MinSize(obj interface{}, min int) *ValidationResult {
	return v.check(MinSize{min}, obj)
}

// Requires an array or string to be at most a given length.
type MaxSize struct {
	Max int
}

func (m MaxSize) IsSatisfied(obj interface{}) bool {
	if arr, ok := obj.([]interface{}); ok {
		return len(arr) <= m.Max
	}
	if str, ok := obj.(string); ok {
		return len(str) <= m.Max
	}
	return false
}

func (m MaxSize) DefaultMessage() string {
	return fmt.Sprintln("Maximum size is", m.Max)
}

func (v *Validation) MaxSize(obj interface{}, max int) *ValidationResult {
	return v.check(MaxSize{max}, obj)
}

// Requires a string to match a given regex.
type Match struct {
	Regexp *regexp.Regexp
}

func (m Match) IsSatisfied(obj interface{}) bool {
	str := obj.(string)
	return m.Regexp.MatchString(str)
}

func (m Match) DefaultMessage() string {
	return fmt.Sprintln("Must match", m.Regexp)
}

func (v *Validation) Match(str string, regex *regexp.Regexp) *ValidationResult {
	return v.check(Match{regex}, str)
}

func (v *Validation) check(chk Check, obj interface{}) *ValidationResult {
	if chk.IsSatisfied(obj) {
		return &ValidationResult{Ok: true}
	}

	// Add the error to the validation context.
	err := &ValidationError{
		Message: chk.DefaultMessage(),
	}
	v.Errors = append(v.Errors, err)

	// Also return it in the result.
	return &ValidationResult{
		Ok:    false,
		Error: err,
	}
}

// Apply a group of Checks to a field, in order, and return the ValidationResult
// from the first Check that fails, or the last one that succeeds.
func (v *Validation) Check(obj interface{}, checks ...Check) *ValidationResult {
	var result *ValidationResult
	for _, check := range checks {
		result = v.check(check, obj)
		if !result.Ok {
			return result
		}
	}
	return result
}
