import { useQuery, getApolloContext } from "@apollo/client";
import { useContext } from "react";

import { QUERY_ME } from "../apollo/queries/organization";
import { Me } from "../apollo/types/Me";
import { useToken } from "./useToken";

export const useMe = () => {
  // if apollo isn't available, return null (e.g. on 404 page)
  const apolloContext = useContext(getApolloContext());
  if (!apolloContext.client) {
    return null;
  }

  const token = useToken();
  if (!token) {
    return null;
  }

  const { loading, error, data } = useQuery<Me>(QUERY_ME);
  if (data) {
    return data.me;
  }

  if (error) {
    console.error("useMe error: ", error);
  }

  return null;
};

export default useMe;
