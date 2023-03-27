package content

import (
	"strings"

	"github.com/readium/go-toolkit/pkg/content/element"
	"github.com/readium/go-toolkit/pkg/content/iterator"
)

type Content interface {
	Text(separator *string) string // Extracts the full raw text, or returns null if no text content can be found.
	Iterator() iterator.Iterator   // Creates a new iterator for this content.
	Elements() []element.Element   // Returns all the elements as a list.
}

// Extracts the full raw text, or returns null if no text content can be found.
func ContentText(content Content, separator *string) string {
	sep := "\n"
	if separator != nil {
		sep = *separator
	}
	var sb strings.Builder
	for _, el := range content.Elements() {
		if txel, ok := el.(element.TextualElement); ok {
			txt := txel.Text()
			if txt != "" {
				sb.WriteString(txel.Text())
				sb.WriteString(sep)
			}
		}
	}
	return strings.TrimSuffix(sb.String(), sep)
}

func ContentElements(content Content) []element.Element {
	var elements []element.Element
	for content.Iterator().HasNext() {
		elements = append(elements, content.Iterator().Next())
	}
	return elements
}
