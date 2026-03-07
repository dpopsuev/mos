parser grammar MosDSLParser;

options { tokenVocab = MosDSLLexer; }

// -----------------------------------------------------------------------------
// File
// -----------------------------------------------------------------------------
file : artifact EOF ;

// Top-level artifact: IDENT (config, contract, rule, ...) or FEATURE (standalone)
artifact : artifactType STRING? block ;

artifactType
  : IDENT
  | FEATURE
  ;

block : LBRACE blockItem* RBRACE ;

blockItem
  : field
  | featureBlock
  | specBlock
  | nestedBlock
  ;

// Field key can be IDENT or any keyword (e.g. given = "...", when = "...")
field : key EQUALS value ;

key
  : IDENT
  | FEATURE
  | BACKGROUND
  | SCENARIO
  | GIVEN
  | WHEN
  | THEN
  | GROUP
  | SPEC
  | INCLUDE
  ;

value
  : STRING
  | TRIPLE_STRING
  | INT
  | FLOAT
  | boolean
  | DATETIME
  | list
  | inlineTable
  ;

boolean : TRUE | FALSE ;

list : LBRACKET ( value ( COMMA value )* COMMA? )? RBRACKET ;

inlineTable : LBRACE ( field ( COMMA field )* COMMA? )? RBRACE ;

// Nested block: block name can be IDENT or keyword (e.g. when { }, harness { })
nestedBlock : blockName STRING? block ;

blockName
  : IDENT
  | FEATURE
  | BACKGROUND
  | SCENARIO
  | GIVEN
  | WHEN
  | THEN
  | GROUP
  | SPEC
  | INCLUDE
  ;

// -----------------------------------------------------------------------------
// Spec block
// -----------------------------------------------------------------------------
specBlock : SPEC LBRACE ( includeDir | featureBlock )* RBRACE ;

includeDir : INCLUDE STRING ;

// -----------------------------------------------------------------------------
// Feature block
// -----------------------------------------------------------------------------
featureBlock : FEATURE STRING? LBRACE backgroundBlock? ( groupBlock | scenarioBlock )* RBRACE ;

backgroundBlock : BACKGROUND LBRACE GIVEN_OPEN stepLine* STEP_RBRACE RBRACE ;

groupBlock : GROUP STRING? LBRACE scenarioBlock* RBRACE ;

scenarioBlock : SCENARIO STRING? LBRACE scenarioContent RBRACE ;

scenarioContent : field* givenBlock? whenBlock? thenBlock? ;

givenBlock : GIVEN_OPEN stepLine* STEP_RBRACE ;
whenBlock  : WHEN_OPEN  stepLine* STEP_RBRACE;
thenBlock  : THEN_OPEN   stepLine* STEP_RBRACE;

stepLine : STEP_TEXT ;
