package gunit

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/gunit/gunit/generate"
)

type Fixture struct{ T T }

func NewFixture(t *testing.T) *Fixture {
	return &Fixture{T: NewTWrapper(t)}
}

// So is a convenience method for reporting assertion failure messages,
// say from the assertion functions found in github.com/smartystreets/assertions/should.
// Example: self.So(actual, should.Equal, expected)
func (self *Fixture) So(actual interface{}, assert func(actual interface{}, expected ...interface{}) string, expected ...interface{}) {
	if ok, failure := assertions.So(actual, assert, expected...); !ok {
		self.T.Fail()
		self.T.Log("\t" + strings.Replace(failure, "\n", "\n\t\t", -1))
	}
}

func (self *Fixture) Finalize() {
	self.T.(finalizer).finalize()
}

//////////////////////////////////////////////////////////////////////////////

// T defines all methods from *testing.T that we need.
type T interface {
	SkipNow()
	Fail()
	Failed() bool
	Log(args ...interface{})
}

type finalizer interface {
	finalize()
}

/////////////////////////////////////////////////////////////////////////////

type TWrapper struct {
	T
	log     *bytes.Buffer
	skipped bool
}

func NewTWrapper(t T) *TWrapper {
	return &TWrapper{T: t, log: &bytes.Buffer{}}
}

func (self *TWrapper) finalize() {
	if verbose || self.T.Failed() {
		fmt.Fprintln(out, self.log.String())
	}
	if self.skipped {
		self.T.SkipNow()
	}
}

func (self *TWrapper) Error(args ...interface{}) {
	self.Log(args...)
	self.T.Fail()
}
func (self *TWrapper) Errorf(message string, args ...interface{}) {
	self.Logf(message, args...)
	self.T.Fail()
}

func (self *TWrapper) Skip(args ...interface{}) {
	self.Log(args...)
	self.skipped = true
}
func (self *TWrapper) Skipf(message string, args ...interface{}) {
	self.Logf(message, args...)
	self.skipped = true
}

func (self *TWrapper) Logf(message string, args ...interface{}) {
	self.log.WriteString("\t" + fmt.Sprintf(message, args...))

}
func (self *TWrapper) Log(args ...interface{}) {
	self.log.WriteString("\t" + fmt.Sprintln(args...))
}

//////////////////////////////////////////////////////////////////////////////

// Validate ensures that the generated checksums match the existing *.go files actually on disk.
// If there is a mismatch, os.Exit(>0) is called to signal the problem and prevent tests from running.
// This is simply a measure to prevent well-meaning users from forgetting to regenerate test code
// whenever a *.go file changes. This function is only intended to be called by code generated by
// the command at github.com/smartystreets/gunit/gunit.
func Validate(checksum string) {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		exit(1, "Unable to resolve the test file from runtime.Caller(...).")
	}
	current, err := generate.Checksum(filepath.Dir(file))
	if err != nil {
		exit(1, "Could not calculate checksum of current go files. Error: %s", err.Error())
	}
	if checksum != current {
		exit(1, "The checksum provided (%d) does not match the current file listing (%d). Please re-run the `gunit` command and try again.", checksum, current)
	}
}

var out io.Writer = os.Stdout
var verbose = testing.Verbose()

func exit(status int, message string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, message, args...)
	os.Exit(status)
}
