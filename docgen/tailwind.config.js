const disabledCss = {
    'code::before': false,
    'code::after': false,
    'blockquote p:first-of-type::before': false,
    'blockquote p:last-of-type::after': false,
    pre: false,
    code: false,
    'pre code': false,
    'code::before': false,
    'code::after': false,
}

const defaultCss = {
    ...disabledCss,
    maxWidth: '96ch',
}

/** @type {import('tailwindcss').Config} */
module.exports = {
    darkMode: 'class',
    content: [
        "./static/*.js",
        "./*.templ"
    ],
    safelist: [
        "anchor",
        "note",
        "tip",
        "info",
        "warning",
        "critical",
        "caution"
    ],
    theme: {
        extend: {
            typography: {
                DEFAULT: { css: defaultCss },
                sm: { css: disabledCss },
                lg: { css: disabledCss },
                xl: { css: disabledCss },
                '2xl': { css: disabledCss },
            },
        },
    },
    plugins: [
        require('@tailwindcss/typography'),
    ],
}
