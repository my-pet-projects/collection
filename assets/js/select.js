customElements.define(
  "app-select",
  class extends HTMLElement {
    constructor() {
      super(...arguments);
      this._data = [];
    }
    connectedCallback() {
      const name = this.getAttribute("name");
      const placeholder = this.getAttribute("data-placeholder");
      const eventName = `${name}-change-choice`;
      const searchFields = [
        "label",
        "value",
        "customProperties.searchableValue",
      ];
      this._data = JSON.parse(this.getAttribute("data-items"));
      this.innerHTML = `<select id="${name}" name="${name}" class="choices"></select>`;

      const choicesEl = this.querySelector(".choices");
      if (choicesEl == null) {
        return;
      }

      choicesEl.addEventListener(
        "choice",
        function (event) {
          const changeEvent = new CustomEvent(eventName, {
            detail: { choice: event.detail.choice },
          });
          document.dispatchEvent(changeEvent);
        },
        false
      );
      const choices = new Choices(choicesEl, {
        placeholder: placeholder != null && placeholder.length > 0,
        placeholderValue: placeholder,
        searchPlaceholderValue: placeholder,
        itemSelectText: "",
        searchFields,
        searchEnabled: true,
        searchChoices: true,
        searchResultLimit: 20,
        fuseOptions: {
          keys: searchFields,
        },
        allowHTML: true,
        sorter: function (a, b) {
          const akey = a["sortValue"] ?? a["label"] ?? "";
          const bkey = b["sortValue"] ?? b["label"] ?? "";
          if (akey.length === 0 || bkey.length == 0) {
            return 0;
          }
          return akey > bkey ? 1 : -1;
        },
        choices: this._data,
        renderChoiceLimit: 300,
      });
      this._choices = choices;

      const selectedValueEl = document.getElementById(`selected-${name}`);
      const selectedValue = selectedValueEl ? selectedValueEl.value : "";
      if (selectedValue) {
        choices.setChoiceByValue(selectedValue.toLowerCase());
        // for some reason initial event is not being caught by hx-trigger attribute of the web-component,
        // so it is kind of useless here, but let's dispatch it anyways.
        const changeEvent = new CustomEvent(eventName, {
          detail: { choice: selectedValue },
        });
        document.dispatchEvent(changeEvent);
      }
    }
    attributeChangedCallback(name, _, newValue) {
      console.log("attributeChangedCallback", name, newValue);
      if (name !== "data-items" || this._choices == null) {
        return;
      }
      this._data = JSON.parse(newValue);
      this._choices.clearChoices();
      this._choices.setChoices(this._data);
    }
  }
);
