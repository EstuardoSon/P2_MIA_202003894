import React, { Component } from "react";

const backend = "http://ec2-3-138-112-7.us-east-2.compute.amazonaws.com:8080/"

class Inicio extends Component {
  state = { raw: "", respuesta: "" };

  readFile = (e) => {
    const file = e.target.files[0];
    const fileReader = new FileReader();

    if (file != null) {
      fileReader.readAsText(file);

      fileReader.onload = () => {

        this.setState({ raw: fileReader.result });
      };

      fileReader.onerror = () => {
        console.log(fileReader.error);
      };
    }
  };

  enviarInfo = () => {
    fetch(backend, {
      method: "POST",
      body: JSON.stringify({ Comando: this.state.raw}),
      headers: { "Content-type": "application/json; charset=UTF-8" },
    })
      .then((data) => {
        console.log(data);
        return data.json();
      })
      .then((res) => {
        this.setState({ respuesta: res.Res });
      });
  };

  render() {
    return (
      <div className="m-4">
        <div>
          <input
            type="file"
            className="form-control"
            id="formFile"
            onChange={this.readFile}
            accept=".eea"
          ></input>
        </div>
        <div className="m-4">
          <textarea
            style={{ maxWidth: "95%" }}
            cols={120}
            rows={10}
            value={this.state.raw}
            onChange={(e) => {
              this.setState({ raw: e.target.value });
            }}
          ></textarea>
        </div>
        <div className="m-4">
          <button
            type="button"
            className="btn btn-secondary"
            onClick={this.enviarInfo}
          >
            Ejecutar
          </button>
        </div>
        <div className="m-4">
          <textarea
            style={{ maxWidth: "95%" }}
            cols={120}
            rows={10}
            value={this.state.respuesta}
            readOnly
          ></textarea>
        </div>
      </div>
    );
  }
}

export default Inicio;
