package clang

// #include <stdlib.h>
// #include "clang-c/Index.h"
// extern int goClangFindVisitor(void*, CXCursor, CXSourceRange);
import "C"

import "unsafe"

type FindVisitor func(interface{}, Cursor, SourceRange) bool
type FindContext struct {
	context interface{}
	visitor FindVisitor
}

func findVisitorToCursorAndRangeVisitor(context interface{}, visitor FindVisitor) C.CXCursorAndRangeVisitor {
	var ret C.CXCursorAndRangeVisitor

	c := new(FindContext)
	c.context = context
	c.visitor = visitor
	ret.context = unsafe.Pointer(c)
	ret.visit = (*[0]byte)(C.goClangFindVisitor)

	return ret
}

//export goClangFindVisitor
func goClangFindVisitor(p unsafe.Pointer, cursor C.CXCursor, source C.CXSourceRange) C.int {
	fc := (*FindContext)(p)
	if fc.visitor(fc.context, Cursor{c: cursor}, SourceRange{c: source}) {
		return C.int(C.CXVisit_Continue)
	} else {
		return C.int(C.CXVisit_Break)
	}
}
