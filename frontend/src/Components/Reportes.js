import React, { Component } from "react";
import DocViewer, { DocViewerRenderers } from "@cyntler/react-doc-viewer";

class Reportes extends Component {
  state = {};

  render() {
    return (
      <div className="m-4">
        <div className="row m-4">
          <input
            type="button"
            onClick={() => this.props.cambio({}, false)}
            value="Cerrar Sesion"
          />
        </div>

        <div className="row m-4">
          <DocViewer
            documents={this.props.Archivos}
            pluginRenderers={DocViewerRenderers}
          />
        </div>
      </div>
    );
  }
}

export default Reportes;
