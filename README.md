# What Is This?

Pipeclean removes sensitive information from large-text data sets efficiently by streaming them from stdin to stdout:

```
$ cat data.sql | pipeclean -m mysql > sanitized.sql
```

It utilizes only streaming parsers, allowing constant memory usage even for large input files. Some modes employ parallelism up to `runtime.NumCPU()` which promotes fast, efficient operation. Sanitizing a 500 MiB MySQL dump takes ~30 seconds with peak memory usage of ~4 GiB on a 2020-era MacBook Pro M1 with eight cores.

Pipeclean attempts to sanitize encapsulated data, too; if a MySQL column contains JSON or YAML, it will parse and traverse the encapsulated document and write sanitized JSON/YAML as the output column value.

Finally, pipeclean can employ language models to generate plausible-looking replacement data; you can train it with actual peoples' names from your database, for example, and at runtime, use the trained model to generate replacement names that have a similar look and feel. Language models can also be used for heuristic scrubbing, allowing pipeclean to recognize data fields that contain a first name regardless of how the field is named or where in the input it appears, and replace each occurrence with a fake name.

## Install

If you have the Go [SDK](https://go.dev/doc/install) installed, you can `go install github.com/xeger/pipeclean@main`. Otherwise, visit the [release page](https://github.com/xeger/pipeclean/releases/latest) to download binaries.

# Usage Overview

This is a summary; for more detailed information, see [the reference guide](REFERENCE.md).

All pipeclean commands accept the `-m` / `--mode` argument which specifies the data type being worked with: `mysql` or `json`.

Pipeclean is invoked with a subcommand that tells it what to do:

```bash
# train models based on input data
cat data.sql | pipeclean --mode mysql learn path/to/models

# sanitize sensitive data; print clean stream to stdout
cat data.sql | pipeclean --mode mysql scrub path/to/models
```

## Configuration

Pipeclean works best when you specify a configuration file to influence its behavior. If none is provided, the [default configuration](scrubbing/policy.go#L26) masks emails, phone numbers and zip codes by scrambling individual characters.

To customize its behavior, author a `pipeclean.json` to define some models and a scrubbing policy of your own:

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

## Training

After defining some models in the configuration, you can train the models using genuine input data. Make sure to create a directory to hold the models. Pipeclean takes the location of its config file and models as CLI parameters:

```bash
cat data.sql | pipeclean learn -m mysql --config pipeclean.json ./data/models
```

To use the trained models, re-run this command but replace the subcommand with `scrub` to generate sanitized output:

```bash
cat data.sql | pipeclean scrub -m mysql -c pipeclean.json ./data/models
```

Pipeclean will use your scrubbing policy to identify input fields and train the corresponding model with the real value. At the end, it will write the trained models to JSON files under `./data/models`, and those models can be used with future invocations of the command.

## Providing Context

The `learn` and `scrub` commands both accept a `-x` / `--context` flag, which is a list of extra files that pipeclean should parse to learn about the structure of data. The contents of these files do not appear in pipeclean's output nor contribute to the training of models.

Context is important! For example, in MySQL dumps, the `INSERT` statement use a shorthand form that does not specify column names. Without context, your pipeclean rules need to refer to columns by their insertion index:

```
{ "in": "users.3", "out": "mask" }
```

If you produce your MySQL dump as two files, a `schema.sql` produced with `mysqldump --no-data` and a `data.sql` produced with `mysqldump --no-create-info`, you can tell pipeclean about your column names with an `-x schema.sql` flag. This allows your configuration to specify column names instead of indices:

```
{ "in": "users.email", "out": "mask" }
```
