import { ApiSidebar } from "./Sidebar";

const burgerMenuClick = (event, props) => {
  event.currentTarget.classList.toggle("change");
  props.clickHandler();
};

export const BurgerMenu = props => (
  <div className="container" onClick={event => burgerMenuClick(event, props)}>
    <div className="bar1" />
    <div className="bar2" />
    <div className="bar3" />
    <style jsx>{`
      .container {
        cursor: pointer;
      }

      .bar1,
      .bar2,
      .bar3 {
        width: 35px;
        height: 5px;
        background-color: #50c0ff;
        margin: 6px 0;
        transition: 0.4s;
      }

      .change .bar1 {
        transform: rotate(-45deg) translate(-9px, 6px);
      }

      .change .bar2 {
        opacity: 0;
      }

      .change .bar3 {
        transform: rotate(45deg) translate(-8px, -8px);
      }
    `}</style>
  </div>
);

export const SideDrawer = props => {
  let drawerClasses = "side-drawer";
  if (props.show) {
    drawerClasses = "side-drawer open";
  }
  return (
    <div>
      <div className={drawerClasses}>
        <ApiSidebar />
      </div>
      <style jsx>{`
        .side-drawer {
          height: 90%;
          background: #101830;
          box-shadow: 1px 0px 7px rgba(0, 0, 0, 0.5);
          position: fixed;
          top: 97px;
          left: 0;
          width: 70%;
          max-width: 400px;
          z-index: 2;
          transform: translateX(-100%);
          transition: transform 0.3s ease-out;
        }

        .side-drawer.open {
          transform: translateX(0);
        }
      `}</style>
    </div>
  );
};