/** @type {import('tailwindcss').Config} */
module.exports = {
  content: {
    relative: true,
    files: [
      "../internal/**/*.{html,templ,txt}",
      "../internal/**/*.*.{html,templ,txt}",
    ],
  },
  theme: {
    extend: {
      colors: {
        "base-neutral-dark": "#171c1b",
      },
    },
  },
  daisyui: {
    themes: ["forest"],
  },
  // todo: check other plugins: forms, typography, animations, fluid-css
  plugins: [require("daisyui")],
};
// TODO: try to replace parcel with farm or rspack
