
package antlrgen // MosDSLParser
import (
	"fmt"
	"strconv"
  	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}


type MosDSLParser struct {
	*antlr.BaseParser
}

var MosDSLParserParserStaticData struct {
  once                   sync.Once
  serializedATN          []int32
  LiteralNames           []string
  SymbolicNames          []string
  RuleNames              []string
  PredictionContextCache *antlr.PredictionContextCache
  atn                    *antlr.ATN
  decisionToDFA          []*antlr.DFA
}

func mosdslparserParserInit() {
  staticData := &MosDSLParserParserStaticData
  staticData.LiteralNames = []string{
    "", "", "", "", "'feature'", "'background'", "'scenario'", "'given'", 
    "'when'", "'then'", "'group'", "'spec'", "'include'", "'true'", "'false'", 
    "'{'", "'}'", "'['", "']'", "'='", "','",
  }
  staticData.SymbolicNames = []string{
    "", "GIVEN_OPEN", "WHEN_OPEN", "THEN_OPEN", "FEATURE", "BACKGROUND", 
    "SCENARIO", "GIVEN", "WHEN", "THEN", "GROUP", "SPEC", "INCLUDE", "TRUE", 
    "FALSE", "LBRACE", "RBRACE", "LBRACKET", "RBRACKET", "EQUALS", "COMMA", 
    "TRIPLE_STRING", "STRING", "DATETIME", "FLOAT", "INT", "IDENT", "COMMENT", 
    "WS", "NL", "STEP_RBRACE", "STEP_TEXT", "STEP_NL",
  }
  staticData.RuleNames = []string{
    "file", "artifact", "artifactType", "block", "blockItem", "field", "key", 
    "value", "boolean", "list", "inlineTable", "nestedBlock", "blockName", 
    "specBlock", "includeDir", "featureBlock", "backgroundBlock", "groupBlock", 
    "scenarioBlock", "scenarioContent", "givenBlock", "whenBlock", "thenBlock", 
    "stepLine",
  }
  staticData.PredictionContextCache = antlr.NewPredictionContextCache()
  staticData.serializedATN = []int32{
	4, 1, 32, 241, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7, 
	4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7, 
	10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15, 
	2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 2, 20, 7, 20, 2, 
	21, 7, 21, 2, 22, 7, 22, 2, 23, 7, 23, 1, 0, 1, 0, 1, 0, 1, 1, 1, 1, 3, 
	1, 54, 8, 1, 1, 1, 1, 1, 1, 2, 1, 2, 1, 3, 1, 3, 5, 3, 62, 8, 3, 10, 3, 
	12, 3, 65, 9, 3, 1, 3, 1, 3, 1, 4, 1, 4, 1, 4, 1, 4, 3, 4, 73, 8, 4, 1, 
	5, 1, 5, 1, 5, 1, 5, 1, 6, 1, 6, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 
	7, 1, 7, 3, 7, 89, 8, 7, 1, 8, 1, 8, 1, 9, 1, 9, 1, 9, 1, 9, 5, 9, 97, 
	8, 9, 10, 9, 12, 9, 100, 9, 9, 1, 9, 3, 9, 103, 8, 9, 3, 9, 105, 8, 9, 
	1, 9, 1, 9, 1, 10, 1, 10, 1, 10, 1, 10, 5, 10, 113, 8, 10, 10, 10, 12, 
	10, 116, 9, 10, 1, 10, 3, 10, 119, 8, 10, 3, 10, 121, 8, 10, 1, 10, 1, 
	10, 1, 11, 1, 11, 3, 11, 127, 8, 11, 1, 11, 1, 11, 1, 12, 1, 12, 1, 13, 
	1, 13, 1, 13, 1, 13, 5, 13, 137, 8, 13, 10, 13, 12, 13, 140, 9, 13, 1, 
	13, 1, 13, 1, 14, 1, 14, 1, 14, 1, 15, 1, 15, 3, 15, 149, 8, 15, 1, 15, 
	1, 15, 3, 15, 153, 8, 15, 1, 15, 1, 15, 5, 15, 157, 8, 15, 10, 15, 12, 
	15, 160, 9, 15, 1, 15, 1, 15, 1, 16, 1, 16, 1, 16, 1, 16, 5, 16, 168, 8, 
	16, 10, 16, 12, 16, 171, 9, 16, 1, 16, 1, 16, 1, 16, 1, 17, 1, 17, 3, 17, 
	178, 8, 17, 1, 17, 1, 17, 5, 17, 182, 8, 17, 10, 17, 12, 17, 185, 9, 17, 
	1, 17, 1, 17, 1, 18, 1, 18, 3, 18, 191, 8, 18, 1, 18, 1, 18, 1, 18, 1, 
	18, 1, 19, 5, 19, 198, 8, 19, 10, 19, 12, 19, 201, 9, 19, 1, 19, 3, 19, 
	204, 8, 19, 1, 19, 3, 19, 207, 8, 19, 1, 19, 3, 19, 210, 8, 19, 1, 20, 
	1, 20, 5, 20, 214, 8, 20, 10, 20, 12, 20, 217, 9, 20, 1, 20, 1, 20, 1, 
	21, 1, 21, 5, 21, 223, 8, 21, 10, 21, 12, 21, 226, 9, 21, 1, 21, 1, 21, 
	1, 22, 1, 22, 5, 22, 232, 8, 22, 10, 22, 12, 22, 235, 9, 22, 1, 22, 1, 
	22, 1, 23, 1, 23, 1, 23, 0, 0, 24, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 
	22, 24, 26, 28, 30, 32, 34, 36, 38, 40, 42, 44, 46, 0, 3, 2, 0, 4, 4, 26, 
	26, 2, 0, 4, 12, 26, 26, 1, 0, 13, 14, 252, 0, 48, 1, 0, 0, 0, 2, 51, 1, 
	0, 0, 0, 4, 57, 1, 0, 0, 0, 6, 59, 1, 0, 0, 0, 8, 72, 1, 0, 0, 0, 10, 74, 
	1, 0, 0, 0, 12, 78, 1, 0, 0, 0, 14, 88, 1, 0, 0, 0, 16, 90, 1, 0, 0, 0, 
	18, 92, 1, 0, 0, 0, 20, 108, 1, 0, 0, 0, 22, 124, 1, 0, 0, 0, 24, 130, 
	1, 0, 0, 0, 26, 132, 1, 0, 0, 0, 28, 143, 1, 0, 0, 0, 30, 146, 1, 0, 0, 
	0, 32, 163, 1, 0, 0, 0, 34, 175, 1, 0, 0, 0, 36, 188, 1, 0, 0, 0, 38, 199, 
	1, 0, 0, 0, 40, 211, 1, 0, 0, 0, 42, 220, 1, 0, 0, 0, 44, 229, 1, 0, 0, 
	0, 46, 238, 1, 0, 0, 0, 48, 49, 3, 2, 1, 0, 49, 50, 5, 0, 0, 1, 50, 1, 
	1, 0, 0, 0, 51, 53, 3, 4, 2, 0, 52, 54, 5, 22, 0, 0, 53, 52, 1, 0, 0, 0, 
	53, 54, 1, 0, 0, 0, 54, 55, 1, 0, 0, 0, 55, 56, 3, 6, 3, 0, 56, 3, 1, 0, 
	0, 0, 57, 58, 7, 0, 0, 0, 58, 5, 1, 0, 0, 0, 59, 63, 5, 15, 0, 0, 60, 62, 
	3, 8, 4, 0, 61, 60, 1, 0, 0, 0, 62, 65, 1, 0, 0, 0, 63, 61, 1, 0, 0, 0, 
	63, 64, 1, 0, 0, 0, 64, 66, 1, 0, 0, 0, 65, 63, 1, 0, 0, 0, 66, 67, 5, 
	16, 0, 0, 67, 7, 1, 0, 0, 0, 68, 73, 3, 10, 5, 0, 69, 73, 3, 30, 15, 0, 
	70, 73, 3, 26, 13, 0, 71, 73, 3, 22, 11, 0, 72, 68, 1, 0, 0, 0, 72, 69, 
	1, 0, 0, 0, 72, 70, 1, 0, 0, 0, 72, 71, 1, 0, 0, 0, 73, 9, 1, 0, 0, 0, 
	74, 75, 3, 12, 6, 0, 75, 76, 5, 19, 0, 0, 76, 77, 3, 14, 7, 0, 77, 11, 
	1, 0, 0, 0, 78, 79, 7, 1, 0, 0, 79, 13, 1, 0, 0, 0, 80, 89, 5, 22, 0, 0, 
	81, 89, 5, 21, 0, 0, 82, 89, 5, 25, 0, 0, 83, 89, 5, 24, 0, 0, 84, 89, 
	3, 16, 8, 0, 85, 89, 5, 23, 0, 0, 86, 89, 3, 18, 9, 0, 87, 89, 3, 20, 10, 
	0, 88, 80, 1, 0, 0, 0, 88, 81, 1, 0, 0, 0, 88, 82, 1, 0, 0, 0, 88, 83, 
	1, 0, 0, 0, 88, 84, 1, 0, 0, 0, 88, 85, 1, 0, 0, 0, 88, 86, 1, 0, 0, 0, 
	88, 87, 1, 0, 0, 0, 89, 15, 1, 0, 0, 0, 90, 91, 7, 2, 0, 0, 91, 17, 1, 
	0, 0, 0, 92, 104, 5, 17, 0, 0, 93, 98, 3, 14, 7, 0, 94, 95, 5, 20, 0, 0, 
	95, 97, 3, 14, 7, 0, 96, 94, 1, 0, 0, 0, 97, 100, 1, 0, 0, 0, 98, 96, 1, 
	0, 0, 0, 98, 99, 1, 0, 0, 0, 99, 102, 1, 0, 0, 0, 100, 98, 1, 0, 0, 0, 
	101, 103, 5, 20, 0, 0, 102, 101, 1, 0, 0, 0, 102, 103, 1, 0, 0, 0, 103, 
	105, 1, 0, 0, 0, 104, 93, 1, 0, 0, 0, 104, 105, 1, 0, 0, 0, 105, 106, 1, 
	0, 0, 0, 106, 107, 5, 18, 0, 0, 107, 19, 1, 0, 0, 0, 108, 120, 5, 15, 0, 
	0, 109, 114, 3, 10, 5, 0, 110, 111, 5, 20, 0, 0, 111, 113, 3, 10, 5, 0, 
	112, 110, 1, 0, 0, 0, 113, 116, 1, 0, 0, 0, 114, 112, 1, 0, 0, 0, 114, 
	115, 1, 0, 0, 0, 115, 118, 1, 0, 0, 0, 116, 114, 1, 0, 0, 0, 117, 119, 
	5, 20, 0, 0, 118, 117, 1, 0, 0, 0, 118, 119, 1, 0, 0, 0, 119, 121, 1, 0, 
	0, 0, 120, 109, 1, 0, 0, 0, 120, 121, 1, 0, 0, 0, 121, 122, 1, 0, 0, 0, 
	122, 123, 5, 16, 0, 0, 123, 21, 1, 0, 0, 0, 124, 126, 3, 24, 12, 0, 125, 
	127, 5, 22, 0, 0, 126, 125, 1, 0, 0, 0, 126, 127, 1, 0, 0, 0, 127, 128, 
	1, 0, 0, 0, 128, 129, 3, 6, 3, 0, 129, 23, 1, 0, 0, 0, 130, 131, 7, 1, 
	0, 0, 131, 25, 1, 0, 0, 0, 132, 133, 5, 11, 0, 0, 133, 138, 5, 15, 0, 0, 
	134, 137, 3, 28, 14, 0, 135, 137, 3, 30, 15, 0, 136, 134, 1, 0, 0, 0, 136, 
	135, 1, 0, 0, 0, 137, 140, 1, 0, 0, 0, 138, 136, 1, 0, 0, 0, 138, 139, 
	1, 0, 0, 0, 139, 141, 1, 0, 0, 0, 140, 138, 1, 0, 0, 0, 141, 142, 5, 16, 
	0, 0, 142, 27, 1, 0, 0, 0, 143, 144, 5, 12, 0, 0, 144, 145, 5, 22, 0, 0, 
	145, 29, 1, 0, 0, 0, 146, 148, 5, 4, 0, 0, 147, 149, 5, 22, 0, 0, 148, 
	147, 1, 0, 0, 0, 148, 149, 1, 0, 0, 0, 149, 150, 1, 0, 0, 0, 150, 152, 
	5, 15, 0, 0, 151, 153, 3, 32, 16, 0, 152, 151, 1, 0, 0, 0, 152, 153, 1, 
	0, 0, 0, 153, 158, 1, 0, 0, 0, 154, 157, 3, 34, 17, 0, 155, 157, 3, 36, 
	18, 0, 156, 154, 1, 0, 0, 0, 156, 155, 1, 0, 0, 0, 157, 160, 1, 0, 0, 0, 
	158, 156, 1, 0, 0, 0, 158, 159, 1, 0, 0, 0, 159, 161, 1, 0, 0, 0, 160, 
	158, 1, 0, 0, 0, 161, 162, 5, 16, 0, 0, 162, 31, 1, 0, 0, 0, 163, 164, 
	5, 5, 0, 0, 164, 165, 5, 15, 0, 0, 165, 169, 5, 1, 0, 0, 166, 168, 3, 46, 
	23, 0, 167, 166, 1, 0, 0, 0, 168, 171, 1, 0, 0, 0, 169, 167, 1, 0, 0, 0, 
	169, 170, 1, 0, 0, 0, 170, 172, 1, 0, 0, 0, 171, 169, 1, 0, 0, 0, 172, 
	173, 5, 30, 0, 0, 173, 174, 5, 16, 0, 0, 174, 33, 1, 0, 0, 0, 175, 177, 
	5, 10, 0, 0, 176, 178, 5, 22, 0, 0, 177, 176, 1, 0, 0, 0, 177, 178, 1, 
	0, 0, 0, 178, 179, 1, 0, 0, 0, 179, 183, 5, 15, 0, 0, 180, 182, 3, 36, 
	18, 0, 181, 180, 1, 0, 0, 0, 182, 185, 1, 0, 0, 0, 183, 181, 1, 0, 0, 0, 
	183, 184, 1, 0, 0, 0, 184, 186, 1, 0, 0, 0, 185, 183, 1, 0, 0, 0, 186, 
	187, 5, 16, 0, 0, 187, 35, 1, 0, 0, 0, 188, 190, 5, 6, 0, 0, 189, 191, 
	5, 22, 0, 0, 190, 189, 1, 0, 0, 0, 190, 191, 1, 0, 0, 0, 191, 192, 1, 0, 
	0, 0, 192, 193, 5, 15, 0, 0, 193, 194, 3, 38, 19, 0, 194, 195, 5, 16, 0, 
	0, 195, 37, 1, 0, 0, 0, 196, 198, 3, 10, 5, 0, 197, 196, 1, 0, 0, 0, 198, 
	201, 1, 0, 0, 0, 199, 197, 1, 0, 0, 0, 199, 200, 1, 0, 0, 0, 200, 203, 
	1, 0, 0, 0, 201, 199, 1, 0, 0, 0, 202, 204, 3, 40, 20, 0, 203, 202, 1, 
	0, 0, 0, 203, 204, 1, 0, 0, 0, 204, 206, 1, 0, 0, 0, 205, 207, 3, 42, 21, 
	0, 206, 205, 1, 0, 0, 0, 206, 207, 1, 0, 0, 0, 207, 209, 1, 0, 0, 0, 208, 
	210, 3, 44, 22, 0, 209, 208, 1, 0, 0, 0, 209, 210, 1, 0, 0, 0, 210, 39, 
	1, 0, 0, 0, 211, 215, 5, 1, 0, 0, 212, 214, 3, 46, 23, 0, 213, 212, 1, 
	0, 0, 0, 214, 217, 1, 0, 0, 0, 215, 213, 1, 0, 0, 0, 215, 216, 1, 0, 0, 
	0, 216, 218, 1, 0, 0, 0, 217, 215, 1, 0, 0, 0, 218, 219, 5, 30, 0, 0, 219, 
	41, 1, 0, 0, 0, 220, 224, 5, 2, 0, 0, 221, 223, 3, 46, 23, 0, 222, 221, 
	1, 0, 0, 0, 223, 226, 1, 0, 0, 0, 224, 222, 1, 0, 0, 0, 224, 225, 1, 0, 
	0, 0, 225, 227, 1, 0, 0, 0, 226, 224, 1, 0, 0, 0, 227, 228, 5, 30, 0, 0, 
	228, 43, 1, 0, 0, 0, 229, 233, 5, 3, 0, 0, 230, 232, 3, 46, 23, 0, 231, 
	230, 1, 0, 0, 0, 232, 235, 1, 0, 0, 0, 233, 231, 1, 0, 0, 0, 233, 234, 
	1, 0, 0, 0, 234, 236, 1, 0, 0, 0, 235, 233, 1, 0, 0, 0, 236, 237, 5, 30, 
	0, 0, 237, 45, 1, 0, 0, 0, 238, 239, 5, 31, 0, 0, 239, 47, 1, 0, 0, 0, 
	28, 53, 63, 72, 88, 98, 102, 104, 114, 118, 120, 126, 136, 138, 148, 152, 
	156, 158, 169, 177, 183, 190, 199, 203, 206, 209, 215, 224, 233,
}
  deserializer := antlr.NewATNDeserializer(nil)
  staticData.atn = deserializer.Deserialize(staticData.serializedATN)
  atn := staticData.atn
  staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
  decisionToDFA := staticData.decisionToDFA
  for index, state := range atn.DecisionToState {
    decisionToDFA[index] = antlr.NewDFA(state, index)
  }
}

// MosDSLParserInit initializes any static state used to implement MosDSLParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewMosDSLParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func MosDSLParserInit() {
  staticData := &MosDSLParserParserStaticData
  staticData.once.Do(mosdslparserParserInit)
}

// NewMosDSLParser produces a new parser instance for the optional input antlr.TokenStream.
func NewMosDSLParser(input antlr.TokenStream) *MosDSLParser {
	MosDSLParserInit()
	this := new(MosDSLParser)
	this.BaseParser = antlr.NewBaseParser(input)
  staticData := &MosDSLParserParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "MosDSLParser.g4"

	return this
}


// MosDSLParser tokens.
const (
	MosDSLParserEOF = antlr.TokenEOF
	MosDSLParserGIVEN_OPEN = 1
	MosDSLParserWHEN_OPEN = 2
	MosDSLParserTHEN_OPEN = 3
	MosDSLParserFEATURE = 4
	MosDSLParserBACKGROUND = 5
	MosDSLParserSCENARIO = 6
	MosDSLParserGIVEN = 7
	MosDSLParserWHEN = 8
	MosDSLParserTHEN = 9
	MosDSLParserGROUP = 10
	MosDSLParserSPEC = 11
	MosDSLParserINCLUDE = 12
	MosDSLParserTRUE = 13
	MosDSLParserFALSE = 14
	MosDSLParserLBRACE = 15
	MosDSLParserRBRACE = 16
	MosDSLParserLBRACKET = 17
	MosDSLParserRBRACKET = 18
	MosDSLParserEQUALS = 19
	MosDSLParserCOMMA = 20
	MosDSLParserTRIPLE_STRING = 21
	MosDSLParserSTRING = 22
	MosDSLParserDATETIME = 23
	MosDSLParserFLOAT = 24
	MosDSLParserINT = 25
	MosDSLParserIDENT = 26
	MosDSLParserCOMMENT = 27
	MosDSLParserWS = 28
	MosDSLParserNL = 29
	MosDSLParserSTEP_RBRACE = 30
	MosDSLParserSTEP_TEXT = 31
	MosDSLParserSTEP_NL = 32
)

// MosDSLParser rules.
const (
	MosDSLParserRULE_file = 0
	MosDSLParserRULE_artifact = 1
	MosDSLParserRULE_artifactType = 2
	MosDSLParserRULE_block = 3
	MosDSLParserRULE_blockItem = 4
	MosDSLParserRULE_field = 5
	MosDSLParserRULE_key = 6
	MosDSLParserRULE_value = 7
	MosDSLParserRULE_boolean = 8
	MosDSLParserRULE_list = 9
	MosDSLParserRULE_inlineTable = 10
	MosDSLParserRULE_nestedBlock = 11
	MosDSLParserRULE_blockName = 12
	MosDSLParserRULE_specBlock = 13
	MosDSLParserRULE_includeDir = 14
	MosDSLParserRULE_featureBlock = 15
	MosDSLParserRULE_backgroundBlock = 16
	MosDSLParserRULE_groupBlock = 17
	MosDSLParserRULE_scenarioBlock = 18
	MosDSLParserRULE_scenarioContent = 19
	MosDSLParserRULE_givenBlock = 20
	MosDSLParserRULE_whenBlock = 21
	MosDSLParserRULE_thenBlock = 22
	MosDSLParserRULE_stepLine = 23
)

// IFileContext is an interface to support dynamic dispatch.
type IFileContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Artifact() IArtifactContext
	EOF() antlr.TerminalNode

	// IsFileContext differentiates from other interfaces.
	IsFileContext()
}

type FileContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFileContext() *FileContext {
	var p = new(FileContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_file
	return p
}

func InitEmptyFileContext(p *FileContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_file
}

func (*FileContext) IsFileContext() {}

func NewFileContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FileContext {
	var p = new(FileContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_file

	return p
}

func (s *FileContext) GetParser() antlr.Parser { return s.parser }

func (s *FileContext) Artifact() IArtifactContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IArtifactContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IArtifactContext)
}

func (s *FileContext) EOF() antlr.TerminalNode {
	return s.GetToken(MosDSLParserEOF, 0)
}

func (s *FileContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FileContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *FileContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterFile(s)
	}
}

func (s *FileContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitFile(s)
	}
}




func (p *MosDSLParser) File() (localctx IFileContext) {
	localctx = NewFileContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, MosDSLParserRULE_file)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(48)
		p.Artifact()
	}
	{
		p.SetState(49)
		p.Match(MosDSLParserEOF)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}



errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// IArtifactContext is an interface to support dynamic dispatch.
type IArtifactContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ArtifactType() IArtifactTypeContext
	Block() IBlockContext
	STRING() antlr.TerminalNode

	// IsArtifactContext differentiates from other interfaces.
	IsArtifactContext()
}

type ArtifactContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyArtifactContext() *ArtifactContext {
	var p = new(ArtifactContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_artifact
	return p
}

func InitEmptyArtifactContext(p *ArtifactContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_artifact
}

func (*ArtifactContext) IsArtifactContext() {}

func NewArtifactContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ArtifactContext {
	var p = new(ArtifactContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_artifact

	return p
}

func (s *ArtifactContext) GetParser() antlr.Parser { return s.parser }

func (s *ArtifactContext) ArtifactType() IArtifactTypeContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IArtifactTypeContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IArtifactTypeContext)
}

func (s *ArtifactContext) Block() IBlockContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBlockContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBlockContext)
}

func (s *ArtifactContext) STRING() antlr.TerminalNode {
	return s.GetToken(MosDSLParserSTRING, 0)
}

func (s *ArtifactContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ArtifactContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *ArtifactContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterArtifact(s)
	}
}

func (s *ArtifactContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitArtifact(s)
	}
}




func (p *MosDSLParser) Artifact() (localctx IArtifactContext) {
	localctx = NewArtifactContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, MosDSLParserRULE_artifact)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(51)
		p.ArtifactType()
	}
	p.SetState(53)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if _la == MosDSLParserSTRING {
		{
			p.SetState(52)
			p.Match(MosDSLParserSTRING)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}

	}
	{
		p.SetState(55)
		p.Block()
	}



errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// IArtifactTypeContext is an interface to support dynamic dispatch.
type IArtifactTypeContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENT() antlr.TerminalNode
	FEATURE() antlr.TerminalNode

	// IsArtifactTypeContext differentiates from other interfaces.
	IsArtifactTypeContext()
}

type ArtifactTypeContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyArtifactTypeContext() *ArtifactTypeContext {
	var p = new(ArtifactTypeContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_artifactType
	return p
}

func InitEmptyArtifactTypeContext(p *ArtifactTypeContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_artifactType
}

func (*ArtifactTypeContext) IsArtifactTypeContext() {}

func NewArtifactTypeContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ArtifactTypeContext {
	var p = new(ArtifactTypeContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_artifactType

	return p
}

func (s *ArtifactTypeContext) GetParser() antlr.Parser { return s.parser }

func (s *ArtifactTypeContext) IDENT() antlr.TerminalNode {
	return s.GetToken(MosDSLParserIDENT, 0)
}

func (s *ArtifactTypeContext) FEATURE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserFEATURE, 0)
}

func (s *ArtifactTypeContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ArtifactTypeContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *ArtifactTypeContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterArtifactType(s)
	}
}

func (s *ArtifactTypeContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitArtifactType(s)
	}
}




func (p *MosDSLParser) ArtifactType() (localctx IArtifactTypeContext) {
	localctx = NewArtifactTypeContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, MosDSLParserRULE_artifactType)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(57)
		_la = p.GetTokenStream().LA(1)

		if !(_la == MosDSLParserFEATURE || _la == MosDSLParserIDENT) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}



errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// IBlockContext is an interface to support dynamic dispatch.
type IBlockContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LBRACE() antlr.TerminalNode
	RBRACE() antlr.TerminalNode
	AllBlockItem() []IBlockItemContext
	BlockItem(i int) IBlockItemContext

	// IsBlockContext differentiates from other interfaces.
	IsBlockContext()
}

type BlockContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBlockContext() *BlockContext {
	var p = new(BlockContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_block
	return p
}

func InitEmptyBlockContext(p *BlockContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_block
}

func (*BlockContext) IsBlockContext() {}

func NewBlockContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *BlockContext {
	var p = new(BlockContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_block

	return p
}

func (s *BlockContext) GetParser() antlr.Parser { return s.parser }

func (s *BlockContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserLBRACE, 0)
}

func (s *BlockContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserRBRACE, 0)
}

func (s *BlockContext) AllBlockItem() []IBlockItemContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IBlockItemContext); ok {
			len++
		}
	}

	tst := make([]IBlockItemContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IBlockItemContext); ok {
			tst[i] = t.(IBlockItemContext)
			i++
		}
	}

	return tst
}

func (s *BlockContext) BlockItem(i int) IBlockItemContext {
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBlockItemContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBlockItemContext)
}

func (s *BlockContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BlockContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *BlockContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterBlock(s)
	}
}

func (s *BlockContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitBlock(s)
	}
}




func (p *MosDSLParser) Block() (localctx IBlockContext) {
	localctx = NewBlockContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, MosDSLParserRULE_block)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(59)
		p.Match(MosDSLParserLBRACE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	p.SetState(63)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	for ((int64(_la) & ^0x3f) == 0 && ((int64(1) << _la) & 67117040) != 0) {
		{
			p.SetState(60)
			p.BlockItem()
		}


		p.SetState(65)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
	    	goto errorExit
	    }
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(66)
		p.Match(MosDSLParserRBRACE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}



errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// IBlockItemContext is an interface to support dynamic dispatch.
type IBlockItemContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Field() IFieldContext
	FeatureBlock() IFeatureBlockContext
	SpecBlock() ISpecBlockContext
	NestedBlock() INestedBlockContext

	// IsBlockItemContext differentiates from other interfaces.
	IsBlockItemContext()
}

type BlockItemContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBlockItemContext() *BlockItemContext {
	var p = new(BlockItemContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_blockItem
	return p
}

func InitEmptyBlockItemContext(p *BlockItemContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_blockItem
}

func (*BlockItemContext) IsBlockItemContext() {}

func NewBlockItemContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *BlockItemContext {
	var p = new(BlockItemContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_blockItem

	return p
}

func (s *BlockItemContext) GetParser() antlr.Parser { return s.parser }

func (s *BlockItemContext) Field() IFieldContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFieldContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFieldContext)
}

func (s *BlockItemContext) FeatureBlock() IFeatureBlockContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFeatureBlockContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFeatureBlockContext)
}

func (s *BlockItemContext) SpecBlock() ISpecBlockContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISpecBlockContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISpecBlockContext)
}

func (s *BlockItemContext) NestedBlock() INestedBlockContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(INestedBlockContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(INestedBlockContext)
}

func (s *BlockItemContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BlockItemContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *BlockItemContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterBlockItem(s)
	}
}

func (s *BlockItemContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitBlockItem(s)
	}
}




func (p *MosDSLParser) BlockItem() (localctx IBlockItemContext) {
	localctx = NewBlockItemContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, MosDSLParserRULE_blockItem)
	p.SetState(72)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 2, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(68)
			p.Field()
		}


	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(69)
			p.FeatureBlock()
		}


	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(70)
			p.SpecBlock()
		}


	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(71)
			p.NestedBlock()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}


errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// IFieldContext is an interface to support dynamic dispatch.
type IFieldContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Key() IKeyContext
	EQUALS() antlr.TerminalNode
	Value() IValueContext

	// IsFieldContext differentiates from other interfaces.
	IsFieldContext()
}

type FieldContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFieldContext() *FieldContext {
	var p = new(FieldContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_field
	return p
}

func InitEmptyFieldContext(p *FieldContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_field
}

func (*FieldContext) IsFieldContext() {}

func NewFieldContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FieldContext {
	var p = new(FieldContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_field

	return p
}

func (s *FieldContext) GetParser() antlr.Parser { return s.parser }

func (s *FieldContext) Key() IKeyContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IKeyContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IKeyContext)
}

func (s *FieldContext) EQUALS() antlr.TerminalNode {
	return s.GetToken(MosDSLParserEQUALS, 0)
}

func (s *FieldContext) Value() IValueContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueContext)
}

func (s *FieldContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FieldContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *FieldContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterField(s)
	}
}

func (s *FieldContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitField(s)
	}
}




func (p *MosDSLParser) Field() (localctx IFieldContext) {
	localctx = NewFieldContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, MosDSLParserRULE_field)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(74)
		p.Key()
	}
	{
		p.SetState(75)
		p.Match(MosDSLParserEQUALS)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(76)
		p.Value()
	}



errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// IKeyContext is an interface to support dynamic dispatch.
type IKeyContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENT() antlr.TerminalNode
	FEATURE() antlr.TerminalNode
	BACKGROUND() antlr.TerminalNode
	SCENARIO() antlr.TerminalNode
	GIVEN() antlr.TerminalNode
	WHEN() antlr.TerminalNode
	THEN() antlr.TerminalNode
	GROUP() antlr.TerminalNode
	SPEC() antlr.TerminalNode
	INCLUDE() antlr.TerminalNode

	// IsKeyContext differentiates from other interfaces.
	IsKeyContext()
}

type KeyContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyKeyContext() *KeyContext {
	var p = new(KeyContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_key
	return p
}

func InitEmptyKeyContext(p *KeyContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_key
}

func (*KeyContext) IsKeyContext() {}

func NewKeyContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *KeyContext {
	var p = new(KeyContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_key

	return p
}

func (s *KeyContext) GetParser() antlr.Parser { return s.parser }

func (s *KeyContext) IDENT() antlr.TerminalNode {
	return s.GetToken(MosDSLParserIDENT, 0)
}

func (s *KeyContext) FEATURE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserFEATURE, 0)
}

func (s *KeyContext) BACKGROUND() antlr.TerminalNode {
	return s.GetToken(MosDSLParserBACKGROUND, 0)
}

func (s *KeyContext) SCENARIO() antlr.TerminalNode {
	return s.GetToken(MosDSLParserSCENARIO, 0)
}

func (s *KeyContext) GIVEN() antlr.TerminalNode {
	return s.GetToken(MosDSLParserGIVEN, 0)
}

func (s *KeyContext) WHEN() antlr.TerminalNode {
	return s.GetToken(MosDSLParserWHEN, 0)
}

func (s *KeyContext) THEN() antlr.TerminalNode {
	return s.GetToken(MosDSLParserTHEN, 0)
}

func (s *KeyContext) GROUP() antlr.TerminalNode {
	return s.GetToken(MosDSLParserGROUP, 0)
}

func (s *KeyContext) SPEC() antlr.TerminalNode {
	return s.GetToken(MosDSLParserSPEC, 0)
}

func (s *KeyContext) INCLUDE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserINCLUDE, 0)
}

func (s *KeyContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *KeyContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *KeyContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterKey(s)
	}
}

func (s *KeyContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitKey(s)
	}
}




func (p *MosDSLParser) Key() (localctx IKeyContext) {
	localctx = NewKeyContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, MosDSLParserRULE_key)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(78)
		_la = p.GetTokenStream().LA(1)

		if !(((int64(_la) & ^0x3f) == 0 && ((int64(1) << _la) & 67117040) != 0)) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}



errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// IValueContext is an interface to support dynamic dispatch.
type IValueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STRING() antlr.TerminalNode
	TRIPLE_STRING() antlr.TerminalNode
	INT() antlr.TerminalNode
	FLOAT() antlr.TerminalNode
	Boolean() IBooleanContext
	DATETIME() antlr.TerminalNode
	List() IListContext
	InlineTable() IInlineTableContext

	// IsValueContext differentiates from other interfaces.
	IsValueContext()
}

type ValueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyValueContext() *ValueContext {
	var p = new(ValueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_value
	return p
}

func InitEmptyValueContext(p *ValueContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_value
}

func (*ValueContext) IsValueContext() {}

func NewValueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ValueContext {
	var p = new(ValueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_value

	return p
}

func (s *ValueContext) GetParser() antlr.Parser { return s.parser }

func (s *ValueContext) STRING() antlr.TerminalNode {
	return s.GetToken(MosDSLParserSTRING, 0)
}

func (s *ValueContext) TRIPLE_STRING() antlr.TerminalNode {
	return s.GetToken(MosDSLParserTRIPLE_STRING, 0)
}

func (s *ValueContext) INT() antlr.TerminalNode {
	return s.GetToken(MosDSLParserINT, 0)
}

func (s *ValueContext) FLOAT() antlr.TerminalNode {
	return s.GetToken(MosDSLParserFLOAT, 0)
}

func (s *ValueContext) Boolean() IBooleanContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBooleanContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBooleanContext)
}

func (s *ValueContext) DATETIME() antlr.TerminalNode {
	return s.GetToken(MosDSLParserDATETIME, 0)
}

func (s *ValueContext) List() IListContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IListContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IListContext)
}

func (s *ValueContext) InlineTable() IInlineTableContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IInlineTableContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IInlineTableContext)
}

func (s *ValueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ValueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *ValueContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterValue(s)
	}
}

func (s *ValueContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitValue(s)
	}
}




func (p *MosDSLParser) Value() (localctx IValueContext) {
	localctx = NewValueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, MosDSLParserRULE_value)
	p.SetState(88)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MosDSLParserSTRING:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(80)
			p.Match(MosDSLParserSTRING)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}


	case MosDSLParserTRIPLE_STRING:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(81)
			p.Match(MosDSLParserTRIPLE_STRING)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}


	case MosDSLParserINT:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(82)
			p.Match(MosDSLParserINT)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}


	case MosDSLParserFLOAT:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(83)
			p.Match(MosDSLParserFLOAT)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}


	case MosDSLParserTRUE, MosDSLParserFALSE:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(84)
			p.Boolean()
		}


	case MosDSLParserDATETIME:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(85)
			p.Match(MosDSLParserDATETIME)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}


	case MosDSLParserLBRACKET:
		p.EnterOuterAlt(localctx, 7)
		{
			p.SetState(86)
			p.List()
		}


	case MosDSLParserLBRACE:
		p.EnterOuterAlt(localctx, 8)
		{
			p.SetState(87)
			p.InlineTable()
		}



	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}


errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// IBooleanContext is an interface to support dynamic dispatch.
type IBooleanContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TRUE() antlr.TerminalNode
	FALSE() antlr.TerminalNode

	// IsBooleanContext differentiates from other interfaces.
	IsBooleanContext()
}

type BooleanContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBooleanContext() *BooleanContext {
	var p = new(BooleanContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_boolean
	return p
}

func InitEmptyBooleanContext(p *BooleanContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_boolean
}

func (*BooleanContext) IsBooleanContext() {}

func NewBooleanContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *BooleanContext {
	var p = new(BooleanContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_boolean

	return p
}

func (s *BooleanContext) GetParser() antlr.Parser { return s.parser }

func (s *BooleanContext) TRUE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserTRUE, 0)
}

func (s *BooleanContext) FALSE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserFALSE, 0)
}

func (s *BooleanContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BooleanContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *BooleanContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterBoolean(s)
	}
}

func (s *BooleanContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitBoolean(s)
	}
}




func (p *MosDSLParser) Boolean() (localctx IBooleanContext) {
	localctx = NewBooleanContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, MosDSLParserRULE_boolean)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(90)
		_la = p.GetTokenStream().LA(1)

		if !(_la == MosDSLParserTRUE || _la == MosDSLParserFALSE) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}



errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// IListContext is an interface to support dynamic dispatch.
type IListContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LBRACKET() antlr.TerminalNode
	RBRACKET() antlr.TerminalNode
	AllValue() []IValueContext
	Value(i int) IValueContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsListContext differentiates from other interfaces.
	IsListContext()
}

type ListContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyListContext() *ListContext {
	var p = new(ListContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_list
	return p
}

func InitEmptyListContext(p *ListContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_list
}

func (*ListContext) IsListContext() {}

func NewListContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ListContext {
	var p = new(ListContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_list

	return p
}

func (s *ListContext) GetParser() antlr.Parser { return s.parser }

func (s *ListContext) LBRACKET() antlr.TerminalNode {
	return s.GetToken(MosDSLParserLBRACKET, 0)
}

func (s *ListContext) RBRACKET() antlr.TerminalNode {
	return s.GetToken(MosDSLParserRBRACKET, 0)
}

func (s *ListContext) AllValue() []IValueContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IValueContext); ok {
			len++
		}
	}

	tst := make([]IValueContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IValueContext); ok {
			tst[i] = t.(IValueContext)
			i++
		}
	}

	return tst
}

func (s *ListContext) Value(i int) IValueContext {
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueContext)
}

func (s *ListContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(MosDSLParserCOMMA)
}

func (s *ListContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(MosDSLParserCOMMA, i)
}

func (s *ListContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ListContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *ListContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterList(s)
	}
}

func (s *ListContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitList(s)
	}
}




func (p *MosDSLParser) List() (localctx IListContext) {
	localctx = NewListContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, MosDSLParserRULE_list)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(92)
		p.Match(MosDSLParserLBRACKET)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	p.SetState(104)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if ((int64(_la) & ^0x3f) == 0 && ((int64(1) << _la) & 65200128) != 0) {
		{
			p.SetState(93)
			p.Value()
		}
		p.SetState(98)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 4, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(94)
					p.Match(MosDSLParserCOMMA)
					if p.HasError() {
							// Recognition error - abort rule
							goto errorExit
					}
				}
				{
					p.SetState(95)
					p.Value()
				}


			}
			p.SetState(100)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
		    	goto errorExit
		    }
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 4, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		p.SetState(102)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)


		if _la == MosDSLParserCOMMA {
			{
				p.SetState(101)
				p.Match(MosDSLParserCOMMA)
				if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
				}
			}

		}

	}
	{
		p.SetState(106)
		p.Match(MosDSLParserRBRACKET)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}



errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// IInlineTableContext is an interface to support dynamic dispatch.
type IInlineTableContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LBRACE() antlr.TerminalNode
	RBRACE() antlr.TerminalNode
	AllField() []IFieldContext
	Field(i int) IFieldContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsInlineTableContext differentiates from other interfaces.
	IsInlineTableContext()
}

type InlineTableContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyInlineTableContext() *InlineTableContext {
	var p = new(InlineTableContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_inlineTable
	return p
}

func InitEmptyInlineTableContext(p *InlineTableContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_inlineTable
}

func (*InlineTableContext) IsInlineTableContext() {}

func NewInlineTableContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *InlineTableContext {
	var p = new(InlineTableContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_inlineTable

	return p
}

func (s *InlineTableContext) GetParser() antlr.Parser { return s.parser }

func (s *InlineTableContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserLBRACE, 0)
}

func (s *InlineTableContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserRBRACE, 0)
}

func (s *InlineTableContext) AllField() []IFieldContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IFieldContext); ok {
			len++
		}
	}

	tst := make([]IFieldContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IFieldContext); ok {
			tst[i] = t.(IFieldContext)
			i++
		}
	}

	return tst
}

func (s *InlineTableContext) Field(i int) IFieldContext {
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFieldContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFieldContext)
}

func (s *InlineTableContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(MosDSLParserCOMMA)
}

func (s *InlineTableContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(MosDSLParserCOMMA, i)
}

func (s *InlineTableContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InlineTableContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *InlineTableContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterInlineTable(s)
	}
}

func (s *InlineTableContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitInlineTable(s)
	}
}




func (p *MosDSLParser) InlineTable() (localctx IInlineTableContext) {
	localctx = NewInlineTableContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, MosDSLParserRULE_inlineTable)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(108)
		p.Match(MosDSLParserLBRACE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	p.SetState(120)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if ((int64(_la) & ^0x3f) == 0 && ((int64(1) << _la) & 67117040) != 0) {
		{
			p.SetState(109)
			p.Field()
		}
		p.SetState(114)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 7, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(110)
					p.Match(MosDSLParserCOMMA)
					if p.HasError() {
							// Recognition error - abort rule
							goto errorExit
					}
				}
				{
					p.SetState(111)
					p.Field()
				}


			}
			p.SetState(116)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
		    	goto errorExit
		    }
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 7, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		p.SetState(118)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)


		if _la == MosDSLParserCOMMA {
			{
				p.SetState(117)
				p.Match(MosDSLParserCOMMA)
				if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
				}
			}

		}

	}
	{
		p.SetState(122)
		p.Match(MosDSLParserRBRACE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}



errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// INestedBlockContext is an interface to support dynamic dispatch.
type INestedBlockContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	BlockName() IBlockNameContext
	Block() IBlockContext
	STRING() antlr.TerminalNode

	// IsNestedBlockContext differentiates from other interfaces.
	IsNestedBlockContext()
}

type NestedBlockContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyNestedBlockContext() *NestedBlockContext {
	var p = new(NestedBlockContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_nestedBlock
	return p
}

func InitEmptyNestedBlockContext(p *NestedBlockContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_nestedBlock
}

func (*NestedBlockContext) IsNestedBlockContext() {}

func NewNestedBlockContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *NestedBlockContext {
	var p = new(NestedBlockContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_nestedBlock

	return p
}

func (s *NestedBlockContext) GetParser() antlr.Parser { return s.parser }

func (s *NestedBlockContext) BlockName() IBlockNameContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBlockNameContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBlockNameContext)
}

func (s *NestedBlockContext) Block() IBlockContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBlockContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBlockContext)
}

func (s *NestedBlockContext) STRING() antlr.TerminalNode {
	return s.GetToken(MosDSLParserSTRING, 0)
}

func (s *NestedBlockContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NestedBlockContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *NestedBlockContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterNestedBlock(s)
	}
}

func (s *NestedBlockContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitNestedBlock(s)
	}
}




func (p *MosDSLParser) NestedBlock() (localctx INestedBlockContext) {
	localctx = NewNestedBlockContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, MosDSLParserRULE_nestedBlock)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(124)
		p.BlockName()
	}
	p.SetState(126)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if _la == MosDSLParserSTRING {
		{
			p.SetState(125)
			p.Match(MosDSLParserSTRING)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}

	}
	{
		p.SetState(128)
		p.Block()
	}



errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// IBlockNameContext is an interface to support dynamic dispatch.
type IBlockNameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENT() antlr.TerminalNode
	FEATURE() antlr.TerminalNode
	BACKGROUND() antlr.TerminalNode
	SCENARIO() antlr.TerminalNode
	GIVEN() antlr.TerminalNode
	WHEN() antlr.TerminalNode
	THEN() antlr.TerminalNode
	GROUP() antlr.TerminalNode
	SPEC() antlr.TerminalNode
	INCLUDE() antlr.TerminalNode

	// IsBlockNameContext differentiates from other interfaces.
	IsBlockNameContext()
}

type BlockNameContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBlockNameContext() *BlockNameContext {
	var p = new(BlockNameContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_blockName
	return p
}

func InitEmptyBlockNameContext(p *BlockNameContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_blockName
}

func (*BlockNameContext) IsBlockNameContext() {}

func NewBlockNameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *BlockNameContext {
	var p = new(BlockNameContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_blockName

	return p
}

func (s *BlockNameContext) GetParser() antlr.Parser { return s.parser }

func (s *BlockNameContext) IDENT() antlr.TerminalNode {
	return s.GetToken(MosDSLParserIDENT, 0)
}

func (s *BlockNameContext) FEATURE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserFEATURE, 0)
}

func (s *BlockNameContext) BACKGROUND() antlr.TerminalNode {
	return s.GetToken(MosDSLParserBACKGROUND, 0)
}

func (s *BlockNameContext) SCENARIO() antlr.TerminalNode {
	return s.GetToken(MosDSLParserSCENARIO, 0)
}

func (s *BlockNameContext) GIVEN() antlr.TerminalNode {
	return s.GetToken(MosDSLParserGIVEN, 0)
}

func (s *BlockNameContext) WHEN() antlr.TerminalNode {
	return s.GetToken(MosDSLParserWHEN, 0)
}

func (s *BlockNameContext) THEN() antlr.TerminalNode {
	return s.GetToken(MosDSLParserTHEN, 0)
}

func (s *BlockNameContext) GROUP() antlr.TerminalNode {
	return s.GetToken(MosDSLParserGROUP, 0)
}

func (s *BlockNameContext) SPEC() antlr.TerminalNode {
	return s.GetToken(MosDSLParserSPEC, 0)
}

func (s *BlockNameContext) INCLUDE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserINCLUDE, 0)
}

func (s *BlockNameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BlockNameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *BlockNameContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterBlockName(s)
	}
}

func (s *BlockNameContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitBlockName(s)
	}
}




func (p *MosDSLParser) BlockName() (localctx IBlockNameContext) {
	localctx = NewBlockNameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, MosDSLParserRULE_blockName)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(130)
		_la = p.GetTokenStream().LA(1)

		if !(((int64(_la) & ^0x3f) == 0 && ((int64(1) << _la) & 67117040) != 0)) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}



errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// ISpecBlockContext is an interface to support dynamic dispatch.
type ISpecBlockContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SPEC() antlr.TerminalNode
	LBRACE() antlr.TerminalNode
	RBRACE() antlr.TerminalNode
	AllIncludeDir() []IIncludeDirContext
	IncludeDir(i int) IIncludeDirContext
	AllFeatureBlock() []IFeatureBlockContext
	FeatureBlock(i int) IFeatureBlockContext

	// IsSpecBlockContext differentiates from other interfaces.
	IsSpecBlockContext()
}

type SpecBlockContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySpecBlockContext() *SpecBlockContext {
	var p = new(SpecBlockContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_specBlock
	return p
}

func InitEmptySpecBlockContext(p *SpecBlockContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_specBlock
}

func (*SpecBlockContext) IsSpecBlockContext() {}

func NewSpecBlockContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SpecBlockContext {
	var p = new(SpecBlockContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_specBlock

	return p
}

func (s *SpecBlockContext) GetParser() antlr.Parser { return s.parser }

func (s *SpecBlockContext) SPEC() antlr.TerminalNode {
	return s.GetToken(MosDSLParserSPEC, 0)
}

func (s *SpecBlockContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserLBRACE, 0)
}

func (s *SpecBlockContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserRBRACE, 0)
}

func (s *SpecBlockContext) AllIncludeDir() []IIncludeDirContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IIncludeDirContext); ok {
			len++
		}
	}

	tst := make([]IIncludeDirContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IIncludeDirContext); ok {
			tst[i] = t.(IIncludeDirContext)
			i++
		}
	}

	return tst
}

func (s *SpecBlockContext) IncludeDir(i int) IIncludeDirContext {
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIncludeDirContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIncludeDirContext)
}

func (s *SpecBlockContext) AllFeatureBlock() []IFeatureBlockContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IFeatureBlockContext); ok {
			len++
		}
	}

	tst := make([]IFeatureBlockContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IFeatureBlockContext); ok {
			tst[i] = t.(IFeatureBlockContext)
			i++
		}
	}

	return tst
}

func (s *SpecBlockContext) FeatureBlock(i int) IFeatureBlockContext {
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFeatureBlockContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFeatureBlockContext)
}

func (s *SpecBlockContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SpecBlockContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *SpecBlockContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterSpecBlock(s)
	}
}

func (s *SpecBlockContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitSpecBlock(s)
	}
}




func (p *MosDSLParser) SpecBlock() (localctx ISpecBlockContext) {
	localctx = NewSpecBlockContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 26, MosDSLParserRULE_specBlock)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(132)
		p.Match(MosDSLParserSPEC)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(133)
		p.Match(MosDSLParserLBRACE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	p.SetState(138)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	for _la == MosDSLParserFEATURE || _la == MosDSLParserINCLUDE {
		p.SetState(136)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case MosDSLParserINCLUDE:
			{
				p.SetState(134)
				p.IncludeDir()
			}


		case MosDSLParserFEATURE:
			{
				p.SetState(135)
				p.FeatureBlock()
			}



		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

		p.SetState(140)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
	    	goto errorExit
	    }
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(141)
		p.Match(MosDSLParserRBRACE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}



errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// IIncludeDirContext is an interface to support dynamic dispatch.
type IIncludeDirContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	INCLUDE() antlr.TerminalNode
	STRING() antlr.TerminalNode

	// IsIncludeDirContext differentiates from other interfaces.
	IsIncludeDirContext()
}

type IncludeDirContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIncludeDirContext() *IncludeDirContext {
	var p = new(IncludeDirContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_includeDir
	return p
}

func InitEmptyIncludeDirContext(p *IncludeDirContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_includeDir
}

func (*IncludeDirContext) IsIncludeDirContext() {}

func NewIncludeDirContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *IncludeDirContext {
	var p = new(IncludeDirContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_includeDir

	return p
}

func (s *IncludeDirContext) GetParser() antlr.Parser { return s.parser }

func (s *IncludeDirContext) INCLUDE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserINCLUDE, 0)
}

func (s *IncludeDirContext) STRING() antlr.TerminalNode {
	return s.GetToken(MosDSLParserSTRING, 0)
}

func (s *IncludeDirContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IncludeDirContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *IncludeDirContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterIncludeDir(s)
	}
}

func (s *IncludeDirContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitIncludeDir(s)
	}
}




func (p *MosDSLParser) IncludeDir() (localctx IIncludeDirContext) {
	localctx = NewIncludeDirContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 28, MosDSLParserRULE_includeDir)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(143)
		p.Match(MosDSLParserINCLUDE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(144)
		p.Match(MosDSLParserSTRING)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}



errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// IFeatureBlockContext is an interface to support dynamic dispatch.
type IFeatureBlockContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	FEATURE() antlr.TerminalNode
	LBRACE() antlr.TerminalNode
	RBRACE() antlr.TerminalNode
	STRING() antlr.TerminalNode
	BackgroundBlock() IBackgroundBlockContext
	AllGroupBlock() []IGroupBlockContext
	GroupBlock(i int) IGroupBlockContext
	AllScenarioBlock() []IScenarioBlockContext
	ScenarioBlock(i int) IScenarioBlockContext

	// IsFeatureBlockContext differentiates from other interfaces.
	IsFeatureBlockContext()
}

type FeatureBlockContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFeatureBlockContext() *FeatureBlockContext {
	var p = new(FeatureBlockContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_featureBlock
	return p
}

func InitEmptyFeatureBlockContext(p *FeatureBlockContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_featureBlock
}

func (*FeatureBlockContext) IsFeatureBlockContext() {}

func NewFeatureBlockContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FeatureBlockContext {
	var p = new(FeatureBlockContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_featureBlock

	return p
}

func (s *FeatureBlockContext) GetParser() antlr.Parser { return s.parser }

func (s *FeatureBlockContext) FEATURE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserFEATURE, 0)
}

func (s *FeatureBlockContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserLBRACE, 0)
}

func (s *FeatureBlockContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserRBRACE, 0)
}

func (s *FeatureBlockContext) STRING() antlr.TerminalNode {
	return s.GetToken(MosDSLParserSTRING, 0)
}

func (s *FeatureBlockContext) BackgroundBlock() IBackgroundBlockContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBackgroundBlockContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBackgroundBlockContext)
}

func (s *FeatureBlockContext) AllGroupBlock() []IGroupBlockContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IGroupBlockContext); ok {
			len++
		}
	}

	tst := make([]IGroupBlockContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IGroupBlockContext); ok {
			tst[i] = t.(IGroupBlockContext)
			i++
		}
	}

	return tst
}

func (s *FeatureBlockContext) GroupBlock(i int) IGroupBlockContext {
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IGroupBlockContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IGroupBlockContext)
}

func (s *FeatureBlockContext) AllScenarioBlock() []IScenarioBlockContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IScenarioBlockContext); ok {
			len++
		}
	}

	tst := make([]IScenarioBlockContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IScenarioBlockContext); ok {
			tst[i] = t.(IScenarioBlockContext)
			i++
		}
	}

	return tst
}

func (s *FeatureBlockContext) ScenarioBlock(i int) IScenarioBlockContext {
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IScenarioBlockContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IScenarioBlockContext)
}

func (s *FeatureBlockContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FeatureBlockContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *FeatureBlockContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterFeatureBlock(s)
	}
}

func (s *FeatureBlockContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitFeatureBlock(s)
	}
}




func (p *MosDSLParser) FeatureBlock() (localctx IFeatureBlockContext) {
	localctx = NewFeatureBlockContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 30, MosDSLParserRULE_featureBlock)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(146)
		p.Match(MosDSLParserFEATURE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	p.SetState(148)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if _la == MosDSLParserSTRING {
		{
			p.SetState(147)
			p.Match(MosDSLParserSTRING)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}

	}
	{
		p.SetState(150)
		p.Match(MosDSLParserLBRACE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	p.SetState(152)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if _la == MosDSLParserBACKGROUND {
		{
			p.SetState(151)
			p.BackgroundBlock()
		}

	}
	p.SetState(158)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	for _la == MosDSLParserSCENARIO || _la == MosDSLParserGROUP {
		p.SetState(156)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case MosDSLParserGROUP:
			{
				p.SetState(154)
				p.GroupBlock()
			}


		case MosDSLParserSCENARIO:
			{
				p.SetState(155)
				p.ScenarioBlock()
			}



		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

		p.SetState(160)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
	    	goto errorExit
	    }
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(161)
		p.Match(MosDSLParserRBRACE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}



errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// IBackgroundBlockContext is an interface to support dynamic dispatch.
type IBackgroundBlockContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	BACKGROUND() antlr.TerminalNode
	LBRACE() antlr.TerminalNode
	GIVEN_OPEN() antlr.TerminalNode
	STEP_RBRACE() antlr.TerminalNode
	RBRACE() antlr.TerminalNode
	AllStepLine() []IStepLineContext
	StepLine(i int) IStepLineContext

	// IsBackgroundBlockContext differentiates from other interfaces.
	IsBackgroundBlockContext()
}

type BackgroundBlockContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBackgroundBlockContext() *BackgroundBlockContext {
	var p = new(BackgroundBlockContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_backgroundBlock
	return p
}

func InitEmptyBackgroundBlockContext(p *BackgroundBlockContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_backgroundBlock
}

func (*BackgroundBlockContext) IsBackgroundBlockContext() {}

func NewBackgroundBlockContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *BackgroundBlockContext {
	var p = new(BackgroundBlockContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_backgroundBlock

	return p
}

func (s *BackgroundBlockContext) GetParser() antlr.Parser { return s.parser }

func (s *BackgroundBlockContext) BACKGROUND() antlr.TerminalNode {
	return s.GetToken(MosDSLParserBACKGROUND, 0)
}

func (s *BackgroundBlockContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserLBRACE, 0)
}

func (s *BackgroundBlockContext) GIVEN_OPEN() antlr.TerminalNode {
	return s.GetToken(MosDSLParserGIVEN_OPEN, 0)
}

func (s *BackgroundBlockContext) STEP_RBRACE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserSTEP_RBRACE, 0)
}

func (s *BackgroundBlockContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserRBRACE, 0)
}

func (s *BackgroundBlockContext) AllStepLine() []IStepLineContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IStepLineContext); ok {
			len++
		}
	}

	tst := make([]IStepLineContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IStepLineContext); ok {
			tst[i] = t.(IStepLineContext)
			i++
		}
	}

	return tst
}

func (s *BackgroundBlockContext) StepLine(i int) IStepLineContext {
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStepLineContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStepLineContext)
}

func (s *BackgroundBlockContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BackgroundBlockContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *BackgroundBlockContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterBackgroundBlock(s)
	}
}

func (s *BackgroundBlockContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitBackgroundBlock(s)
	}
}




func (p *MosDSLParser) BackgroundBlock() (localctx IBackgroundBlockContext) {
	localctx = NewBackgroundBlockContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 32, MosDSLParserRULE_backgroundBlock)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(163)
		p.Match(MosDSLParserBACKGROUND)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(164)
		p.Match(MosDSLParserLBRACE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(165)
		p.Match(MosDSLParserGIVEN_OPEN)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	p.SetState(169)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	for _la == MosDSLParserSTEP_TEXT {
		{
			p.SetState(166)
			p.StepLine()
		}


		p.SetState(171)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
	    	goto errorExit
	    }
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(172)
		p.Match(MosDSLParserSTEP_RBRACE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(173)
		p.Match(MosDSLParserRBRACE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}



errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// IGroupBlockContext is an interface to support dynamic dispatch.
type IGroupBlockContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	GROUP() antlr.TerminalNode
	LBRACE() antlr.TerminalNode
	RBRACE() antlr.TerminalNode
	STRING() antlr.TerminalNode
	AllScenarioBlock() []IScenarioBlockContext
	ScenarioBlock(i int) IScenarioBlockContext

	// IsGroupBlockContext differentiates from other interfaces.
	IsGroupBlockContext()
}

type GroupBlockContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyGroupBlockContext() *GroupBlockContext {
	var p = new(GroupBlockContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_groupBlock
	return p
}

func InitEmptyGroupBlockContext(p *GroupBlockContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_groupBlock
}

func (*GroupBlockContext) IsGroupBlockContext() {}

func NewGroupBlockContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *GroupBlockContext {
	var p = new(GroupBlockContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_groupBlock

	return p
}

func (s *GroupBlockContext) GetParser() antlr.Parser { return s.parser }

func (s *GroupBlockContext) GROUP() antlr.TerminalNode {
	return s.GetToken(MosDSLParserGROUP, 0)
}

func (s *GroupBlockContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserLBRACE, 0)
}

func (s *GroupBlockContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserRBRACE, 0)
}

func (s *GroupBlockContext) STRING() antlr.TerminalNode {
	return s.GetToken(MosDSLParserSTRING, 0)
}

func (s *GroupBlockContext) AllScenarioBlock() []IScenarioBlockContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IScenarioBlockContext); ok {
			len++
		}
	}

	tst := make([]IScenarioBlockContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IScenarioBlockContext); ok {
			tst[i] = t.(IScenarioBlockContext)
			i++
		}
	}

	return tst
}

func (s *GroupBlockContext) ScenarioBlock(i int) IScenarioBlockContext {
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IScenarioBlockContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IScenarioBlockContext)
}

func (s *GroupBlockContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *GroupBlockContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *GroupBlockContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterGroupBlock(s)
	}
}

func (s *GroupBlockContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitGroupBlock(s)
	}
}




func (p *MosDSLParser) GroupBlock() (localctx IGroupBlockContext) {
	localctx = NewGroupBlockContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 34, MosDSLParserRULE_groupBlock)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(175)
		p.Match(MosDSLParserGROUP)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	p.SetState(177)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if _la == MosDSLParserSTRING {
		{
			p.SetState(176)
			p.Match(MosDSLParserSTRING)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}

	}
	{
		p.SetState(179)
		p.Match(MosDSLParserLBRACE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	p.SetState(183)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	for _la == MosDSLParserSCENARIO {
		{
			p.SetState(180)
			p.ScenarioBlock()
		}


		p.SetState(185)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
	    	goto errorExit
	    }
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(186)
		p.Match(MosDSLParserRBRACE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}



errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// IScenarioBlockContext is an interface to support dynamic dispatch.
type IScenarioBlockContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SCENARIO() antlr.TerminalNode
	LBRACE() antlr.TerminalNode
	ScenarioContent() IScenarioContentContext
	RBRACE() antlr.TerminalNode
	STRING() antlr.TerminalNode

	// IsScenarioBlockContext differentiates from other interfaces.
	IsScenarioBlockContext()
}

type ScenarioBlockContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyScenarioBlockContext() *ScenarioBlockContext {
	var p = new(ScenarioBlockContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_scenarioBlock
	return p
}

func InitEmptyScenarioBlockContext(p *ScenarioBlockContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_scenarioBlock
}

func (*ScenarioBlockContext) IsScenarioBlockContext() {}

func NewScenarioBlockContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ScenarioBlockContext {
	var p = new(ScenarioBlockContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_scenarioBlock

	return p
}

func (s *ScenarioBlockContext) GetParser() antlr.Parser { return s.parser }

func (s *ScenarioBlockContext) SCENARIO() antlr.TerminalNode {
	return s.GetToken(MosDSLParserSCENARIO, 0)
}

func (s *ScenarioBlockContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserLBRACE, 0)
}

func (s *ScenarioBlockContext) ScenarioContent() IScenarioContentContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IScenarioContentContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IScenarioContentContext)
}

func (s *ScenarioBlockContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserRBRACE, 0)
}

func (s *ScenarioBlockContext) STRING() antlr.TerminalNode {
	return s.GetToken(MosDSLParserSTRING, 0)
}

func (s *ScenarioBlockContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ScenarioBlockContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *ScenarioBlockContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterScenarioBlock(s)
	}
}

func (s *ScenarioBlockContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitScenarioBlock(s)
	}
}




func (p *MosDSLParser) ScenarioBlock() (localctx IScenarioBlockContext) {
	localctx = NewScenarioBlockContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 36, MosDSLParserRULE_scenarioBlock)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(188)
		p.Match(MosDSLParserSCENARIO)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	p.SetState(190)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if _la == MosDSLParserSTRING {
		{
			p.SetState(189)
			p.Match(MosDSLParserSTRING)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}

	}
	{
		p.SetState(192)
		p.Match(MosDSLParserLBRACE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(193)
		p.ScenarioContent()
	}
	{
		p.SetState(194)
		p.Match(MosDSLParserRBRACE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}



errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// IScenarioContentContext is an interface to support dynamic dispatch.
type IScenarioContentContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllField() []IFieldContext
	Field(i int) IFieldContext
	GivenBlock() IGivenBlockContext
	WhenBlock() IWhenBlockContext
	ThenBlock() IThenBlockContext

	// IsScenarioContentContext differentiates from other interfaces.
	IsScenarioContentContext()
}

type ScenarioContentContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyScenarioContentContext() *ScenarioContentContext {
	var p = new(ScenarioContentContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_scenarioContent
	return p
}

func InitEmptyScenarioContentContext(p *ScenarioContentContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_scenarioContent
}

func (*ScenarioContentContext) IsScenarioContentContext() {}

func NewScenarioContentContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ScenarioContentContext {
	var p = new(ScenarioContentContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_scenarioContent

	return p
}

func (s *ScenarioContentContext) GetParser() antlr.Parser { return s.parser }

func (s *ScenarioContentContext) AllField() []IFieldContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IFieldContext); ok {
			len++
		}
	}

	tst := make([]IFieldContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IFieldContext); ok {
			tst[i] = t.(IFieldContext)
			i++
		}
	}

	return tst
}

func (s *ScenarioContentContext) Field(i int) IFieldContext {
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFieldContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFieldContext)
}

func (s *ScenarioContentContext) GivenBlock() IGivenBlockContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IGivenBlockContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IGivenBlockContext)
}

func (s *ScenarioContentContext) WhenBlock() IWhenBlockContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IWhenBlockContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IWhenBlockContext)
}

func (s *ScenarioContentContext) ThenBlock() IThenBlockContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IThenBlockContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IThenBlockContext)
}

func (s *ScenarioContentContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ScenarioContentContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *ScenarioContentContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterScenarioContent(s)
	}
}

func (s *ScenarioContentContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitScenarioContent(s)
	}
}




func (p *MosDSLParser) ScenarioContent() (localctx IScenarioContentContext) {
	localctx = NewScenarioContentContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 38, MosDSLParserRULE_scenarioContent)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(199)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	for ((int64(_la) & ^0x3f) == 0 && ((int64(1) << _la) & 67117040) != 0) {
		{
			p.SetState(196)
			p.Field()
		}


		p.SetState(201)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
	    	goto errorExit
	    }
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(203)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if _la == MosDSLParserGIVEN_OPEN {
		{
			p.SetState(202)
			p.GivenBlock()
		}

	}
	p.SetState(206)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if _la == MosDSLParserWHEN_OPEN {
		{
			p.SetState(205)
			p.WhenBlock()
		}

	}
	p.SetState(209)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if _la == MosDSLParserTHEN_OPEN {
		{
			p.SetState(208)
			p.ThenBlock()
		}

	}



errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// IGivenBlockContext is an interface to support dynamic dispatch.
type IGivenBlockContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	GIVEN_OPEN() antlr.TerminalNode
	STEP_RBRACE() antlr.TerminalNode
	AllStepLine() []IStepLineContext
	StepLine(i int) IStepLineContext

	// IsGivenBlockContext differentiates from other interfaces.
	IsGivenBlockContext()
}

type GivenBlockContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyGivenBlockContext() *GivenBlockContext {
	var p = new(GivenBlockContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_givenBlock
	return p
}

func InitEmptyGivenBlockContext(p *GivenBlockContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_givenBlock
}

func (*GivenBlockContext) IsGivenBlockContext() {}

func NewGivenBlockContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *GivenBlockContext {
	var p = new(GivenBlockContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_givenBlock

	return p
}

func (s *GivenBlockContext) GetParser() antlr.Parser { return s.parser }

func (s *GivenBlockContext) GIVEN_OPEN() antlr.TerminalNode {
	return s.GetToken(MosDSLParserGIVEN_OPEN, 0)
}

func (s *GivenBlockContext) STEP_RBRACE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserSTEP_RBRACE, 0)
}

func (s *GivenBlockContext) AllStepLine() []IStepLineContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IStepLineContext); ok {
			len++
		}
	}

	tst := make([]IStepLineContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IStepLineContext); ok {
			tst[i] = t.(IStepLineContext)
			i++
		}
	}

	return tst
}

func (s *GivenBlockContext) StepLine(i int) IStepLineContext {
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStepLineContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStepLineContext)
}

func (s *GivenBlockContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *GivenBlockContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *GivenBlockContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterGivenBlock(s)
	}
}

func (s *GivenBlockContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitGivenBlock(s)
	}
}




func (p *MosDSLParser) GivenBlock() (localctx IGivenBlockContext) {
	localctx = NewGivenBlockContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 40, MosDSLParserRULE_givenBlock)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(211)
		p.Match(MosDSLParserGIVEN_OPEN)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	p.SetState(215)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	for _la == MosDSLParserSTEP_TEXT {
		{
			p.SetState(212)
			p.StepLine()
		}


		p.SetState(217)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
	    	goto errorExit
	    }
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(218)
		p.Match(MosDSLParserSTEP_RBRACE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}



errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// IWhenBlockContext is an interface to support dynamic dispatch.
type IWhenBlockContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	WHEN_OPEN() antlr.TerminalNode
	STEP_RBRACE() antlr.TerminalNode
	AllStepLine() []IStepLineContext
	StepLine(i int) IStepLineContext

	// IsWhenBlockContext differentiates from other interfaces.
	IsWhenBlockContext()
}

type WhenBlockContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyWhenBlockContext() *WhenBlockContext {
	var p = new(WhenBlockContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_whenBlock
	return p
}

func InitEmptyWhenBlockContext(p *WhenBlockContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_whenBlock
}

func (*WhenBlockContext) IsWhenBlockContext() {}

func NewWhenBlockContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *WhenBlockContext {
	var p = new(WhenBlockContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_whenBlock

	return p
}

func (s *WhenBlockContext) GetParser() antlr.Parser { return s.parser }

func (s *WhenBlockContext) WHEN_OPEN() antlr.TerminalNode {
	return s.GetToken(MosDSLParserWHEN_OPEN, 0)
}

func (s *WhenBlockContext) STEP_RBRACE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserSTEP_RBRACE, 0)
}

func (s *WhenBlockContext) AllStepLine() []IStepLineContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IStepLineContext); ok {
			len++
		}
	}

	tst := make([]IStepLineContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IStepLineContext); ok {
			tst[i] = t.(IStepLineContext)
			i++
		}
	}

	return tst
}

func (s *WhenBlockContext) StepLine(i int) IStepLineContext {
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStepLineContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStepLineContext)
}

func (s *WhenBlockContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *WhenBlockContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *WhenBlockContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterWhenBlock(s)
	}
}

func (s *WhenBlockContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitWhenBlock(s)
	}
}




func (p *MosDSLParser) WhenBlock() (localctx IWhenBlockContext) {
	localctx = NewWhenBlockContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 42, MosDSLParserRULE_whenBlock)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(220)
		p.Match(MosDSLParserWHEN_OPEN)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	p.SetState(224)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	for _la == MosDSLParserSTEP_TEXT {
		{
			p.SetState(221)
			p.StepLine()
		}


		p.SetState(226)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
	    	goto errorExit
	    }
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(227)
		p.Match(MosDSLParserSTEP_RBRACE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}



errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// IThenBlockContext is an interface to support dynamic dispatch.
type IThenBlockContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	THEN_OPEN() antlr.TerminalNode
	STEP_RBRACE() antlr.TerminalNode
	AllStepLine() []IStepLineContext
	StepLine(i int) IStepLineContext

	// IsThenBlockContext differentiates from other interfaces.
	IsThenBlockContext()
}

type ThenBlockContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyThenBlockContext() *ThenBlockContext {
	var p = new(ThenBlockContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_thenBlock
	return p
}

func InitEmptyThenBlockContext(p *ThenBlockContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_thenBlock
}

func (*ThenBlockContext) IsThenBlockContext() {}

func NewThenBlockContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ThenBlockContext {
	var p = new(ThenBlockContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_thenBlock

	return p
}

func (s *ThenBlockContext) GetParser() antlr.Parser { return s.parser }

func (s *ThenBlockContext) THEN_OPEN() antlr.TerminalNode {
	return s.GetToken(MosDSLParserTHEN_OPEN, 0)
}

func (s *ThenBlockContext) STEP_RBRACE() antlr.TerminalNode {
	return s.GetToken(MosDSLParserSTEP_RBRACE, 0)
}

func (s *ThenBlockContext) AllStepLine() []IStepLineContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IStepLineContext); ok {
			len++
		}
	}

	tst := make([]IStepLineContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IStepLineContext); ok {
			tst[i] = t.(IStepLineContext)
			i++
		}
	}

	return tst
}

func (s *ThenBlockContext) StepLine(i int) IStepLineContext {
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStepLineContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStepLineContext)
}

func (s *ThenBlockContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ThenBlockContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *ThenBlockContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterThenBlock(s)
	}
}

func (s *ThenBlockContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitThenBlock(s)
	}
}




func (p *MosDSLParser) ThenBlock() (localctx IThenBlockContext) {
	localctx = NewThenBlockContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 44, MosDSLParserRULE_thenBlock)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(229)
		p.Match(MosDSLParserTHEN_OPEN)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	p.SetState(233)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	for _la == MosDSLParserSTEP_TEXT {
		{
			p.SetState(230)
			p.StepLine()
		}


		p.SetState(235)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
	    	goto errorExit
	    }
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(236)
		p.Match(MosDSLParserSTEP_RBRACE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}



errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


// IStepLineContext is an interface to support dynamic dispatch.
type IStepLineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STEP_TEXT() antlr.TerminalNode

	// IsStepLineContext differentiates from other interfaces.
	IsStepLineContext()
}

type StepLineContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyStepLineContext() *StepLineContext {
	var p = new(StepLineContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_stepLine
	return p
}

func InitEmptyStepLineContext(p *StepLineContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MosDSLParserRULE_stepLine
}

func (*StepLineContext) IsStepLineContext() {}

func NewStepLineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *StepLineContext {
	var p = new(StepLineContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MosDSLParserRULE_stepLine

	return p
}

func (s *StepLineContext) GetParser() antlr.Parser { return s.parser }

func (s *StepLineContext) STEP_TEXT() antlr.TerminalNode {
	return s.GetToken(MosDSLParserSTEP_TEXT, 0)
}

func (s *StepLineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StepLineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *StepLineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.EnterStepLine(s)
	}
}

func (s *StepLineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MosDSLParserListener); ok {
		listenerT.ExitStepLine(s)
	}
}




func (p *MosDSLParser) StepLine() (localctx IStepLineContext) {
	localctx = NewStepLineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 46, MosDSLParserRULE_stepLine)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(238)
		p.Match(MosDSLParserSTEP_TEXT)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}



errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}


