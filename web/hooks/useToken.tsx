import { useQuery } from "@apollo/client";

import { GET_TOKEN } from "../apollo/queries/local/token";
import { Token } from "../apollo/types/Token";

export const useToken = () => {
  const { loading, error, data } = useQuery<Token>(GET_TOKEN);
  if (data) {
    const { token } = data;
    return token;
  }

  return null;
};
