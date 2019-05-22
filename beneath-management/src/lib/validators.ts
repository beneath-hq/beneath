import { registerDecorator, ValidationOptions, ValidationArguments } from "class-validator";

import avro from "avsc";

export function IsAvroSchema(validationOptions?: ValidationOptions) {
  return function (object: Object, propertyName: string) {
    registerDecorator({
      name: "isAvroSchema",
      target: object.constructor,
      propertyName: propertyName,
      constraints: [],
      options: validationOptions,
      validator: {
        validate(value: any, args: ValidationArguments) {
          let parses = false;
          try {
            if (typeof value === "string") {
              const schema = JSON.parse(value);
              avro.Type.forSchema(schema, { noAnonymousTypes: true });
              parses = true;
            }
          } catch {
          }
          return parses;
        }
      }
    });
  };
}
