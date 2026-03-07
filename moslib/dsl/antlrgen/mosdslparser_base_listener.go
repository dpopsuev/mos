
package antlrgen // MosDSLParser
import "github.com/antlr4-go/antlr/v4"

// BaseMosDSLParserListener is a complete listener for a parse tree produced by MosDSLParser.
type BaseMosDSLParserListener struct{}

var _ MosDSLParserListener = &BaseMosDSLParserListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseMosDSLParserListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseMosDSLParserListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseMosDSLParserListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseMosDSLParserListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterFile is called when production file is entered.
func (s *BaseMosDSLParserListener) EnterFile(ctx *FileContext) {}

// ExitFile is called when production file is exited.
func (s *BaseMosDSLParserListener) ExitFile(ctx *FileContext) {}

// EnterArtifact is called when production artifact is entered.
func (s *BaseMosDSLParserListener) EnterArtifact(ctx *ArtifactContext) {}

// ExitArtifact is called when production artifact is exited.
func (s *BaseMosDSLParserListener) ExitArtifact(ctx *ArtifactContext) {}

// EnterArtifactType is called when production artifactType is entered.
func (s *BaseMosDSLParserListener) EnterArtifactType(ctx *ArtifactTypeContext) {}

// ExitArtifactType is called when production artifactType is exited.
func (s *BaseMosDSLParserListener) ExitArtifactType(ctx *ArtifactTypeContext) {}

// EnterBlock is called when production block is entered.
func (s *BaseMosDSLParserListener) EnterBlock(ctx *BlockContext) {}

// ExitBlock is called when production block is exited.
func (s *BaseMosDSLParserListener) ExitBlock(ctx *BlockContext) {}

// EnterBlockItem is called when production blockItem is entered.
func (s *BaseMosDSLParserListener) EnterBlockItem(ctx *BlockItemContext) {}

// ExitBlockItem is called when production blockItem is exited.
func (s *BaseMosDSLParserListener) ExitBlockItem(ctx *BlockItemContext) {}

// EnterField is called when production field is entered.
func (s *BaseMosDSLParserListener) EnterField(ctx *FieldContext) {}

// ExitField is called when production field is exited.
func (s *BaseMosDSLParserListener) ExitField(ctx *FieldContext) {}

// EnterKey is called when production key is entered.
func (s *BaseMosDSLParserListener) EnterKey(ctx *KeyContext) {}

// ExitKey is called when production key is exited.
func (s *BaseMosDSLParserListener) ExitKey(ctx *KeyContext) {}

// EnterValue is called when production value is entered.
func (s *BaseMosDSLParserListener) EnterValue(ctx *ValueContext) {}

// ExitValue is called when production value is exited.
func (s *BaseMosDSLParserListener) ExitValue(ctx *ValueContext) {}

// EnterBoolean is called when production boolean is entered.
func (s *BaseMosDSLParserListener) EnterBoolean(ctx *BooleanContext) {}

// ExitBoolean is called when production boolean is exited.
func (s *BaseMosDSLParserListener) ExitBoolean(ctx *BooleanContext) {}

// EnterList is called when production list is entered.
func (s *BaseMosDSLParserListener) EnterList(ctx *ListContext) {}

// ExitList is called when production list is exited.
func (s *BaseMosDSLParserListener) ExitList(ctx *ListContext) {}

// EnterInlineTable is called when production inlineTable is entered.
func (s *BaseMosDSLParserListener) EnterInlineTable(ctx *InlineTableContext) {}

// ExitInlineTable is called when production inlineTable is exited.
func (s *BaseMosDSLParserListener) ExitInlineTable(ctx *InlineTableContext) {}

// EnterNestedBlock is called when production nestedBlock is entered.
func (s *BaseMosDSLParserListener) EnterNestedBlock(ctx *NestedBlockContext) {}

// ExitNestedBlock is called when production nestedBlock is exited.
func (s *BaseMosDSLParserListener) ExitNestedBlock(ctx *NestedBlockContext) {}

// EnterBlockName is called when production blockName is entered.
func (s *BaseMosDSLParserListener) EnterBlockName(ctx *BlockNameContext) {}

// ExitBlockName is called when production blockName is exited.
func (s *BaseMosDSLParserListener) ExitBlockName(ctx *BlockNameContext) {}

// EnterSpecBlock is called when production specBlock is entered.
func (s *BaseMosDSLParserListener) EnterSpecBlock(ctx *SpecBlockContext) {}

// ExitSpecBlock is called when production specBlock is exited.
func (s *BaseMosDSLParserListener) ExitSpecBlock(ctx *SpecBlockContext) {}

// EnterIncludeDir is called when production includeDir is entered.
func (s *BaseMosDSLParserListener) EnterIncludeDir(ctx *IncludeDirContext) {}

// ExitIncludeDir is called when production includeDir is exited.
func (s *BaseMosDSLParserListener) ExitIncludeDir(ctx *IncludeDirContext) {}

// EnterFeatureBlock is called when production featureBlock is entered.
func (s *BaseMosDSLParserListener) EnterFeatureBlock(ctx *FeatureBlockContext) {}

// ExitFeatureBlock is called when production featureBlock is exited.
func (s *BaseMosDSLParserListener) ExitFeatureBlock(ctx *FeatureBlockContext) {}

// EnterBackgroundBlock is called when production backgroundBlock is entered.
func (s *BaseMosDSLParserListener) EnterBackgroundBlock(ctx *BackgroundBlockContext) {}

// ExitBackgroundBlock is called when production backgroundBlock is exited.
func (s *BaseMosDSLParserListener) ExitBackgroundBlock(ctx *BackgroundBlockContext) {}

// EnterGroupBlock is called when production groupBlock is entered.
func (s *BaseMosDSLParserListener) EnterGroupBlock(ctx *GroupBlockContext) {}

// ExitGroupBlock is called when production groupBlock is exited.
func (s *BaseMosDSLParserListener) ExitGroupBlock(ctx *GroupBlockContext) {}

// EnterScenarioBlock is called when production scenarioBlock is entered.
func (s *BaseMosDSLParserListener) EnterScenarioBlock(ctx *ScenarioBlockContext) {}

// ExitScenarioBlock is called when production scenarioBlock is exited.
func (s *BaseMosDSLParserListener) ExitScenarioBlock(ctx *ScenarioBlockContext) {}

// EnterScenarioContent is called when production scenarioContent is entered.
func (s *BaseMosDSLParserListener) EnterScenarioContent(ctx *ScenarioContentContext) {}

// ExitScenarioContent is called when production scenarioContent is exited.
func (s *BaseMosDSLParserListener) ExitScenarioContent(ctx *ScenarioContentContext) {}

// EnterGivenBlock is called when production givenBlock is entered.
func (s *BaseMosDSLParserListener) EnterGivenBlock(ctx *GivenBlockContext) {}

// ExitGivenBlock is called when production givenBlock is exited.
func (s *BaseMosDSLParserListener) ExitGivenBlock(ctx *GivenBlockContext) {}

// EnterWhenBlock is called when production whenBlock is entered.
func (s *BaseMosDSLParserListener) EnterWhenBlock(ctx *WhenBlockContext) {}

// ExitWhenBlock is called when production whenBlock is exited.
func (s *BaseMosDSLParserListener) ExitWhenBlock(ctx *WhenBlockContext) {}

// EnterThenBlock is called when production thenBlock is entered.
func (s *BaseMosDSLParserListener) EnterThenBlock(ctx *ThenBlockContext) {}

// ExitThenBlock is called when production thenBlock is exited.
func (s *BaseMosDSLParserListener) ExitThenBlock(ctx *ThenBlockContext) {}

// EnterStepLine is called when production stepLine is entered.
func (s *BaseMosDSLParserListener) EnterStepLine(ctx *StepLineContext) {}

// ExitStepLine is called when production stepLine is exited.
func (s *BaseMosDSLParserListener) ExitStepLine(ctx *StepLineContext) {}
