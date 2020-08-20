import { useMutation } from "@apollo/client";
import { Field, Form, Formik } from "formik";
import { useRouter } from "next/router";
import React, { FC } from "react";
import Moment from "react-moment";

import { makeStyles, Typography, Container } from "@material-ui/core";

import { UPDATE_ORGANIZATION } from "../../apollo/queries/organization";
import { REGISTER_USER_CONSENT } from "../../apollo/queries/user";
import { OrganizationByName_organizationByName_PrivateOrganization } from "../../apollo/types/OrganizationByName";
import { RegisterUserConsent, RegisterUserConsentVariables } from "../../apollo/types/RegisterUserConsent";
import { UpdateOrganization, UpdateOrganizationVariables } from "../../apollo/types/UpdateOrganization";
import { toBackendName, toURLName } from "../../lib/names";
import { Checkbox as FormikCheckbox, handleSubmitMutation, TextField as FormikTextField } from "../formik";
import SubmitControl from "../forms/SubmitControl";

export interface EditOrganizationProps {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
}

const EditOrganization: FC<EditOrganizationProps> = ({ organization }) => {
  const router = useRouter();

  const [updateOrganization] = useMutation<UpdateOrganization, UpdateOrganizationVariables>(UPDATE_ORGANIZATION, {
    onCompleted: (data) => {
      if (data?.updateOrganization) {
        const name = toURLName(data.updateOrganization.name);
        if (name !== router.query.organization_name) {
          const href = `/organization?organization_name=${name}&tab=edit`;
          const as = `/${name}/-/edit`;
          router.replace(href, as, { shallow: true });
        }
      }
    },
  });

  const [registerUserConsent] = useMutation<RegisterUserConsent, RegisterUserConsentVariables>(REGISTER_USER_CONSENT);

  const initialValues = {
    name: organization.name || "",
    displayName: organization.displayName || "",
    description: organization.description || "",
    photoURL: organization.photoURL || "",
    email: organization.personalUser?.email,
    consentNewsletter: organization.personalUser?.consentNewsletter || false,
  };

  return (
    <Container maxWidth={"sm"}>
      <Formik
        initialValues={initialValues}
        onSubmit={async (values, actions) => {
          const mutationPromises = [];
          mutationPromises.push(
            updateOrganization({
              variables: {
                organizationID: organization.organizationID,
                name: toBackendName(values.name),
                displayName: values.displayName,
                description: values.description,
                photoURL: values.photoURL,
              },
            })
          );
          if (organization.personalUser?.userID) {
            if (values.consentNewsletter !== organization.personalUser.consentNewsletter) {
              mutationPromises.push(
                registerUserConsent({
                  variables: {
                    userID: organization.personalUser.userID,
                    newsletter: values.consentNewsletter,
                  },
                })
              );
            }
          }
          handleSubmitMutation(values, actions, ...mutationPromises);
        }}
      >
        {({ isSubmitting, status }) => (
          <Form>
            <Field
              name="name"
              validate={(val: string) => {
                if (!val || val.length < 3 || val.length > 40) {
                  return "Usernames should be between 3 and 40 characters long";
                }
                if (!val.match(/^[_\-a-z][_\-a-z0-9]+$/)) {
                  return "Usernames should consist of lowercase letters, numbers, underscores and dashes (cannot start with a number)";
                }
              }}
              component={FormikTextField}
              label={organization.personalUser ? "Username" : "Name"}
              required
            />
            <Field
              name="displayName"
              validate={(val: string) => {
                if (!val || val.length === 0 || val.length > 50) {
                  return "Display names should be between 1 and 50 characters";
                }
              }}
              component={FormikTextField}
              label="Display Name"
              required
            />
            <Field
              name="description"
              validate={(val: string) => {
                if (val && val.length > 255) {
                  return "Bios should be shorter than 255 characters";
                }
              }}
              component={FormikTextField}
              label="Bio"
              multiline
              rows={1}
              rowsMax={3}
            />
            <Field name="photoURL" component={FormikTextField} label="Photo URL" disabled />
            {organization.personalUser && (
              <Field name="email" component={FormikTextField} type="email" label="Email" required disabled />
            )}
            {organization.personalUser && (
              <Field
                name="consentNewsletter"
                component={FormikCheckbox}
                type="checkbox"
                label="Subscribe to newsletter"
              />
            )}
            <SubmitControl
              label="Save changes"
              createdOn={organization.createdOn}
              updatedOn={organization.updatedOn}
              errorAlert={status}
              disabled={isSubmitting}
            />
          </Form>
        )}
      </Formik>
    </Container>
  );
};

export default EditOrganization;
