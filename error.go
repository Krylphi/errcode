package errcode

import (
	"bytes"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"
	"sync/atomic"
)

type (
	// ErrSeed is a builder for GeneralError types
	ErrSeed interface {
		SubType(errSubType string) ErrSeed
		ExternalErrMess(err error) ErrSeed
		Message(message string) ErrSeed
		MessageF(format string, a ...interface{}) ErrSeed
		Make() GeneralError
	}

	// GeneralError is custom error type builder designed for convenient error handling using
	// native errors 1.13 implementation
	GeneralError interface {
		Error() string
		FmtResponse(template string) string
		ErrorCode() string
		CodeNote() string
		SysMessage() string
		Unwrap() error
		Is(err error) bool
		Produce() ErrSeed
	}

	// generalError is GeneralError implementation
	generalError struct {
		codeGen  func(string) string
		sumCodes func(old, new string) string
		err      error
		cause    error
		export   atomic.Value
		code     string
		codeNote string
		mes      string
	}
)

// NewGeneralError is GeneralError constructor with default codes generators
func NewGeneralError(code string) ErrSeed {
	return NewGeneralErrorWithCustomCodes(code, errCodeGen, sumCodes)
}

// NewGeneralErrorWithCustomCodes is GeneralError constructor with custom code generation
func NewGeneralErrorWithCustomCodes(code string, codeGen func(string string) string, inheritCodeGen func(old, new string) string) ErrSeed {
	c := codeGen(code)
	n := code
	e := errors.New(c)
	r := &generalError{
		codeGen:  codeGen,
		sumCodes: inheritCodeGen,
		err:      e,
		cause:    e,
		code:     c,
		codeNote: concatenate(c, ":", n),
	}
	r.export.Store("")
	return r
}

// ToGeneralError try to convert err to GeneralError
func ToGeneralError(err error) (GeneralError, bool) { e, ok := err.(GeneralError); return e, ok }

// IsGeneralError checks whether err is GeneralError type
func IsGeneralError(err error) bool { _, ok := ToGeneralError(err); return ok }

// Error returns string for error
func (r generalError) Error() string {
	export := r.export.Load()
	emptyExp := func() bool {
		switch {
		case export == nil:
			return true
		case strings.Contains(export.(string), r.SysMessage()):
			return false
		case strings.Contains(export.(string), r.ErrorCode()):
			return false
		default:
			return true
		}
	}

	if emptyExp() {
		export = fmt.Sprintf("[%v: %v] %v", r.ErrorCode(), r.err, r.SysMessage())
		// concatenate("[", r.ErrorCode(), ":", r.err.Error(), "] ", r.SysMessage())
		r.export.Store(export)
	}

	return r.export.Load().(string)
}

// ExternalErrMess is adding message from error to GeneralError if it's not GeneralError entity
func (r *generalError) ExternalErrMess(err error) ErrSeed {
	if err == nil {
		return r
	}
	e, ok := ToGeneralError(err)
	if ok && e.(*generalError).mes != "" {
		return r
	}

	mes := func() string {
		if r.mes == "" {
			return err.Error()
		}
		return r.mes
	}

	return &generalError{
		codeGen:  r.codeGen,
		sumCodes: r.sumCodes,
		err:      fmt.Errorf("%w: %v", r.err, err.Error()),
		cause:    err,
		export:   nil,
		code:     r.code,
		codeNote: r.codeNote,
		mes:      mes(),
	}
}

// SubType creates a subtype of GeneralError
func (r generalError) SubType(errSubType string) ErrSeed {
	c := r.codeGen(errSubType)
	return &generalError{
		codeGen:  r.codeGen,
		sumCodes: r.sumCodes,
		err:      fmt.Errorf("%w.%v", r.err, c),
		cause:    r.cause,
		code:     r.sumCodes(r.code, c),
		codeNote: fmt.Sprintf("%v ~> %v:%v", r.codeNote, c, errSubType),
		// concatenate("%v ~> %v:%v", r.codeNote, " ~> ", c, ":",errSubType),
		mes: r.mes,
	}
}

// Message describes error
func (r generalError) Message(message string) ErrSeed {
	return &generalError{
		codeGen:  r.codeGen,
		sumCodes: r.sumCodes,
		err:      r.err,
		cause:    r.cause,
		code:     r.code,
		codeNote: r.codeNote,
		mes:      message,
	}
}

// MessageF describes error with formatting
func (r generalError) MessageF(format string, a ...interface{}) ErrSeed {
	return &generalError{
		codeGen:  r.codeGen,
		sumCodes: r.sumCodes,
		err:      r.err,
		cause:    r.cause,
		code:     r.code,
		codeNote: r.codeNote,
		mes:      fmt.Sprintf(format, a...),
	}
}

// Unwrap is used for errors.Is(...) and returns inner error
func (r generalError) Unwrap() error {
	return r.err
}

// Cause is deprecated. Deprecated: used for backwards compatibility with pkg/errors
func (r generalError) Cause() error {
	return r.cause
}

// Is checks whether this instance is descendant of provided error
func (r generalError) Is(err error) bool {
	e, ok := ToGeneralError(err)
	checkErr := err
	if ok {
		checkErr = e.Unwrap()
	}
	is := errors.Is(r.err, checkErr)
	if !is {
		// for non GeneralError compatibility
		is = errors.Is(r.cause, checkErr)
		if !is {
			// here we should not use checkErr, and direct err because
			// if it was unwrapped early due to GeneralError type,
			// then we will get cause of unwrapped error, and not cause of this one
			hasCause, cok := err.(interface{ Cause() error })
			if cok {
				errors.Is(r.cause, hasCause.Cause())
			}
		}
	}
	return is
}

// ErrorCode returns generated error code
func (r *generalError) ErrorCode() string {
	return r.code
}

// CodeNote returns chain of notes for error codes
func (r *generalError) CodeNote() string {
	return r.codeNote
}

// SysMessage returns message for this error
func (r generalError) SysMessage() string {
	if r.mes == "" {
		return r.codeNote
	}
	return r.mes
}

// FmtResponse formats response according to template with code of this error
func (r generalError) FmtResponse(template string) string {
	return concatenate(template, "\nCODE: ", r.ErrorCode())
}

// Make produces instance of GeneralError
func (r *generalError) Make() GeneralError {
	return r
}

// Produce is for producing subtype of GeneralError
func (r *generalError) Produce() ErrSeed {
	return r
}

func errCodeGen(base string) string {
	h := fnv.New32a()
	_, err := h.Write([]byte(base))
	if err != nil {
		return ""
	}
	v := h.Sum32()
	return uint32toStr36(v)
}

func sumCodes(code1, code2 string) string {
	var base36map = map[rune]uint32{
		'0': 0, '1': 1, '2': 2, '3': 3, '4': 4, '5': 5,
		'6': 6, '7': 7, '8': 8, '9': 9, 'A': 10, 'B': 11,
		'C': 12, 'D': 13, 'E': 14, 'F': 15, 'G': 16, 'H': 17,
		'I': 18, 'J': 19, 'K': 20, 'L': 21, 'M': 22, 'N': 23,
		'O': 24, 'P': 25, 'Q': 26, 'R': 27, 'S': 18, 'T': 29,
		'U': 30, 'V': 31, 'W': 32, 'X': 33, 'Y': 34, 'Z': 35,
	}

	var (
		base36 = [...]string{
			"0", "1", "2", "3", "4", "5",
			"6", "7", "8", "9", "A", "B",
			"C", "D", "E", "F", "G", "H",
			"I", "J", "K", "L", "M", "N",
			"O", "P", "Q", "R", "S", "T",
			"U", "V", "W", "X", "Y", "Z",
		}
	)

	touint32 := func(s string) uint32 {
		radix := uint32(len(base36))
		var res uint32
		var weight uint32 = 1
		for _, k := range s {
			res += weight * base36map[k]
			weight *= radix
		}
		return res
	}
	return uint32toStr36(touint32(code1) + touint32(code2))
}

func uint32toStr36(value uint32) string {
	var (
		base36 = [...]string{
			"0", "1", "2", "3", "4", "5",
			"6", "7", "8", "9", "A", "B",
			"C", "D", "E", "F", "G", "H",
			"I", "J", "K", "L", "M", "N",
			"O", "P", "Q", "R", "S", "T",
			"U", "V", "W", "X", "Y", "Z",
		}
	)

	radix := uint32(len(base36))
	var buffer bytes.Buffer
	for i := value; i > 0; i /= radix {
		k := i % radix
		_, err := buffer.WriteString(base36[k])
		if err != nil {
			return ""
		}
	}
	r := []rune(buffer.String())
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func concatenate(strs ...string) string {
	b := strings.Builder{}
	for _, s := range strs {
		b.WriteString(s)
	}
	return b.String()
}
