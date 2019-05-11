import React from "react";
import Link from "next/link";

const NextMuiLink = React.forwardRef(({ as, href, prefetch, replace, shallow, ...props }, ref) => {
  return (
    <Link href={href} prefetch={prefetch} as={as} ref={ref} replace={replace} shallow={shallow}>
      <a {...props} />
    </Link>
  );
});

export default NextMuiLink;
