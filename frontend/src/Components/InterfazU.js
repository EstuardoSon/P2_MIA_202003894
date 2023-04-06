import React, { Component } from "react";
import Login from "./Login";
import Reportes from "./Reportes";

class Interfaz extends Component {
  state = {
    Archivos: {},
    Verificar: false,
  };

  cambio = (a, v) => {
    this.setState({
      Archivos: a,
      Verificar: v,
    });
  };

  render() {
    if (!this.state.Verificar) {
      return <Login cambio={this.cambio}></Login>;
    } else {
      return <Reportes cambio={this.cambio} Archivos={this.state.Archivos}/>;
    }
  }
}

export default Interfaz;
