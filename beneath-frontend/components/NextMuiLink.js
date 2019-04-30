import Link from "next/link";

const NextMuiLink = ({ as, href, prefetch, ...props }) => {
  return (
    <Link href={href} prefetch={prefetch} as={as}>
      <a {...props} />
    </Link>
  );
};

export default NextMuiLink;