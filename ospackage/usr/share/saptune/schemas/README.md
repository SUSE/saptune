# JSON schema for saptune

This directory contains one directory for each schema version.
Currently:

- `1.0/`: for saptune v3.1
- `1.1/`: for saptune v3.2

Each directory contains:

- `examples/`\
   Contains:
   - JSON example out for various commands: `saptune_COMMAND.json`
   - the script `make_examples` to create the examples
   - the script `valdate_examples` to check the examples with the schema files.
   - a `README.md` describing how to create and validate the examples
   Read `README.md` if schemas or command output has been altered and you need to check the changed schemas. 

   > :bulb: The `README.md` and both scripts `make_examples` and `valdate_examples` have been added in v1.1.

- `templates/`\
  Contains:
  - the schema templates for the commands including `common.schema.json.template`
  - the schema build script `generate_unsupported.sh` for commands without JSON support
  - the schema build script `build.py` for commands with JSON support
  - a `README.md` describing how to build the schemas from the templates
  Read `README.md` if schemas need to be altered. 

- `CHANGELOG`\
    Describes the changes from the previous version. Obviously not present in `1.0/`

- `saptune_*.schema.json`\
    The generated JSON schema files for each command with the naming scheme:
    `saptune_COMMAND.schema.json` as well as `saptune_invalid.schema.json` for invalid commands.
