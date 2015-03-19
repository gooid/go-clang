package clang

// #include <stdlib.h>
// #include "clang-c/Index.h"
//
import "C"

import "fmt"

// SourceRange identifies a half-open character range in the source code.
//
// Use clang_getRangeStart() and clang_getRangeEnd() to retrieve the
// starting and end locations from a source range, respectively.
type SourceRange struct {
	c C.CXSourceRange
}

// NewNullRange creates a NULL (invalid) source range.
func NewNullRange() SourceRange {
	return SourceRange{C.clang_getNullRange()}
}

// NewRange creates a source range given the beginning and ending source
// locations.
func NewRange(beg, end SourceLocation) SourceRange {
	o := C.clang_getRange(beg.c, end.c)
	return SourceRange{o}
}

// EqualRanges determines whether two ranges are equivalent.
func (r1 SourceRange) IsEqual(r2 SourceRange) bool {
	o := C.clang_equalRanges(r1.c, r2.c)
	if o != C.uint(0) {
		return true
	}
	return false
}

// IsNull checks if the underlying source range is null.
func (r SourceRange) IsNull() bool {
	o := C.clang_Range_isNull(r.c)
	if o != C.int(0) {
		return true
	}
	return false
}

/**
 * \brief Retrieve a source location representing the first character within a
 * source range.
 */
func (s SourceRange) Start() SourceLocation {
	o := C.clang_getRangeStart(s.c)
	return SourceLocation{o}
}

/**
 * \brief Retrieve a source location representing the last character within a
 * source range.
 */
func (s SourceRange) End() SourceLocation {
	o := C.clang_getRangeEnd(s.c)
	return SourceLocation{o}
}

// Intersect
func (s SourceRange) intersect(i SourceRange) SourceRange {
	fS, _, _, offsetS := s.Start().GetFileLocation()
	fI, _, _, offsetI := i.Start().GetFileLocation()
	if fS.Name() == fI.Name() {

		var sl, el SourceLocation

		start := offsetS
		if start < offsetI {
			start = offsetI
			sl = i.Start()
		} else {
			sl = s.Start()
		}

		fS, _, _, offsetS = s.End().GetFileLocation()
		fI, _, _, offsetI = i.End().GetFileLocation()
		if fS.Name() == fI.Name() {

			end := offsetS
			if end > offsetI {
				end = offsetI
				el = i.End()
			} else {
				el = s.End()
			}

			if start < end {
				return NewRange(sl, el)
			}
		}
	}
	return NewNullRange()
}

// IsInside
func (s SourceRange) IsInside(e SourceRange) bool {
	return s.intersect(e).IsEqual(s)
}

// String
func (s SourceRange) String() string {
	if s.IsNull() {
		return "null"
	}
	f, line, column, offset := s.Start().GetFileLocation()
	str := fmt.Sprint(f.Name(), ":", line, ":", column, "(", offset, ")")
	_, line, column, offset = s.End().GetFileLocation()
	return str + fmt.Sprint(" - :", line, ":", column, "(", offset, ")")
}
