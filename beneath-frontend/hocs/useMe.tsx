import React from "react";
import { useQuery } from "react-apollo";

import { QUERY_ME } from "../apollo/queries/user";
import { useToken } from "./auth";

import { Me } from "../apollo/types/Me";

export const useMe = () => {
  const token = useToken();
  if (!token) {
    return null;
  }

  const { loading, error, data } = useQuery<Me>(QUERY_ME);
  if (data) {
    return data.me;
  }

  if (error) {
    console.error("withMe error: ", error);
  }

  return null;
};

export default useMe;
