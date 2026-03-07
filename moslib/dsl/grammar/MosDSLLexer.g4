lexer grammar MosDSLLexer;

// =============================================================================
// i18n (Internationalization) — Keyword Localization
// =============================================================================
//
// All keyword tokens (FEATURE, SCENARIO, GIVEN, WHEN, THEN, GROUP, SPEC,
// INCLUDE, BACKGROUND, TRUE, FALSE) are English defaults. The Mos DSL supports
// localized keywords (e.g., Spanish "dado" for "given") via a preprocessing
// layer:
//
// 1. The vocabulary artifact (e.g., .mos/vocabulary/default.mos) defines
//    keyword mappings in a `keywords { ... }` block.
// 2. Before lexing, a preprocessor translates human-protocol strings to
//    machine-protocol (English) equivalents.
// 3. The lexer always sees English keywords; the grammar remains simple.
//
// This approach avoids parametric keywords and grammar complexity. The
// vocabulary artifact provides the keyword mapping; consumers apply it
// before invoking the ANTLR parser.
//
// =============================================================================

// -----------------------------------------------------------------------------
// Step block openers — must appear before GIVEN/WHEN/THEN to take precedence.
// When "given", "when", or "then" is followed by optional whitespace and '{',
// we match the combined token and push into StepText mode for free-form step
// content. Note: "when" as a nested block name (e.g. "when { artifact_kind = ... }"
// in rules) is lexed as WHEN_OPEN too — such files are not supported by the
// ANTLR grammar; the hand-written parser handles both via context.
// -----------------------------------------------------------------------------
GIVEN_OPEN : 'given' WS_CHARS* '{' -> pushMode(StepText) ;
WHEN_OPEN   : 'when'  WS_CHARS* '{' -> pushMode(StepText) ;
THEN_OPEN   : 'then'  WS_CHARS* '{' -> pushMode(StepText) ;

// -----------------------------------------------------------------------------
// Keywords (English defaults; i18n via preprocessing)
// -----------------------------------------------------------------------------
FEATURE    : 'feature' ;
BACKGROUND : 'background' ;
SCENARIO   : 'scenario' ;
GIVEN      : 'given' ;
WHEN       : 'when' ;
THEN       : 'then' ;
GROUP      : 'group' ;
SPEC       : 'spec' ;
INCLUDE    : 'include' ;
TRUE       : 'true' ;
FALSE      : 'false' ;

// -----------------------------------------------------------------------------
// Punctuation
// -----------------------------------------------------------------------------
LBRACE    : '{' ;
RBRACE    : '}' ;
LBRACKET  : '[' ;
RBRACKET  : ']' ;
EQUALS    : '=' ;
COMMA     : ',' ;

// -----------------------------------------------------------------------------
// Strings — triple-quoted must precede double-quoted
// -----------------------------------------------------------------------------
TRIPLE_STRING : '"""' (.)*? '"""' ;
STRING        : '"' ( ESC | ~["\\\r\n] )* '"' ;

fragment ESC : '\\' ( [nrt"\\] | UNICODE_ESC ) ;
fragment UNICODE_ESC : 'u' HEX HEX HEX HEX ;
fragment HEX : [0-9a-fA-F] ;

// -----------------------------------------------------------------------------
// Numbers and datetime
// -----------------------------------------------------------------------------
DATETIME : [0-9][0-9][0-9][0-9] '-' [0-9][0-9] '-' [0-9][0-9] 'T' [0-9][0-9] ( ':' [0-9][0-9] ( ':' [0-9][0-9] ( '.' [0-9]+ )? )? )? ( 'Z' | ( '+' | '-' ) [0-9][0-9] ( ':' [0-9][0-9] )? ) ;
FLOAT    : '-'? [0-9]+ '.' [0-9]+ ;
INT      : '-'? [0-9]+ ;

// -----------------------------------------------------------------------------
// Identifier — letter followed by letters, digits, underscore, hyphen
// -----------------------------------------------------------------------------
IDENT : [a-zA-Z] ( [a-zA-Z0-9_-] )* ;

// -----------------------------------------------------------------------------
// Trivia — skipped
// -----------------------------------------------------------------------------
COMMENT : '#' ~[\r\n]* -> skip ;
WS      : [ \t\r]+ -> skip ;
NL      : [\r\n]+ -> skip ;

// -----------------------------------------------------------------------------
// Fragment for step opener whitespace
// -----------------------------------------------------------------------------
fragment WS_CHARS : [ \t\r\n] ;

// =============================================================================
// StepText mode — active inside given { }, when { }, then { }
// Free-form step lines until closing '}'. Step text is line-based: each line
// is one STEP_TEXT token; '}' is only recognized when it appears at the start
// of a line (after optional whitespace), matching the hand-written lexer.
// STEP_RBRACE must match [ \t]* '}' so "      }" is recognized as closing.
// =============================================================================
mode StepText ;
STEP_RBRACE : [ \t]* '}' -> popMode ;
STEP_TEXT   : ~[\r\n]+ ;
STEP_NL     : [\r\n]+ -> skip ;
