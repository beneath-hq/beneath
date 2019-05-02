import React from "react";
import Link from "next/link";

const NextMuiLink = React.forwardRef(({ as, href, prefetch, ...props }, ref) => {
  return (
    <Link href={href} prefetch={prefetch} as={as} ref={ref}>
      <a {...props} />
    </Link>
  );
});

export default NextMuiLink;
