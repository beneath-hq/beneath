import Link, { LinkProps } from "next/link";
import React from "react";

type NextMuiLinkListProps = React.AnchorHTMLAttributes<HTMLAnchorElement> & LinkProps;

const NextMuiLinkList = React.forwardRef(function NextComposed(props: NextMuiLinkListProps, ref: React.Ref<any>) {
  const { href, as, replace, scroll, shallow, passHref, prefetch, ...others } = props;

  return (
    <Link
      href={href}
      as={as}
      replace={replace}
      scroll={scroll}
      shallow={shallow}
      passHref={passHref}
      prefetch={prefetch}
    >
      <a ref={ref} {...others} />
    </Link>
  );
});

export default NextMuiLinkList;
