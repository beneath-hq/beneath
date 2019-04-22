import css from "styled-jsx/css";
import { devices } from "../lib/theme";

import Header from "./Header";

export default ({ children }) => (
  <div className="main">
    <Header />
    {children}
    {/* <style jsx global>
      {global}
    </style>
    <style jsx>{`
      div.main {
        display: flex;
        flex-direction: column;
        height: 100vh;
        flex-grow: 1;
      }
    `}</style> */}
  </div>
);

const global = css.global`
  * {
    font-family: SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", Courier, monospace;
  }

  body {
    margin: 0;
    padding: 0;
    height: 100vh;
    color: #e1e1e1;
    background-color: #101830; // #050d20
  }

  @media ${devices.mobileLOrLarger} {
    p, button, aside {
      font-size: 14px;
      line-height: 1.6;
    }  
    h2 {
      font-size: 1.5em;
    }
    h3 {
      font-size: 1.17em;
    }
  }

  @media ${devices.mobileLOrLarger} {
    p, button, aside {
      font-size: 17px;
      line-height: 1.8;
    }
    h2 {
      font-size: 1.5em;
    }
    h3 {
      font-size: 1.17em;
    }  
  }

  a {
    color: #e1e1e1;
    text-decoration: none;
  }

  p > a {
    color: #50c0ff;
    font-weight: 600;
    text-decoration: none;
  }

  p > a:hover {
    text-decoration: underline;
  }

  article {
    margin: 0 auto;
    max-width: 900px;
  }

  pre {
    border: 1px solid #405070;
    padding: 10px;
    background: hsla(0, 0% ,100% , 0.05);
  }

  @media ${devices.smallerThanLaptop} {
    article {
      padding-left: 20px;
      padding-right: 20px;
    }
  }

  button {
    cursor: pointer;
    outline: none; 
    background: none;
    align-items: center;
    white-space: nowrap;
    padding: 10px 20px;

    border-width: 1px;
    border-style: solid;
    border-color: white;

    color: white;
    display: flex;
  }

  button:hover, button:focus, button:active {
    color: #50c0ff;
    border-color: #50c0ff;
  }
`;
