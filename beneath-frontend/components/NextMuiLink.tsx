import Link from "next/link";
import React, { FC } from "react";

interface IProps {
  href: string;
  as?: string;
  prefetch?: boolean;
  replace?: boolean;
  shallow?: boolean;
}

const NextMuiLink = React.forwardRef<any, IProps>(({ as, href, prefetch, replace, shallow, ...props }, ref) => {
  return (
    <Link href={href} prefetch={prefetch} as={as} ref={ref} replace={replace} shallow={shallow}>
      <a {...props} />
    </Link>
  );
});

export default NextMuiLink;
