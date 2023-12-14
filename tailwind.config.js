/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["internal/component/*.templ", "./node_modules/flowbite/**/*.js"],
  theme: {
    extend: {},
  },
  plugins: [require("@tailwindcss/forms"), require("flowbite/plugin")],
};
