import { Link, Route, Routes } from "react-router-dom";
import { useState } from "react";
import "./App.css";
import Inicio from "./Components/Inicio";
import Interfaz from "./Components/InterfazU";

function App() {
  const [mostrar, setMostrar] = useState({
    mostrar: "collapse navbar-collapse",
    area: "false",
  });

  const clickHandle = () => {
    if (mostrar.mostrar === "collapse navbar-collapse") {
      setMostrar({
        mostrar: "collapse navbar-collapse show",
        area: "true",
      });
    } else {
      setMostrar({
        mostrar: "collapse navbar-collapse",
        area: "false",
      });
    }
  };

  return (
    <div className="App">
      <nav className="navbar navbar-expand-lg navbar-dark bg-dark">
        <div className="container-fluid">
          <button
            className="navbar-toggler"
            type="button"
            data-bs-toggle="collapse"
            data-bs-target="#navbarColor02"
            aria-controls="navbarColor02"
            aria-expanded={mostrar.area}
            aria-label="Toggle navigation"
            onClick={clickHandle}
          >
            <span className="navbar-toggler-icon"></span>
          </button>
          <div className={mostrar.mostrar} id="navbarColor02">
            <ul className="navbar-nav me-auto">
              <li className="nav-item">
                <Link className="nav-link" to="/">
                  Home
                </Link>
              </li>
              <li>
                <Link className="nav-link" to="/reportes">
                  Login
                </Link>
              </li>
            </ul>
          </div>
        </div>
      </nav>
      <Routes>
        <Route path="/" element={<Inicio />}></Route>
        <Route path="/reportes" element={<Interfaz />}></Route>
      </Routes>
    </div>
  );
}

export default App;
