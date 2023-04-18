import siteConfig from '@generated/docusaurus.config';
export default function prismIncludeLanguages(PrismObject) {
    const {
        themeConfig: {
            prism
        },
    } = siteConfig;
    const {
        additionalLanguages
    } = prism;
    // Prism components work on the Prism instance on the window, while prism-
    // react-renderer uses its own Prism instance. We temporarily mount the
    // instance onto window, import components to enhance it, then remove it to
    // avoid polluting global namespace.
    // You can mutate PrismObject: registering plugins, deleting languages... As
    // long as you don't re-assign it
    globalThis.Prism = PrismObject;
    additionalLanguages.forEach((lang) => {
        // eslint-disable-next-line global-require, import/no-dynamic-require
        require(`prismjs/components/prism-${lang}`);
    });
    var go = globalThis.Prism.languages.extend('go', {
        'keyword': /\b(?:break|case|chan|const|continue|default|defer|else|fallthrough|for|func|go(?:to)?|if|import|interface|map|package|range|return|select|struct|switch|type|var|templ|css|script)\b/,
    });

    var space = /(?:\s|\/\/.*(?!.)|\/\*(?:[^*]|\*(?!\/))\*\/)/.source;
    var braces = /(?:\{(?:\{(?:\{[^{}]*\}|[^{}])*\}|[^{}])*\})/.source;
    var spread = /(?:\{<S>*\.{3}(?:[^{}]|<BRACES>)*\})/.source;

    /**
     * @param {string} source
     * @param {string} [flags]
     */
    function re(source, flags) {
        source = source
            .replace(/<S>/g, function() {
                return space;
            })
            .replace(/<BRACES>/g, function() {
                return braces;
            })
            .replace(/<SPREAD>/g, function() {
                return spread;
            });
        return RegExp(source, flags);
    }

    spread = re(spread).source;


    globalThis.Prism.languages.templ = globalThis.Prism.languages.extend('markup', go);
    globalThis.Prism.languages.templ.tag.pattern = re(
        /<\/?(?:[\w.:-]+(?:<S>+(?:[\w.:$-]+(?:=(?:"(?:\\[\s\S]|[^\\"])*"|'(?:\\[\s\S]|[^\\'])*'|[^\s{'"/>=]+|<BRACES>))?|<SPREAD>))*<S>*\/?)?>/.source
    );

    globalThis.Prism.languages.templ.tag.inside['tag'].pattern = /^<\/?[^\s>\/]*/;
    globalThis.Prism.languages.templ.tag.inside['attr-value'].pattern = /=(?!\{)(?:"(?:\\[\s\S]|[^\\"])*"|'(?:\\[\s\S]|[^\\'])*'|[^\s'">]+)/;
    globalThis.Prism.languages.templ.tag.inside['tag'].inside['class-name'] = /^[A-Z]\w*(?:\.[A-Z]\w*)*$/;
    globalThis.Prism.languages.templ.tag.inside['comment'] = go['comment'];

    globalThis.Prism.languages.insertBefore('inside', 'attr-name', {
        'spread': {
            pattern: re(/<SPREAD>/.source),
            inside: globalThis.Prism.languages.templ
        }
    }, globalThis.Prism.languages.templ.tag);

    globalThis.Prism.languages.insertBefore('inside', 'special-attr', {
        'script': {
            // Allow for two levels of nesting
            pattern: re(/=<BRACES>/.source),
            alias: 'language-go',
            inside: {
                'script-punctuation': {
                    pattern: /^=(?=\{)/,
                    alias: 'punctuation'
                },
                rest: globalThis.Prism.languages.templ
            },
        }
    }, globalThis.Prism.languages.templ.tag);

    // The following will handle plain text inside tags
    var stringifyToken = function(token) {
        if (!token) {
            return '';
        }
        if (typeof token === 'string') {
            return token;
        }
        if (typeof token.content === 'string') {
            return token.content;
        }
        return token.content.map(stringifyToken).join('');
    };

    var walkTokens = function(tokens) {
        var openedTags = [];
        for (var i = 0; i < tokens.length; i++) {
            var token = tokens[i];
            var notTagNorBrace = false;

            if (typeof token !== 'string') {
                if (token.type === 'tag' && token.content[0] && token.content[0].type === 'tag') {
                    // We found a tag, now find its kind

                    if (token.content[0].content[0].content === '</') {
                        // Closing tag
                        if (openedTags.length > 0 && openedTags[openedTags.length - 1].tagName === stringifyToken(token.content[0].content[1])) {
                            // Pop matching opening tag
                            openedTags.pop();
                        }
                    } else {
                        if (token.content[token.content.length - 1].content === '/>') {
                            // Autoclosed tag, ignore
                        } else {
                            // Opening tag
                            openedTags.push({
                                tagName: stringifyToken(token.content[0].content[1]),
                                openedBraces: 0
                            });
                        }
                    }
                } else if (openedTags.length > 0 && token.type === 'punctuation' && token.content === '{') {

                    // Here we might have entered a templ context inside a tag
                    openedTags[openedTags.length - 1].openedBraces++;

                } else if (openedTags.length > 0 && openedTags[openedTags.length - 1].openedBraces > 0 && token.type === 'punctuation' && token.content === '}') {

                    // Here we might have left a templ context inside a tag
                    openedTags[openedTags.length - 1].openedBraces--;

                } else {
                    notTagNorBrace = true;
                }
            }
            if (notTagNorBrace || typeof token === 'string') {
                if (openedTags.length > 0 && openedTags[openedTags.length - 1].openedBraces === 0) {
                    // Here we are inside a tag, and not inside a templ context.
                    // That's plain text: drop any tokens matched.
                    var plainText = stringifyToken(token);

                    // And merge text with adjacent text
                    if (i < tokens.length - 1 && (typeof tokens[i + 1] === 'string' || tokens[i + 1].type === 'plain-text')) {
                        plainText += stringifyToken(tokens[i + 1]);
                        tokens.splice(i + 1, 1);
                    }
                    if (i > 0 && (typeof tokens[i - 1] === 'string' || tokens[i - 1].type === 'plain-text')) {
                        plainText = stringifyToken(tokens[i - 1]) + plainText;
                        tokens.splice(i - 1, 1);
                        i--;
                    }

                    tokens[i] = new globalThis.Prism.Token('plain-text', plainText, null, plainText);
                }
            }

            if (token.content && typeof token.content !== 'string') {
                walkTokens(token.content);
            }
        }
    };

    globalThis.Prism.hooks.add('after-tokenize', function(env) {
        if (env.language !== 'templ') {
            return;
        }
        walkTokens(env.tokens);
    });
}
