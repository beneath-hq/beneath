import React, { Component } from "react";
import App from "../../components/App";
import { attemptLogin } from "../../lib/auth";

class Login extends Component {
  componentDidMount() {
    attemptLogin();
  }

  render() {
    return <App />;
  }
}

export default Login;
