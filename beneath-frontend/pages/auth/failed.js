import React, { Component } from "react";
import Error from "../../pages/_error";

class Failed extends Component {
  render() {
    return (
      <Error message="Your login attempt failed. Did you try to create a user? Beneath does not currently allow public sign up." />
    );
  }
}

export default Failed;
