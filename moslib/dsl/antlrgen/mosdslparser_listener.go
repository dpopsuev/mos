
package antlrgen // MosDSLParser
import "github.com/antlr4-go/antlr/v4"


// MosDSLParserListener is a complete listener for a parse tree produced by MosDSLParser.
type MosDSLParserListener interface {
	antlr.ParseTreeListener

	// EnterFile is called when entering the file production.
	EnterFile(c *FileContext)

	// EnterArtifact is called when entering the artifact production.
	EnterArtifact(c *ArtifactContext)

	// EnterArtifactType is called when entering the artifactType production.
	EnterArtifactType(c *ArtifactTypeContext)

	// EnterBlock is called when entering the block production.
	EnterBlock(c *BlockContext)

	// EnterBlockItem is called when entering the blockItem production.
	EnterBlockItem(c *BlockItemContext)

	// EnterField is called when entering the field production.
	EnterField(c *FieldContext)

	// EnterKey is called when entering the key production.
	EnterKey(c *KeyContext)

	// EnterValue is called when entering the value production.
	EnterValue(c *ValueContext)

	// EnterBoolean is called when entering the boolean production.
	EnterBoolean(c *BooleanContext)

	// EnterList is called when entering the list production.
	EnterList(c *ListContext)

	// EnterInlineTable is called when entering the inlineTable production.
	EnterInlineTable(c *InlineTableContext)

	// EnterNestedBlock is called when entering the nestedBlock production.
	EnterNestedBlock(c *NestedBlockContext)

	// EnterBlockName is called when entering the blockName production.
	EnterBlockName(c *BlockNameContext)

	// EnterSpecBlock is called when entering the specBlock production.
	EnterSpecBlock(c *SpecBlockContext)

	// EnterIncludeDir is called when entering the includeDir production.
	EnterIncludeDir(c *IncludeDirContext)

	// EnterFeatureBlock is called when entering the featureBlock production.
	EnterFeatureBlock(c *FeatureBlockContext)

	// EnterBackgroundBlock is called when entering the backgroundBlock production.
	EnterBackgroundBlock(c *BackgroundBlockContext)

	// EnterGroupBlock is called when entering the groupBlock production.
	EnterGroupBlock(c *GroupBlockContext)

	// EnterScenarioBlock is called when entering the scenarioBlock production.
	EnterScenarioBlock(c *ScenarioBlockContext)

	// EnterScenarioContent is called when entering the scenarioContent production.
	EnterScenarioContent(c *ScenarioContentContext)

	// EnterGivenBlock is called when entering the givenBlock production.
	EnterGivenBlock(c *GivenBlockContext)

	// EnterWhenBlock is called when entering the whenBlock production.
	EnterWhenBlock(c *WhenBlockContext)

	// EnterThenBlock is called when entering the thenBlock production.
	EnterThenBlock(c *ThenBlockContext)

	// EnterStepLine is called when entering the stepLine production.
	EnterStepLine(c *StepLineContext)

	// ExitFile is called when exiting the file production.
	ExitFile(c *FileContext)

	// ExitArtifact is called when exiting the artifact production.
	ExitArtifact(c *ArtifactContext)

	// ExitArtifactType is called when exiting the artifactType production.
	ExitArtifactType(c *ArtifactTypeContext)

	// ExitBlock is called when exiting the block production.
	ExitBlock(c *BlockContext)

	// ExitBlockItem is called when exiting the blockItem production.
	ExitBlockItem(c *BlockItemContext)

	// ExitField is called when exiting the field production.
	ExitField(c *FieldContext)

	// ExitKey is called when exiting the key production.
	ExitKey(c *KeyContext)

	// ExitValue is called when exiting the value production.
	ExitValue(c *ValueContext)

	// ExitBoolean is called when exiting the boolean production.
	ExitBoolean(c *BooleanContext)

	// ExitList is called when exiting the list production.
	ExitList(c *ListContext)

	// ExitInlineTable is called when exiting the inlineTable production.
	ExitInlineTable(c *InlineTableContext)

	// ExitNestedBlock is called when exiting the nestedBlock production.
	ExitNestedBlock(c *NestedBlockContext)

	// ExitBlockName is called when exiting the blockName production.
	ExitBlockName(c *BlockNameContext)

	// ExitSpecBlock is called when exiting the specBlock production.
	ExitSpecBlock(c *SpecBlockContext)

	// ExitIncludeDir is called when exiting the includeDir production.
	ExitIncludeDir(c *IncludeDirContext)

	// ExitFeatureBlock is called when exiting the featureBlock production.
	ExitFeatureBlock(c *FeatureBlockContext)

	// ExitBackgroundBlock is called when exiting the backgroundBlock production.
	ExitBackgroundBlock(c *BackgroundBlockContext)

	// ExitGroupBlock is called when exiting the groupBlock production.
	ExitGroupBlock(c *GroupBlockContext)

	// ExitScenarioBlock is called when exiting the scenarioBlock production.
	ExitScenarioBlock(c *ScenarioBlockContext)

	// ExitScenarioContent is called when exiting the scenarioContent production.
	ExitScenarioContent(c *ScenarioContentContext)

	// ExitGivenBlock is called when exiting the givenBlock production.
	ExitGivenBlock(c *GivenBlockContext)

	// ExitWhenBlock is called when exiting the whenBlock production.
	ExitWhenBlock(c *WhenBlockContext)

	// ExitThenBlock is called when exiting the thenBlock production.
	ExitThenBlock(c *ThenBlockContext)

	// ExitStepLine is called when exiting the stepLine production.
	ExitStepLine(c *StepLineContext)
}
