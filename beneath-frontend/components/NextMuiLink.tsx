import Link, { LinkProps } from "next/link";
import React, { FC } from "react";

const NextMuiLink = React.forwardRef<Link, LinkProps>((props, ref) => {
  const { href, as, replace, scroll, shallow, passHref, onError, prefetch, ...others } = props;
  return (
    <Link href={href} as={as} replace={replace} scroll={scroll} shallow={shallow} passHref={passHref} prefetch={prefetch}>
      <a {...others} />
    </Link>
  );
});

export default NextMuiLink;
