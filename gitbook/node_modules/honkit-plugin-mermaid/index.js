module.exports = {
    book: {
        assets: './assets',
        js: ['mermaid-load.js', 'mermaid.min.js'],
        css: ['mermaid.css']
    },
    blocks: {
        code: (block) => {
            const lang = block.kwargs.language;
            if (lang !== 'mermaid') {
                return block;
            }
            block.body = toSvg(block.body);
            return block;
        }
    },
    hooks: {
        "page:before": page => {
            page.content = page.content.replace(/```http request/g, "```");
            return page;
        }
    }
};
const toSvg = (source) => {
    return "<div class='mermaid'>" + source.replace(/</g, "&lt;").replace(/>/g, "&gt;") + "</div>";
}
