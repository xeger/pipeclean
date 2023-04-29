# Pipeclean Reference Guide

## Mode

All pipeclean subcommands accept a `-m` / `--mode` flag that defines the data format being worked with; currently, only `mysql` is well-tested and `json` is provided as a proof of concept.

In `mysql` mode, the input **must** be formatted in the style of `mysqldump` with one statement per line. If a statement spans lines, it will fail to parse and pipeclean will emit it, unchanged, to the output. MySQL scrubbing uses parallelism.

In the (experimental) `json` mode, the input is a JSON document and that is parsed by `encoding/json.NewDecoder()` so it _may_ stream, but this has not been explored. JSON mode does not (yet) use parallelism and **may not properly apply rules**.

## Scrubbing

The `scrub` command parses fragments of structured data from stdin, applies sanitization rules, and prints the result to stdout.

```
pipeclean scrub < -m mode > [ modelsDir1, [ modelsDir2, ... ] ]
```

With `scrub`, the command-line parameters specify a list of directories where model files are stored. Pipeclean loads every recognized file from every specified directory, and validates the loaded models against the configuration before parsing stdin.

### Configuration for Scrubbing

The `scrubbing` section of config defines which input fields will be sanitized. There are two types of scrubbing rule:

1. `fieldname` rules, which specify a regular expression to match the _name_ of an input variable
2. `heuristic` rules, which specify a model that can recognize the _values_ of input variables

#### Field-Name Rules

Field-name rules are good for information that absolutely must be sanitized: PII, financial information, etc. Heuristic rules are useful for defense in depth, and also to handle complex cases such as sanitizing fields that store a particular format of data (base64, YAML, etc) regardless of the field names involved.

The `in` of a field-name rule specifies the matching criteria (a field-name pattern or a model name)). It is parsed as a regular expression, allowing wildcard and other matching characters. (Note that the `.` character is syntactically significant and should theoretically be escaped, although it is typically harmless to simply type `.` and allow that to match a literal `.` or any other single character.)

The `out` of a rule specifies what to do when the rule is matched:

1. `mask` the data by scrambling numbers and letters
2. `erase` the data (replace id with `NULL` in SQL or falsey values in JSON)
3. `generate(modelName)` to create plausible surrogate data from a model

Generation is deterministic and reproducible: given an input string S, the same model will always generate the same derived string S'. Determinism is important because it preserves referential consistency of the data set: if two people share a phone number, address, etc, then that fact is preserved in the sanitized output.

#### Heuristic Rules

Heuristic rules specify a model name as the `in`. The `out` is identical to field-name rules. When the value of a field is recognized by the model, that heuristic rule is used to scrub the field.

### Rule Ordering

Field-name rules are matched **in the order that they appear in configuration**. Make sure to specify more specific field-names _first_ to avoid rule-matching ambiguity. For example, we might have two tables with special handling of their email field, as well as a catch-all rule for email fields in general:

```json
{ "in": "users.email", out: "generate(userEmail)" },
{ "in": "password_reset_requests.email", "out": "erase" },
{ "in": "email", "out": "erase" },
```

Field-name rules are always evaluated first, followed by heuristic rules.

### Matching Fields By Multiple Names

When pipeclean applies field-name rules to a piece of data, it matches the rule set against _several_ potential field names for that data. As an example, let us say that it is handling a column of a MySQL `INSERT` statement. It knows that it is processing column 3 of a in row the `users` table, and it has parsed the table definition from a schema provided via `--context`, so it knows column 3 is named `email`. It will attempt to find a scrubbing rule that matches _any_ of the following field names, trying to match each name in order against all of the rules, in sequence:
- `email`
- `users.email`
- `users.0`

## Learning

The `learn` command parses fragments of structured data from stdin, infers the relevant model for each field, and if that model is trainable, uses the field data to train the model. It trains all models concurrently from the same input data.

```
pipeclean scrub < -m mode > [ -c configFile ] [ modelsDir1, [ modelsDir2, ... ] ]
```

Similar to the `scrub` command, `learn` accepts a model-file directory as a mandatory command line argument. (One difference is that `learn` can only use a single directory, because it outputs new files to that directory.) By default, `learn` replaces existing models entirely; use the `-a` / `--append` flag to add to them instead.

Learning does not yet employ parallelism, so it may run slower than scrubbing.

### Model Files

Pipeclean serializes its models as JSON or text files with corresponding extensions. A sub-extension defines the type of model. Example:

```bash
name.markov.json
phone.match.txt
state-us.dict.txt
```

Markov models record the statistic distirbution of letter and word sequences. Match models specify a regular expression. Dict models use a lookup table of known-good values.

### Configuration for Learning

The `learning` section of config defines some models, each identified by a unique name, and specifies some parameters so that the `learn` command.

Only Markov models need to be trained; match and dict models are static and human-defined.

The learning section specifies some parameters for each named Markov model:
- **delim** specifies how to decompose input strings into sequences
  - an empty string `""` (the default) specifies a word-based model
  - a space `" "` specifies a phrase-based model
- **order** controls how many tokens to look back when deciding on the probability of the next token

## Auxiliary Commands

**TODO:** cover train, extract, generate, recognize

