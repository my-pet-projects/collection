import {
  Uppy,
  Dashboard,
  XHRUpload,
} from "https://releases.transloadit.com/uppy/v3.27.0/uppy.min.mjs";

customElements.define(
  "app-upload",
  class extends HTMLElement {
    constructor() {
      super(...arguments);
      this._data = [];
    }
    connectedCallback() {
      this.innerHTML = `<div id="uppy"></div>`;

      const uppy = new Uppy({
        debug: true,
        meta: null,
      });

      uppy.use(Dashboard, {
        target: "#uppy",
        inline: true,
        hideRetryButton: false,
        hideCancelButton: false,
      });

      uppy.use(XHRUpload, {
        bundle: true,
        endpoint: "/workspace/images/uploads",
        fieldName: "files",
        formData: true,
        limit: 10,
        timeout: 0,
      });
    }
  }
);
