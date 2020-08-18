import { NextPage } from "next";
import React from "react";

import ErrorPage from "../components/ErrorPage";

const FourOhFourPage: NextPage = () => {
  return <ErrorPage statusCode={404} />;
};

export default FourOhFourPage;
