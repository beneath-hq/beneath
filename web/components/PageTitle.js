import Head from "next/head";

export default (props) => (
  <Head>
    <title>
      {props.title ? props.title + " | " : ""} Beneath – Data Science for the Decentralised Economy
    </title>
  </Head>
);
