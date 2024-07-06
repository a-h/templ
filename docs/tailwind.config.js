const disabledCss = {
  "code::before": false,
  "code::after": false,
  "blockquote p:first-of-type::before": false,
  "blockquote p:last-of-type::after": false,
  pre: false,
  code: false,
  "pre code": false,
  "code::before": false,
  "code::after": false,
};

const defaultCss = {
  ...disabledCss,
  maxWidth: "96ch",
};

/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: "class",
  content: ["./static/*.js", "./src/components/*.templ"],
  safelist: ["anchor", "note", "tip", "info", "warning", "critical", "caution"],
  theme: {
    extend: {
      typography: {
        DEFAULT: { css: defaultCss },
        sm: { css: disabledCss },
        lg: { css: disabledCss },
        xl: { css: disabledCss },
        "2xl": { css: disabledCss },
      },
      colors: {
        "c-cyan": {
          100: "#00B3C7", // --ifm-color-primary-lightest: #00B3C7;
          200: "#00A5B8", // --ifm-color-primary-lighter: #00A5B8;
          300: "#0093A3", // --ifm-color-primary-light: #0093A3;
          400: "#008391", // --ifm-color-primary: #008391;
          500: "#007380", // --ifm-color-primary-dark: #007380;
          600: "#006570", // --ifm-color-primary-darker: #006570;
          700: "#005761", // --ifm-color-primary-darkest: #005761;
        },
        "c-yellow": {
          100: "#F7E078", // --ifm-color-primary-lightest: #F7E078;
          200: "#F1D65F", // --ifm-color-primary-lighter: #F1D65F;
          300: "#E7C946", // --ifm-color-primary-light: #E7C946;
          400: "#DBBC30", // --ifm-color-primary: #DBBC30;
          500: "#D0B125", // --ifm-color-primary-dark: #D0B125;
          600: "#BA9E21", // --ifm-color-primary-darker: #BA9E21;
          700: "#A0881C", // --ifm-color-primary-darkest: #A0881C;
        },
        "c-teal": "#003238", // --ifm-footer-background-color: #003238;
        "c-highlighted-code-light": "rgba(0, 0, 0, 0.1)", // --docusaurus-highlighted-code-line-bg: rgba(0, 0, 0, 0.1);
        "c-highlighted-code-dark": "rgba(0, 0, 0, 0.3)", // --docusaurus-highlighted-code-line-bg: rgba(0, 0, 0, 0.3);
      },
    },
  },
  plugins: [require("@tailwindcss/typography")],
};
