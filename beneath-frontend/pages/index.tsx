import { NextPage } from "next";
import Router from "next/router";
import React from "react";

const Index: NextPage = (props) => {
  return <></>;
};

Index.getInitialProps = async ({ res }) => {
  if (res) {
    res.writeHead(302, {
      Location: "/explore",
    });
    res.end();
  } else {
    Router.replace("/explore");
  }
  return {};
};

export default Index;
