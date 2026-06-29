// Tiny Go syntax highlighter for static snippets used by the home page.
export function highlightGo(src: string): string {
	const esc = (s: string) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

	const token =
		/("(?:\\.|[^"\\\n])*"|`[^`]*`)|(\/\/[^\n]*)|\b(package|import|func|return|if|else|for|range|var|const|type|struct|interface|map|chan|go|defer|switch|case|break|continue|nil|true|false)\b|\b([A-Z][A-Za-z0-9_]*)\b|([a-zA-Z_][a-zA-Z0-9_]*)(\()/g;

	return esc(src).replace(token, (match, stringLiteral, comment, keyword, typeName, functionName, openParen) => {
		if (stringLiteral) {
			return `<span class="s">${stringLiteral}</span>`;
		}

		if (comment) {
			return `<span class="c">${comment}</span>`;
		}

		if (keyword) {
			return `<span class="k">${keyword}</span>`;
		}

		if (typeName) {
			return `<span class="t">${typeName}</span>`;
		}

		if (functionName) {
			return `<span class="fn">${functionName}</span>${openParen}`;
		}

		return match;
	});
}
