import Head from "next/head";

export default (props) => (
  <Head>
    <title>
      {props.title ? props.title + " | " : ""} Beneath
    </title>
  </Head>
);
