import React, { Component } from "react";

const backend = "http://ec2-3-138-112-7.us-east-2.compute.amazonaws.com:8080/"

class Login extends Component {
  state = {
    id : "",
    usuario : "",
    pass : "",
  };

  enviarInfo = (e) => {
    console.log({ Id: this.state.id, Pass: this.state.pass, User:this.state.usuario})
    e.preventDefault()
    console.log(JSON.stringify({ Comando: this.state.raw}))
    fetch(backend+"Login", {
      method: "POST",
      body: JSON.stringify({ Id: this.state.id, Pass: this.state.pass, User:this.state.usuario}),
      headers: { "Content-type": "application/json; charset=UTF-8" },
    })
      .then((data) => {
        return data.json();
      })
      .then((res) => {
        console.log(res)
        let docs = []
        for (let a of res.Archivos){
            console.log(a.Uri)
            docs.push({uri: backend+a.Uri})
        }

        this.props.cambio(docs,res.Res)
      });
  };

  render() {
    return (
      <div className="m-4">
        <h2>Login</h2>
        <form>
          <div className="form-group row">
            <label className="col-sm-2 col-form-label">
              Id Particion
            </label>
            <div className="col-sm-10">
              <input
                type="text"
                className="form-control"
                id="IdPart"
                onChange={(e) => this.setState({id:e.target.value})}
              />
            </div>
          </div>
          <div className="form-group row">
            <label className="col-sm-2 col-form-label">
              User
            </label>
            <div className="col-sm-10">
              <input
                type="text"
                className="form-control"
                id="User"
                onChange={(e) => this.setState({usuario:e.target.value})}
              />
            </div>
          </div>
          <div className="form-group row">
            <label className="col-sm-2 col-form-label">
              Password
            </label>
            <div className="col-sm-10">
              <input
                type="text"
                className="form-control"
                id="Pass"
                onChange={(e) => this.setState({pass:e.target.value})}
              />
            </div>
          </div>
          <input type="button" className="btn btn-secondary" value="Ingresar" onClick={this.enviarInfo}/>
        </form>
      </div>
    );
  }
}

export default Login;
