import Head from "next/head";

const PageTitle = props => (
  <Head>
    <title>
      {props.subtitle ? props.subtitle + " | " : ""}
      Beneath – Data Science for the Decentralised Economy
    </title>
  </Head>
);

export default PageTitle;
