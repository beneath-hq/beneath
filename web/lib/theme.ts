import { createMuiTheme, ThemeOptions } from "@material-ui/core/styles";

const theme: ThemeOptions = {}; // the theme we're building
const baseTheme = createMuiTheme(); // for reference to non-overridden values

theme.typography = {
  fontFamily: `-apple-system,BlinkMacSystemFont,Segoe UI,Roboto,Helvetica Neue,Ubuntu,sans-serif`,
  button: {
    fontSize: "1rem",
    fontWeight: 600,
  },
  h1: {
    fontWeight: 600,
    fontSize: "1.75rem",
    letterSpacing: ".05rem",
    wordSpacing: ".15rem",
  },
  h2: {
    fontWeight: 600,
    fontSize: "1.5rem",
    letterSpacing: ".05rem",
    wordSpacing: ".15rem",
  },
  h3: {
    fontWeight: 600,
    fontSize: "1.2rem",
  },
  h4: {
    fontWeight: 600,
    fontSize: "1rem",
  },
};

theme.palette = {
  type: "dark",
  common: {
    black: "rgba(0, 0, 0, 1)",
    white: "rgba(255, 255, 255, 1)",
  },
  background: {
    paper: "rgba(26, 39, 75, 1)",
    default: "rgba(16, 24, 46, 1)",
  },
  divider: "rgba(45, 51, 71, 1)",
  primary: {
    light: "rgba(28, 198, 234, 1)",
    main: "rgba(12, 172, 234, 1)",
    dark: "rgba(12, 134, 210, 1)",
    contrastText: "rgba(255, 255, 255, 0.9)",
  },
  secondary: {
    light: "rgba(251, 149, 54, 1)",
    main: "rgba(100, 120, 140, 1)",
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
    primary: "rgba(255, 255, 255, 0.95)",
    secondary: "rgba(255, 255, 255, 0.7)",
    disabled: "rgba(255, 255, 255, 0.4)",
    hint: "rgba(255, 255, 255, 0.5)",
  },
};

theme.props = {
  MuiButtonBase: {
    disableRipple: true,
  },
  MuiInputLabel: {
    shrink: true,
  },
  MuiTextField: {
    autoComplete: "off",
    inputProps: {
      spellCheck: false,
    },
  },
};

theme.overrides = {
  MuiAppBar: {
    colorPrimary: {
      backgroundColor: theme.palette.background?.default,
    },
  },
  MuiAvatar: {
    root: {
      borderRadius: "4px",
    },
  },
  MuiIconButton: {
    root: {
      color: theme.palette.text?.secondary,
      borderRadius: "4px",
    },
  },
  MuiButton: {
    root: {
      borderRadius: "4px",
      textTransform: "none",
    },
    sizeLarge: {
      padding: "10px 30px",
      fontSize: "1.2rem",
    },
    outlined: {
      borderWidth: 2,
    },
    outlinedPrimary: {
      borderWidth: 2,
      "&:hover": {
        borderWidth: 2,
      },
    },
    outlinedSecondary: {
      borderWidth: 2,
      "&:hover": {
        borderWidth: 2,
      },
    },
  },
  MuiChip: {
    root: {
      // background: "none",
      backgroundColor: theme.palette.background?.default,
      fontWeight: "bold",
      borderRadius: 0,
      paddingLeft: "0px",
      color: theme.palette.text?.secondary
    },
    clickable: {
      "&:hover": {
        background: "none",
        // textDecoration: `underline ${theme.palette.common?.white}`,
        color: theme.palette.text?.primary,
      }
    }
  },
  MuiLinearProgress: {
    colorPrimary: {
      backgroundColor: "rgba(60, 170, 255, 0.25)",
    },
    barColorPrimary: {
      backgroundColor: "rgb(60, 170, 255)",
    },
  },
  MuiDialog: {
    paper: {
      backgroundColor: theme.palette.background?.default,
    },
  },
  MuiTab: {
    root: {
      fontSize: theme.typography.button?.fontSize,
      fontWeight: theme.typography.button?.fontWeight,
      textTransform: "none",
      [baseTheme.breakpoints.up("sm")]: {
        minWidth: "0",
        paddingLeft: "25px",
        paddingRight: "25px",
      },
      "&:hover": {
        color: theme.palette.text?.primary,
      },
    },
  },
};

theme.transitions = {
  duration: {
    shortest: 35,
    shorter: 50,
    short: 60,
    standard: 75,
    complex: 90,
    enteringScreen: 55,
    leavingScreen: 50,
  },
};

theme.zIndex = {
  drawer: 1100,
  appBar: 1200,
  modal: 1300,
};

export default createMuiTheme(theme);
