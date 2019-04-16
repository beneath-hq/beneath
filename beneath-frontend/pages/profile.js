import React, { Component } from "react";
import Page from "../components/Page";
import { MainSidebar } from "../components/Sidebar";
import { Query } from "react-apollo";
import gql from "graphql-tag";

const USER_QUERY = gql`
  query {
    me {name}
  }
`;

class UserForm extends Component {
    constructor(props) {
    super(props);

    this.state = {
      username: "Test Hest",
    };

    this.handleChangeUsername = this.handleChangeUsername.bind(this);
    this.handleSubmit = this.handleSubmit.bind(this);
  }

  handleChangeUsername(event) {
    const username = event.target.value
    this.setState({
        username
    });
  };

  handleSubmit(event) {
    event.preventDefault();
    alert(`You submitted name: ${this.state.username}`);
  };

  render() {
    return (
      <form onSubmit={this.handleSubmit}>
        Name:
        <input type="text" value={this.state.username} onChange={this.handleChangeUsername} />
        <button type="submit">Save</button>
      </form>
    );
  }
}

export default () => (
  <Page title="Profile" sidebar={<MainSidebar />}>
    <div className="section">
      <div className="title">
        <h1>Profile</h1>
        <UserForm />
      </div>
      {/* 
        Section 1: Edit name and bio
        Section 2: Create new API key -- readonly (green) or readwrite (red)
        Section 3: View all API keys (show type)
      */}
    </div>
  </Page>
);
