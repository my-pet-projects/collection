/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["internal/component/*.templ"],
  theme: {
    extend: {},
  },
  plugins: [require("@tailwindcss/forms")],
};
