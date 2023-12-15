/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./internal/component/**/*.templ",
    "./node_modules/flowbite/**/*.js",
  ],
  plugins: [require("flowbite/plugin")],
};
