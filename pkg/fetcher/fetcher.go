package fetcher

import "github.com/readium/go-toolkit/pkg/pub"

type Fetcher interface {

	/**
	 * Known resources available in the medium, such as file paths on the file system
	 * or entries in a ZIP archive. This list is not exhaustive, and additional
	 * unknown resources might be reachable.
	 *
	 * If the medium has an inherent resource order, it should be followed.
	 * Otherwise, HREFs are sorted alphabetically.
	 */
	Links() []pub.Link

	/**
	 * Returns the [Resource] at the given [link]'s HREF.
	 *
	 * A [Resource] is always returned, since for some cases we can't know if it exists before
	 * actually fetching it, such as HTTP. Therefore, errors are handled at the Resource level.
	 */
	Get(link pub.Link) *Resource

	// Closes this object and releases any resources associated with it.
	// If the object is already closed then invoking this method has no effect.
	Close()
}

func Get(f Fetcher, href string) *Resource {
	return f.Get(pub.Link{
		Href: href,
	})
}

// A [Fetcher] providing no resources at all.
type EmptyFetcher struct{}

func (f EmptyFetcher) Links() []pub.Link {
	return []pub.Link{}
}

func (f EmptyFetcher) Get(link pub.Link) Resource {
	return NewFailureResource(link, NotFound(nil))
}

func (f EmptyFetcher) Close() {}
