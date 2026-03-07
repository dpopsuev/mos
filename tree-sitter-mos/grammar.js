module.exports = grammar({
  name: 'mos',

  extras: $ => [/\s/, $.comment],

  rules: {
    source_file: $ => repeat($.artifact),

    artifact: $ => seq(
      $.artifact_type,
      optional($.string),
      $.block
    ),

    artifact_type: $ => $.identifier,

    block: $ => seq('{', repeat($.block_item), '}'),

    block_item: $ => choice(
      $.field,
      $.feature_block,
      $.spec_block,
      $.nested_block
    ),

    field: $ => seq($.key, '=', $.value),

    key: $ => $.identifier,

    value: $ => choice(
      $.string,
      $.triple_string,
      $.float,
      $.integer,
      $.boolean,
      $.datetime,
      $.list,
      $.inline_table
    ),

    list: $ => seq(
      '[',
      optional(seq(
        $.value,
        repeat(seq(optional(','), $.value)),
        optional(',')
      )),
      ']'
    ),

    inline_table: $ => seq(
      '{',
      optional(seq(
        $.field,
        repeat(seq(optional(','), $.field)),
        optional(',')
      )),
      '}'
    ),

    nested_block: $ => seq(
      $.block_name,
      optional($.string),
      $.block
    ),

    block_name: $ => $.identifier,

    // Spec block: spec { include "..." | feature_block }
    spec_block: $ => seq(
      'spec',
      seq('{', repeat(choice($.include_dir, $.feature_block)), '}')
    ),

    include_dir: $ => seq('include', $.string),

    // Feature block: feature "..."? { background? (group | scenario)* }
    feature_block: $ => seq(
      'feature',
      optional($.string),
      seq(
        '{',
        optional($.background_block),
        repeat(choice($.group_block, $.scenario_block)),
        '}'
      )
    ),

    background_block: $ => seq(
      'background',
      seq('{', $.given_block, '}')
    ),

    group_block: $ => seq(
      'group',
      optional($.string),
      seq('{', repeat($.scenario_block), '}')
    ),

    scenario_block: $ => seq(
      'scenario',
      optional($.string),
      seq('{', optional($.scenario_content), '}')
    ),

    // Must match at least one element (tree-sitter disallows rules matching empty)
    scenario_content: $ => choice(
      seq(repeat1($.field)),
      seq(repeat($.field), $.given_block, optional($.when_block), optional($.then_block)),
      seq(repeat($.field), optional($.given_block), $.when_block, optional($.then_block)),
      seq(repeat($.field), optional($.given_block), optional($.when_block), $.then_block),
      seq(repeat($.field), $.given_block, $.when_block, optional($.then_block)),
      seq(repeat($.field), $.given_block, optional($.when_block), $.then_block),
      seq(repeat($.field), optional($.given_block), $.when_block, $.then_block),
      seq(repeat($.field), $.given_block, $.when_block, $.then_block)
    ),

    // Step blocks: given/when/then { step_line* }
    // Precedence: these must be recognized before nested_block when in scenario context.
    given_block: $ => seq('given', seq('{', repeat($.step_line), '}')),
    when_block: $ => seq('when', seq('{', repeat($.step_line), '}')),
    then_block: $ => seq('then', seq('{', repeat($.step_line), '}')),

    // Step line: free text; excludes lines that are just whitespace + }
    // Match: optional indent, then at least one char that's not } or newline
    step_line: $ => /[ \t]*[^ \t\n}][^\n]*/,

    // Primitives
    string: $ => seq(
      '"',
      repeat(choice(
        /[^"\\\n\r]/,
        seq('\\', choice(/[nrt"\\]/, seq('u', /[0-9a-fA-F]{4}/)))
      )),
      '"'
    ),

    triple_string: $ => token(seq('"""', /[\s\S]*?/, '"""')),

    integer: $ => /-?[0-9]+/,

    float: $ => /-?[0-9]+\.[0-9]+/,

    boolean: $ => choice('true', 'false'),

    datetime: $ => seq(
      /[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}/,
      optional(seq(':', /[0-9]{2}/, optional(seq(':', /[0-9]{2}/, optional(seq('.', /[0-9]+/)))))),
      choice('Z', seq(/[+-]/, /[0-9]{2}/, optional(seq(':', /[0-9]{2}/))))
    ),

    identifier: $ => /[a-zA-Z][a-zA-Z0-9_-]*/,

    comment: $ => seq('#', /[^\n]*/)
  }
});
