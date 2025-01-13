const plugin = require("tailwindcss/plugin");

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: {
    relative: true,
    files: [
      "../internal/**/*.{html,templ,txt}",
      "../internal/**/*.*.{html,templ,txt}",
    ],
  },
  safelist: [
    "datastar-swapping", // Force this class to always be included
  ],
  theme: {
    extend: {
      colors: {
        "base-neutral-dark": "#171c1b",
      },
    },
  },
  daisyui: {
    themes: ["forest"],
    logs: false,
  },
  // todo: check other plugins: forms, typography, animations, fluid-css
  plugins: [
    require("daisyui"),
    plugin(function ({ addUtilities }) {
      addUtilities({
        ".datastar-swapping": {
          opacity: "0",
          transition: "opacity 1s ease-out",
        },
      });
    }),
  ],
};
// TODO: try to replace parcel with farm or rspack
