The `fanout` service consumes the `raw-changes` table produced by the `generator` service and writes each change to its own schemaful table.

Schema evolution is not yet supported.