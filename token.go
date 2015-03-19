package clang

// #include <stdlib.h>
// #include "clang-c/Index.h"
//
import "C"

import (
	"fmt"
	"unsafe"
)

// TokenKind describes a kind of token
type TokenKind uint32

const (
	/**
	 * \brief A token that contains some kind of punctuation.
	 */
	TK_Punctuation = C.CXToken_Punctuation

	/**
	 * \brief A language keyword.
	 */
	TK_Keyword = C.CXToken_Keyword

	/**
	 * \brief An identifier (that is not a keyword).
	 */
	TK_Identifier = C.CXToken_Identifier

	/**
	 * \brief A numeric, string, or character literal.
	 */
	TK_Literal = C.CXToken_Literal

	/**
	 * \brief A comment.
	 */
	TK_Comment = C.CXToken_Comment
)

func (tk TokenKind) String() string {
	switch tk {
	case TK_Punctuation:
		return "Punctuation"
	case TK_Keyword:
		return "Keyword"
	case TK_Identifier:
		return "Identifier"
	case TK_Literal:
		return "Literal"
	case TK_Comment:
		return "Comment"
	default:
		panic(fmt.Errorf("clang: invalid TokenKind value (%d)", uint32(tk)))
	}
}

// Token is a single preprocessing token.
type Token struct {
	c  C.CXToken
	tu C.CXTranslationUnit
}

// Kind determines the kind of this token
func (t Token) Kind() TokenKind {
	o := C.clang_getTokenKind(t.c)
	return TokenKind(o)
}

/**
 * \brief Determine the spelling of the given token.
 *
 * The spelling of a token is the textual representation of that token, e.g.,
 * the text of an identifier or keyword.
 */
func (t Token) Spelling() string {
	cstr := cxstring{C.clang_getTokenSpelling(t.tu, t.c)}
	defer cstr.Dispose()
	return cstr.String()
}

/**
 * \brief Retrieve the source location of the given token.
 */
func (t Token) Location() SourceLocation {
	o := C.clang_getTokenLocation(t.tu, t.c)
	return SourceLocation{o}
}

/**
 * \brief Retrieve a source range that covers the given token.
 */
func (t Token) Extent() SourceRange {
	o := C.clang_getTokenExtent(t.tu, t.c)
	return SourceRange{o}
}

/**
 * \brief Tokenize the source code described by the given range into raw
 * lexical tokens.
 *
 * \param TU the translation unit whose text is being tokenized.
 *
 * \param Range the source range in which text should be tokenized. All of the
 * tokens produced by tokenization will fall within this source range,
 *
 * \param Tokens this pointer will be set to point to the array of tokens
 * that occur within the given source range. The returned pointer must be
 * freed with clang_disposeTokens() before the translation unit is destroyed.
 *
 * \param NumTokens will be set to the number of tokens in the \c *Tokens
 * array.
 *
 */
func Tokenize(tu TranslationUnit, src SourceRange) Tokens {
	tokens := Tokens{}
	tokens.tu = tu.c
	var c *C.CXToken
	var n C.uint

	C.clang_tokenize(tu.c, src.c, &c, &n)
	if unsafe.Pointer(c) != nil {
		tokens.c = (*[1 << 26]C.CXToken)(unsafe.Pointer(c))[:n]
	}
	return tokens
}

func (c Cursor) Tokenize() Tokens {
	return Tokenize(c.TranslationUnit(), c.Extent())
}

func (c Cursor) Tokens() (ts []Token) {
	e := c.Extent()
	tokens := Tokenize(c.TranslationUnit(), e)
	if len(tokens.c) == 0 {
		return
	}

	for i := 0; i < tokens.Count(); i++ {
		t := tokens.token(i)
		if t.Extent().IsInside(e) {
			ts = append(ts, *t)
		}
	}
	tokens.Dispose()
	return
}

// an array of tokens
type Tokens struct {
	tu C.CXTranslationUnit
	c  []C.CXToken
	//n  C.uint
}

/**
 * \brief Annotate the given set of tokens by providing cursors for each token
 * that can be mapped to a specific entity within the abstract syntax tree.
 *
 * This token-annotation routine is equivalent to invoking
 * clang_getCursor() for the source locations of each of the
 * tokens. The cursors provided are filtered, so that only those
 * cursors that have a direct correspondence to the token are
 * accepted. For example, given a function call \c f(x),
 * clang_getCursor() would provide the following cursors:
 *
 *   * when the cursor is over the 'f', a DeclRefExpr cursor referring to 'f'.
 *   * when the cursor is over the '(' or the ')', a CallExpr referring to 'f'.
 *   * when the cursor is over the 'x', a DeclRefExpr cursor referring to 'x'.
 *
 * Only the first and last of these cursors will occur within the
 * annotate, since the tokens "f" and "x' directly refer to a function
 * and a variable, respectively, but the parentheses are just a small
 * part of the full syntax of the function call expression, which is
 * not provided as an annotation.
 *
 * \param TU the translation unit that owns the given tokens.
 *
 * \param Tokens the set of tokens to annotate.
 *
 * \param NumTokens the number of tokens in \p Tokens.
 *
 * \param Cursors an array of \p NumTokens cursors, whose contents will be
 * replaced with the cursors corresponding to each token.
 */
func (t Tokens) Annotate() []Cursor {
	n := len(t.c)
	cursors := make([]Cursor, n)
	if n <= 0 {
		return cursors
	}
	c_cursors := make([]C.CXCursor, int(n))
	C.clang_annotateTokens(t.tu, &t.c[0], C.uint(n), &c_cursors[0])
	for i, _ := range cursors {
		cursors[i] = Cursor{c_cursors[i]}
	}
	return cursors
}

/**
 * \brief Free the given set of tokens.
 */
func (t Tokens) Dispose() {
	C.clang_disposeTokens(t.tu, &t.c[0], C.uint(len(t.c)))
}

func (t Tokens) Count() int {
	return len(t.c)
}

func (t Tokens) token(idx int) *Token {
	if idx >= len(t.c) {
		return nil
	}
	return &Token{c: t.c[idx], tu: t.tu}
}

// EOF
