; Keywords
"feature" @keyword
"background" @keyword
"scenario" @keyword
"given" @keyword
"when" @keyword
"then" @keyword
"group" @keyword
"spec" @keyword
"include" @keyword.import
"true" @constant.builtin
"false" @constant.builtin

; Strings
(string) @string
(triple_string) @string

; Numbers
(integer) @number
(float) @number.float

; Dates
(datetime) @string.special

; Comments
(comment) @comment

; Fields
(field (key (identifier) @property))

; Artifact type (first identifier in artifact)
(artifact (artifact_type (identifier) @type))

; Block names
(nested_block (block_name (identifier) @tag))
(nested_block (string) @string.special)

; Feature/scenario names
(feature_block (string) @string.special)
(scenario_block (string) @string.special)

; Step text (free text inside given/when/then)
(step_line) @string.special.symbol

; Operators
"=" @operator

; Punctuation
"{" @punctuation.bracket
"}" @punctuation.bracket
"[" @punctuation.bracket
"]" @punctuation.bracket
"," @punctuation.delimiter
