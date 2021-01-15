import { colors } from "@material-ui/core";
import { createMuiTheme, ThemeOptions, responsiveFontSizes, rgbToHex } from "@material-ui/core/styles";

// "hack" to add custom properties to MUI theme
declare module "@material-ui/core/styles/createTypography" {
  interface TypographyOptions {
    fontFamilyMonospaced: React.CSSProperties["fontFamily"];
  }
  interface Typography {
    fontFamilyMonospaced: React.CSSProperties["fontFamily"];
  }
}
declare module "@material-ui/core/styles/createPalette" {
  interface TypeBackground {
    medium: string;
  }
  interface PaletteOptions {
    border: {
      background: string;
      paper: string;
    };
    purple: {
      main: string;
    };
    rainbow: string[];
  }
  interface Palette {
    border: {
      background: string;
      paper: string;
    };
    purple: {
      main: string;
    };
    rainbow: string[];
  }
}

const theme: ThemeOptions = {}; // the theme we're building
const baseTheme = createMuiTheme(); // for reference to non-overridden values

theme.typography = {
  fontFamily: `-apple-system,BlinkMacSystemFont,Segoe UI,Roboto,Helvetica Neue,Ubuntu,sans-serif`,
  fontFamilyMonospaced: `Menlo,Monaco,Consolas,"Liberation Mono","Courier New",monospace`,
  button: {
    fontSize: "1rem",
    fontWeight: 600,
  },
  h1: {
    fontWeight: 600,
    fontSize: "1.75rem",
    wordSpacing: ".15rem",
  },
  h2: {
    fontWeight: 600,
    fontSize: "1.5rem",
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
    paper: rgbToHex("rgb(26, 39, 75)"),
    medium: rgbToHex("rgb(21, 31, 60)"),
    default: rgbToHex("rgb(16, 24, 46)"),
  },
  divider: "rgba(45, 51, 71, 1)",
  border: {
    background: "rgba(45, 51, 71, 1)",
    paper: "rgba(39, 55, 90)",
  },
  primary: {
    light: "rgba(28, 198, 234, 1)",
    main: "rgba(12, 172, 234, 1)",
    dark: "rgba(12, 134, 210, 1)",
    contrastText: "rgba(255, 255, 255, 0.9)",
  },
  secondary: {
    light: "rgba(130, 160, 190, 1)",
    main: "rgba(44, 58, 86, 1)",
    dark: "rgba(80, 95, 120, 1)",
    contrastText: "#fff",
  },
  purple: {
    main: "rgba(146, 0, 255, 1)",
  },
  rainbow: [
    colors.amber["A100"],
    colors.blue["A100"],
    colors.blueGrey["A100"],
    colors.brown["A100"],
    colors.cyan["A100"],
    colors.deepOrange["A100"],
    colors.deepPurple["A100"],
    colors.green["A100"],
    colors.indigo["A100"],
    colors.lightBlue["A100"],
    colors.lime["A100"],
    colors.orange["A100"],
    colors.pink["A100"],
    colors.purple["A100"],
    colors.red["A100"],
    colors.teal["A100"],
    colors.yellow["A100"],
  ],
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
  MuiAppBar: {
    elevation: 0,
  },
  MuiButtonBase: {
    disableRipple: true,
  },
  MuiButtonGroup: {
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
    root: {
      boxShadow: `inset 0 -1px 0 0 ${theme.palette.border.paper}`, // boxShadow doesn't stack with tab indicator, unlike borderBottom
    },
    colorPrimary: {
      backgroundColor: theme.palette.background?.medium,
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
  MuiRadio: {
    root: {
      borderRadius: "50%",
    },
  },
  MuiTypography: {
    gutterBottom: {
      marginBottom: "1.0rem",
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
      borderWidth: 1,
    },
    outlinedPrimary: {
      borderWidth: 1,
      "&:hover": {
        borderWidth: 1,
      },
    },
    outlinedSecondary: {
      borderWidth: 1,
      "&:hover": {
        borderWidth: 1,
      },
    },
  },
  MuiChip: {
    root: {
      backgroundColor: theme.palette.border.background,
      borderRadius: "4px",
      height: "28px",
    },
    sizeSmall: {
      height: "18px",
    },
    label: {
      fontSize: theme.typography.caption?.fontSize,
      paddingLeft: "8px",
      paddingRight: "8px",
    },
    labelSmall: {
      fontSize: theme.typography.caption?.fontSize,
      paddingLeft: "6px",
      paddingRight: "6px",
    },
    outlined: {
      borderWidth: 1,
    },
    clickable: {
      "&:hover": {
        background: "none",
        color: theme.palette.text?.primary,
      },
    },
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
  MuiPaper: {
    outlined: {
      border: `1px solid ${theme.palette.border.paper}`,
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
  MuiTableRow: {
    root: {
      backgroundColor: theme.palette.background?.paper,
      "&:last-child": {
        "& .MuiTableCell-root": {
          borderBottom: "none",
        },
      },
    },
    head: {
      backgroundColor: theme.palette.background?.medium,
    },
  },
  MuiTableCell: {
    root: {
      borderBottom: `1px solid ${theme.palette.border.paper}`,
    },
  },
  MuiFormControl: {
    marginNormal: {
      marginTop: 20,
      marginBottom: 10,
    },
    marginDense: {
      marginTop: 10,
      marginBottom: 5,
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

export default responsiveFontSizes(createMuiTheme(theme));
