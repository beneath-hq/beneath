import MuiLink, { LinkProps as MuiLinkProps } from "@material-ui/core/Link";
import NextLink, { LinkProps as NextLinkProps } from "next/link";
import React, { FC } from "react";

export const NakedLink = React.forwardRef<any, React.PropsWithChildren<NextLinkProps>>((props, ref) => {
  const { href, as, replace, scroll, shallow, passHref, prefetch, ...others } = props;

  // if it's an external href, use a normal anchor
  if (typeof href === "string" && href.indexOf("http") === 0) {
    return <a ref={ref} style={{ textDecoration: "none" }} href={href} {...others} />;
  }

  return (
    <NextLink
      href={href}
      as={as}
      replace={replace}
      scroll={scroll}
      shallow={shallow}
      passHref={passHref}
      prefetch={prefetch}
    >
      <a ref={ref} style={{ textDecoration: "none" }} {...others} />
    </NextLink>
  );
});

type LinkProps = NextLinkProps & MuiLinkProps & {
  href: string;
  bold?: boolean;
  children?: any;
};

export const Link: FC<LinkProps> = (props: LinkProps) => {
  const { bold, children, ...others } = props;
  return <MuiLink component={NakedLink} {...others}>{children}</MuiLink>;
};
