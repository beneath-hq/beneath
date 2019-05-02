import { createMuiTheme } from "@material-ui/core/styles";

// const baseTheme = createMuiTheme();

export default createMuiTheme({
  palette: {
    type: "dark",
    common: {
      black: "rgba(0, 0, 0, 1)",
      white: "rgba(255, 255, 255, 1)",
    },
    background: {
      paper: "rgba(35, 47, 74, 1)",
      default: "rgba(16, 24, 46, 1)",
    },
    primary: {
      light: "rgba(28, 198, 234, 1)",
      main: "rgba(12, 172, 234, 1)",
      dark: "rgba(12, 134, 210, 1)",
      contrastText: "rgba(255, 255, 255, 0.9)",
    },
    secondary: {
      light: "rgba(251, 149, 54, 1)",
      main: "rgba(223, 110, 40, 1)",
      dark: "rgba(217, 86, 35, 1)",
      contrastText: "#fff",
    },
    error: {
      light: "rgba(252, 86, 50, 1)",
      main: "rgba(235, 30, 7, 1)",
      dark: "rgba(192, 26, 7, 1)",
      contrastText: "#fff",
    },
    text: {
      primary: "rgba(255, 255, 255, 0.9)",
      secondary: "rgba(255, 255, 255, 0.7)",
      disabled: "rgba(255, 255, 255, 0.4)",
      hint: "rgba(255, 255, 255, 0.5)",
    },
  },
  props: {
    MuiButtonBase: {
      disableRipple: true,
    },
  },
  overrides: {
    MuiButton: {
      root: {
        borderRadius: 0,
      },
    },
    MuiAppBar: {
      colorPrimary: {
        backgroundColor: "rgba(16, 24, 46, 1)",
      }
    },
  },
  transitions: {
    duration: {
      shortest: 75,
      shorter: 100,
      short: 125,
      standard: 150,
      complex: 190,
      enteringScreen: 115,
      leavingScreen: 100,
    },
  },
  typography: {
    useNextVariants: true,
    fontFamily: `SFMono-Regular,Menlo,Monaco,Consolas,"Liberation Mono",Courier,monospace`,
    h1: {
      fontWeight: 600,
    },
    h2: {
      fontWeight: 600,
    },
    h3: {
      fontWeight: 600,
    },
    h4: {
      fontWeight: 600,
    },
  },
  zIndex: {
    appBar: 1300,
    drawer: 1100,
    modal: 1200,
  },
});
