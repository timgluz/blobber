window.onload = function () {
  //<editor-fold desc="Changeable Configuration Block">
  var currentHost = window.location.host;
  var protocol = window.location.protocol;
  var specUrl = protocol + "//" + currentHost + "/static/openapi.yaml";
  // the following lines will be replaced by docker/configurator, when it runs in a docker-container
  window.ui = SwaggerUIBundle({
    url: specUrl,
    dom_id: "#swagger-ui",
    deepLinking: true,
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
    plugins: [SwaggerUIBundle.plugins.DownloadUrl],
    layout: "StandaloneLayout",
  });

  //</editor-fold>
};
