package gmk

import "sync"

// messageFunc is the function Error and Warning call back into to fetch their
// text. A function's result is not expanded again, so routing the text through
// one keeps a $ or an unbalanced parenthesis in it from being read as make
// syntax, which no escaping of a literal $(error ...) argument can do for a
// parenthesis.
const messageFunc = "go-gmk-message"

var (
	messageOnce sync.Once
	messageMu   sync.Mutex
	message     string
)

// Error stops make with msg, as $(error msg) would: make prints it after the
// makefile and line being expanded and exits with status 2. It does not
// return when called from inside an expansion, which is where a [Func] runs.
func Error(msg string) {
	report("error", msg)
}

// Warning prints msg the way $(warning msg) would, after the makefile and line
// being expanded, and returns.
func Warning(msg string) {
	report("warning", msg)
}

func report(fn, msg string) {
	messageOnce.Do(func() {
		AddFunction(messageFunc, func(string, []string) string {
			messageMu.Lock()
			defer messageMu.Unlock()
			return message
		}, 0, 0, FuncDefault)
	})

	messageMu.Lock()
	message = msg
	messageMu.Unlock()

	// A call with no arguments still needs the space: $(name) is a variable.
	Expand("$(" + fn + " $(" + messageFunc + " ))")
}
