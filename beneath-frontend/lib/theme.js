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

export const muiTheme = createMuiTheme({
  palette: {
    type: "dark",
  },
  typography: {
    useNextVariants: true,
    fontFamily: `SFMono-Regular,Menlo,Monaco,Consolas,"Liberation Mono",Courier,monospace`,
  },
});
