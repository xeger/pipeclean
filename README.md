## What Is This?

Pipeclean sanitizes large data sets efficiently by streaming them from
stdin to stdout:

```
$ cat data.sql | pipeclean -m mysql > sanitized.sql
```

### Install

If you have the Go [SDK](https://go.dev/doc/install) installed, you can `go install github.com/xeger/pipeclean@main`. Otherwise, visit the [release page](https://github.com/xeger/pipeclean/releases/latest) to download binaries.

## Usage

All pipeclean commands accept the `-m` / `--mode` argument which specifies the data type being worked with: `mysql` or `json`.

Pipeclean is invoked with a subcommand that tells it what to do:

```bash
# train models based on input data
cat data.sql | pipeclean --mode mysql learn path/to/models

# sanitize sensitive data; print clean stream to stdout
cat data.sql | pipeclean --mode mysql scrub path/to/models
```

### Configuration

Pipeclean works best when you specify a configuration file to influence its behavior. If none is provided, the [default configuration](scrubbing/policy.go#L26) masks emails, phone numbers and zip codes.

To do something more interesting, author a `pipeclean.json` to define the policy for learning, then scrubbing data:

```json
{
  "learning": {
    "givenName": {
      "markov": {
        "order": 2
      }
    },
    "sn": {
      "markov": {
        "order": 2
      }
    },
  },
  "scrubbing": {
    "fieldname": [
      {
        "in": "email",
        "out": "mask"
      },
      {
        "in": "first_name",
        "out": "generate(givenName)"
      },
      {
        "in": "last_name",
        "out": "generate(sn)"
      }
    ]
  }
}
```

After defining some models, you can train the models using genuine input data. Make sure to create a directory to hold the models. Pipeclean takes the location of its config file and models as CLI parameters:

```bash
cat data.sql | pipeclean learn -m mysql --config pipeclean.json ./data/models
```

To use the trained models, re-run this command but replace the subcommand with `scrub` to generate sanitized output:

```bash
cat data.sql | pipeclean scrub -m mysql -c pipeclean.json ./data/models
```

## Reference Guide

### Models

Pipeclean serializes its models as JSON or text files with corresponding extensions. A sub-extension defines the type of model. Example:

```bash
name.markov.json
phone.match.txt
state-us.dict.txt
```

Markov models record the statistic distirbution of letter and word sequences. Match models specify a regular expression. Dict models use a lookup table of known-good values.

### Training

The `learning` section of config defines some models, each identified by a unique name, and specifies some parameters so that the `learn` command.

Only Markov models need to be trained; match and dict models are static and human-defined.

The learning section specifies some parameters for each named Markov model:
- **delim** specifies how to decompose input strings into sequences
  - an empty string `""` (the default) specifies a word-based model
  - a space `" "` specifies a phrase-based model
- **order** controls how many tokens to look back when deciding on the probability of the next token

### Scrubbing Policy

The `scrubbing` section of config defines which input fields will be sanitized. There are two types of scrubbing rule:
- `fieldname` rules, which specify a regular expression to match the _name_ of an input variable
- `heuristic` rules, which specify a model that can recognize the _values_ of input variables

Field-name rules are good for information that absolutely must be sanitized: PII, financial information, etc. Heuristic rules are useful for defense in depth, and also to handle complex cases such as sanitizing fields that store a particular format of data (base64, YAML, etc) regardless of the field names involved.

The `in` of a rule specifies the matching criteria (a field-name pattern or a model name)); the `out` rule specifies what to do when the rule is matched:

1. `mask` the data by scrambling numbers and letters
2. `erase` the data (replace id with `NULL` in SQL or falsey values in JSON)
3. `generate(modelName)` to create plausible surrogate data from a model

Generation is deterministic and reproducible: given an input string S, the same model will always generate the same derived string S'.
