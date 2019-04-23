import { createMuiTheme } from "@material-ui/core/styles";

const screenWidths = {
  mobileS: 320,
  mobileM: 375,
  mobileL: 425,
  tablet: 768,
  laptop: 1024,
  laptopL: 1440,
  desktop: 2560,
};

export const devices = {
  mobileSOrLarger: `screen and (min-width: ${screenWidths.mobileS}px)`,
  mobileMOrLarger: `screen and (min-width: ${screenWidths.mobileM}px)`,
  mobileLOrLarger: `screen and (min-width: ${screenWidths.mobileL}px)`,
  tabletOrLarger: `screen and (min-width: ${screenWidths.tablet}px)`,
  laptopOrLarger: `screen and (min-width: ${screenWidths.laptop}px)`,
  laptopLOrLarger: `screen and (min-width: ${screenWidths.laptopL}px)`,
  desktopOrLarger: `screen and (min-width: ${screenWidths.desktop}px)`,
  desktopLOrLarger: `screen and (min-width: ${screenWidths.desktop}px)`,
  smallerThanMobileS: `screen and (max-width: ${screenWidths.mobileS - 1}px)`,
  smallerThanMobileM: `screen and (max-width: ${screenWidths.mobileM - 1}px)`,
  smallerThanMobileL: `screen and (max-width: ${screenWidths.mobileL - 1}px)`,
  smallerThanTablet: `screen and (max-width: ${screenWidths.tablet - 1}px)`,
  smallerThanLaptop: `screen and (max-width: ${screenWidths.laptop - 1}px)`,
  smallerThanLaptopL: `screen and (max-width: ${screenWidths.laptopL - 1}px)`,
  smallerThanDesktop: `screen and (max-width: ${screenWidths.desktop - 1}px)`,
  smallerThanDesktopL: `screen and (max-width: ${screenWidths.desktop - 1}px)`,
};

const baseTheme = createMuiTheme();

export const muiTheme = createMuiTheme({
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
  overrides: {
    MuiButton: {
      root: {
        borderRadius: 0,
      },
    },
  },
});
