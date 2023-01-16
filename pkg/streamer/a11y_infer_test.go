package streamer

import (
	"testing"

	"github.com/readium/go-toolkit/pkg/manifest"
	"github.com/readium/go-toolkit/pkg/mediatype"
	"github.com/stretchr/testify/assert"
)

func TestReturnsAugmentedA11yMetadataWithInference(t *testing.T) {
	a11y := manifest.NewA11y()
	a11y.ConformsTo = []manifest.A11yProfile{"unknown"}

	m := manifest.Manifest{
		Metadata: manifest.Metadata{
			Accessibility: &a11y,
		},
		ReadingOrder: []manifest.Link{
			newLink(mediatype.HTML, "html"),
		},
	}

	inferreddA11y := a11y
	inferreddA11y.AccessModesSufficient = [][]manifest.A11yPrimaryAccessMode{{manifest.A11yPrimaryAccessModeTextual}}

	res := inferA11yMetadataFromManifest(m)
	assert.Equal(t, &inferreddA11y, res)

	// Original manifest should not be modified.
	assert.Equal(t, &a11y, m.Metadata.Accessibility)
}

func newLink(mt mediatype.MediaType, extension string) manifest.Link {
	return manifest.Link{
		Href: "file." + extension,
		Type: mt.String(),
	}
}

// If the publication contains a reference to an audio or video resource
// (inspect "resources" and "readingOrder" in RWPM).
func TestInferAuditoryAccessMode(t *testing.T) {
	assertAccessMode(t, manifest.A11yAccessModeAuditory, "mp3", mediatype.MP3)
	assertAccessMode(t, manifest.A11yAccessModeAuditory, "mpeg", mediatype.MPEG)
}

// If the publications contains a reference to an image or a video resource
// (inspect "resources" and "readingOrder" in RWPM)
func TestInferVisualAccessMode(t *testing.T) {
	assertAccessMode(t, manifest.A11yAccessModeVisual, "jpg", mediatype.JPEG)
	assertAccessMode(t, manifest.A11yAccessModeVisual, "mpeg", mediatype.MPEG)
}

func assertAccessMode(t *testing.T, accessMode manifest.A11yAccessMode, extension string, mt mediatype.MediaType) {
	testManifest := func(m manifest.Manifest) {
		res := inferA11yMetadataFromManifest(m)
		assert.NotNil(t, res)
		assert.Contains(t, res.AccessModes, accessMode)
	}

	link := newLink(mt, extension)

	testManifest(manifest.Manifest{
		ReadingOrder: []manifest.Link{link},
	})
	testManifest(manifest.Manifest{
		Resources: []manifest.Link{link},
	})
}

// If the publication is partially or fully accessible (WCAG A or above)
func TestInferTextualAccessModeSufficientFromProfile(t *testing.T) {
	test := func(profile manifest.A11yProfile) {
		a11y := manifest.NewA11y()
		a11y.ConformsTo = []manifest.A11yProfile{profile}
		m := manifest.Manifest{
			Metadata: manifest.Metadata{
				Accessibility: &a11y,
			},
		}
		res := inferA11yMetadataFromManifest(m)
		assert.NotNil(t, res)
		assert.Contains(t, res.AccessModesSufficient, []manifest.A11yPrimaryAccessMode{manifest.A11yPrimaryAccessModeTextual})
	}

	test(manifest.EPUBA11y10WCAG20A)
	test(manifest.EPUBA11y10WCAG20AA)
	test(manifest.EPUBA11y10WCAG20AAA)
}

// Or if the publication does not contain any image, audio or video resource
// (inspect "resources" and "readingOrder" in RWPM)
func TestInferTextualAccessModeSufficientFromLackOfMedia(t *testing.T) {
	testManifest := func(contains bool, m manifest.Manifest) {
		res := inferA11yMetadataFromManifest(m)
		assert.NotNil(t, res)
		ams := []manifest.A11yPrimaryAccessMode{manifest.A11yPrimaryAccessModeTextual}

		if contains {
			assert.Contains(t, res.AccessModesSufficient, ams)
		} else {
			assert.NotContains(t, res.AccessModesSufficient, ams)
		}
	}

	a11y := manifest.NewA11y()
	a11y.ConformsTo = []manifest.A11yProfile{"unknown"}

	testReadingOrder := func(contains bool, mt mediatype.MediaType, extension string) {
		testManifest(contains, manifest.Manifest{
			Metadata:     manifest.Metadata{Accessibility: &a11y},
			ReadingOrder: []manifest.Link{newLink(mt, extension)},
		})
	}

	testReadingOrder(true, mediatype.HTML, "html")
	testReadingOrder(false, mediatype.JPEG, "jpg")
	testReadingOrder(false, mediatype.MP3, "mp3")
	testReadingOrder(false, mediatype.MPEG, "mpeg")
	testReadingOrder(false, mediatype.PDF, "pdf")

	testResources := func(contains bool, mt mediatype.MediaType, extension string) {
		testManifest(contains, manifest.Manifest{
			Metadata:  manifest.Metadata{Accessibility: &a11y},
			Resources: []manifest.Link{newLink(mt, extension)},
		})
	}

	testResources(true, mediatype.HTML, "html")
	testResources(false, mediatype.JPEG, "jpg")
	testResources(false, mediatype.MP3, "mp3")
	testResources(false, mediatype.MPEG, "mpeg")
	testResources(false, mediatype.PDF, "pdf")
}

// If the publication contains only references to audio resources (inspect "resources" and "readingOrder" in RWPM)
func TestInferAuditoryAccessModeSufficient(t *testing.T) {
	testManifest := func(contains bool, m manifest.Manifest) {
		res := inferA11yMetadataFromManifest(m)
		assert.NotNil(t, res)
		ams := []manifest.A11yPrimaryAccessMode{manifest.A11yPrimaryAccessModeAuditory}

		if contains {
			assert.Contains(t, res.AccessModesSufficient, ams)
		} else {
			assert.NotContains(t, res.AccessModesSufficient, ams)
		}
	}

	a11y := manifest.NewA11y()
	a11y.ConformsTo = []manifest.A11yProfile{"unknown"}

	testReadingOrder := func(contains bool, links ...manifest.Link) {
		testManifest(contains, manifest.Manifest{
			Metadata:     manifest.Metadata{Accessibility: &a11y},
			ReadingOrder: links,
		})
	}

	html := newLink(mediatype.HTML, "html")
	mp3 := newLink(mediatype.MP3, "mp3")

	testReadingOrder(false, html, html)
	testReadingOrder(false, html, mp3)
	testReadingOrder(true, mp3, mp3)

	testResources := func(contains bool, links ...manifest.Link) {
		testManifest(contains, manifest.Manifest{
			Metadata:  manifest.Metadata{Accessibility: &a11y},
			Resources: links,
		})
	}

	testResources(false, html, html)
	testResources(false, html, mp3)
	testResources(true, mp3, mp3)
}

// If the publications contains a table of contents (check for the presence of
// a "toc" collection in RWPM)
func TestInferFeatureTableOfContents(t *testing.T) {
	m := manifest.Manifest{
		TableOfContents: []manifest.Link{newLink(mediatype.HTML, "html")},
	}
	assertFeature(t, m, manifest.A11yFeatureTableOfContents)
}

// If the publications contains a page list (check for the presence of a
// "pageList" collection in RWPM)
func TestInferFeaturePageList(t *testing.T) {
	m := manifest.Manifest{
		Metadata: manifest.Metadata{
			ConformsTo: []manifest.Profile{manifest.ProfileEPUB},
		},
		Subcollections: map[string][]manifest.PublicationCollection{
			"pageList": {
				manifest.PublicationCollection{
					Links: []manifest.Link{newLink(mediatype.HTML, "html")},
				},
			},
		},
		ReadingOrder: []manifest.Link{newLink(mediatype.HTML, "html")},
	}
	assertFeature(t, m, manifest.A11yFeaturePrintPageNumbers)
}

// If the publication contains any resource with MathML (check for the presence
// of the "contains" property where the value is "mathml" in "readingOrder" or
// "resources" in RWPM)
func TestInferFeatureMathML(t *testing.T) {
	link := newLink(mediatype.HTML, "html")
	link.Properties = map[string]interface{}{
		"contains": []string{"mathml"},
	}
	m := manifest.Manifest{
		Metadata: manifest.Metadata{
			ConformsTo: []manifest.Profile{manifest.ProfileEPUB},
		},
		ReadingOrder: []manifest.Link{link},
	}
	assertFeature(t, m, manifest.A11yFeatureMathML)
}

// If the publication is fully accessible (WCAG AA or above)
//
// This property should only be inferred for reflowable EPUB files as it
// doesn't apply to other formats (FXL, PDF, audiobooks, CBZ/CBR).
func TestInferFeatureDisplayTransformability(t *testing.T) {
	test := func(contains bool, profile manifest.A11yProfile, layout manifest.EPUBLayout) {
		a11y := manifest.NewA11y()
		a11y.ConformsTo = []manifest.A11yProfile{profile}

		m := manifest.Manifest{
			Metadata: manifest.Metadata{
				ConformsTo:    []manifest.Profile{manifest.ProfileEPUB},
				Accessibility: &a11y,
				Presentation: &manifest.Presentation{
					Layout: &layout,
				},
			},
			ReadingOrder: []manifest.Link{newLink(mediatype.HTML, "html")},
		}

		res := inferA11yMetadataFromManifest(m)
		assert.NotNil(t, res)
		if contains {
			assert.Contains(t, res.Features, manifest.A11yFeatureDisplayTransformability)
		} else {
			assert.NotContains(t, res.Features, manifest.A11yFeatureDisplayTransformability)
		}
	}

	test(false, manifest.EPUBA11y10WCAG20A, manifest.EPUBLayoutReflowable)
	test(true, manifest.EPUBA11y10WCAG20AA, manifest.EPUBLayoutReflowable)
	test(true, manifest.EPUBA11y10WCAG20AAA, manifest.EPUBLayoutReflowable)
	test(false, manifest.EPUBA11y10WCAG20AAA, manifest.EPUBLayoutFixed)
}

func assertFeature(t *testing.T, m manifest.Manifest, feature manifest.A11yFeature) {
	res := inferA11yMetadataFromManifest(m)
	assert.NotNil(t, res)
	assert.Contains(t, res.Features, feature)
}
