import { useQuery } from "@apollo/client";

import { QUERY_ME } from "../apollo/queries/organization";
import { Me } from "../apollo/types/Me";
import { useToken } from "./useToken";

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
    console.error("useMe error: ", error);
  }

  return null;
};

export default useMe;
