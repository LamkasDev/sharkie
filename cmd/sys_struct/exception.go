package sys_struct

// EXCEPTION_CONTINUE_EXECUTION (-1)
//
//	Exception is dismissed. Continue execution at the point where the exception occurred.
//
// EXCEPTION_CONTINUE_SEARCH (0)
//
//	Exception isn't recognized. Continue to search up the stack for a handler, first for containing try-except statements, then for handlers with the next highest precedence.
//
// EXCEPTION_EXECUTE_HANDLER (1)
//
//	Exception is recognized. Transfer control to the exception handler by executing the __except compound statement, then continue execution after the __except block.
//
// https://learn.microsoft.com/en-us/cpp/cpp/try-except-statement?view=msvc-170#remarks
const (
	EXCEPTION_CONTINUE_EXECUTION = 0xFFFFFFFF
	EXCEPTION_CONTINUE_SEARCH    = 0
	EXCEPTION_EXECUTE_HANDLER    = 1
)
