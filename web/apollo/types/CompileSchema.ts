/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { CompileSchemaInput } from "./globalTypes";

// ====================================================
// GraphQL query operation: CompileSchema
// ====================================================

export interface CompileSchema_compileSchema {
  __typename: "CompileSchemaOutput";
  canonicalIndexes: string;
}

export interface CompileSchema {
  compileSchema: CompileSchema_compileSchema;
}

export interface CompileSchemaVariables {
  input: CompileSchemaInput;
}
