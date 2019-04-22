// For MUI + Next.js, see https://github.com/mui-org/material-ui/blob/master/examples/nextjs/src/getPageContext.js

import { SheetsRegistry } from "jss";
import { createGenerateClassName } from "@material-ui/core/styles";
import { muiTheme } from "./theme";

function createPageContext() {
  return {
    theme: muiTheme,
    // This is needed in order to deduplicate the injection of CSS in the page.
    sheetsManager: new Map(),
    // This is needed in order to inject the critical CSS.
    sheetsRegistry: new SheetsRegistry(),
    // The standard class name generator.
    generateClassName: createGenerateClassName(),
  };
}

let pageContext;

export default function getPageContext() {
  // Make sure to create a new context for every server-side request so that data
  // isn't shared between connections (which would be bad).
  if (!process.browser) {
    return createPageContext();
  }

  // Reuse context on the client-side.
  if (!pageContext) {
    pageContext = createPageContext();
  }

  return pageContext;
}
