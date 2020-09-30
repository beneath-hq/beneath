import { FetchResult } from "@apollo/client";
import { FormikHelpers } from "formik";

/**
 * Populates Formik errors based on the results of awaiting the supplied
 * Apollo GraphQL mutation promises.
 */
export async function handleSubmitMutation<Values extends {}>(
  values: Values,
  actions: FormikHelpers<Values>,
  ...mutationPromises: Promise<FetchResult<any>>[]
) {
  let status;
  try {
    const responses = await Promise.all(mutationPromises);
    for (const { errors } of responses) {
      if (errors) {
        const unknownErrors = [];
        for (const error of errors) {
          if (error.path && error.path.length > 1) {
            const field = error.path[1];
            if (typeof field === "string" && values.hasOwnProperty(field)) {
              actions.setFieldError(field, error.message);
              continue;
            }
          }
          // error couldn't be matched to a field
          unknownErrors.push(error);
        }
        if (unknownErrors.length > 0) {
          status = unknownErrors.map((error) => {
            const message = error.message.length > 0 ? error.message : JSON.stringify(error);
            return message;
          }).join("\n");
        }
      }
    }
  } catch (error) {
    const message = typeof error?.message === "string" && error.message.length > 0 ? error.message : JSON.stringify(error);
    status = message;
    console.error(error);
  }
  actions.setStatus(status);
  // TODO: Apollo refreshes the entire form (not so good), so cannot call setSubmitting as Formik is unmounted
  // See: https://github.com/formium/formik/issues/772
  // actions.setSubmitting(false);
  return;
}

export default handleSubmitMutation;
