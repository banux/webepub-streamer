package epub

import (
	"strconv"

	"github.com/antchfx/xmlquery"
	"github.com/pkg/errors"
	"github.com/readium/go-toolkit/pkg/manifest"
	"github.com/readium/go-toolkit/pkg/util"
)

type PackageDocument struct {
	Path               string
	EPUBVersion        float64
	uniqueIdentifierID string
	metadata           EPUBMetadata
	Manifest           []Item
	Spine              Spine
}

func ParsePackageDocument(document *xmlquery.Node, filePath string) (*PackageDocument, error) {
	pkg := document.SelectElement("/package")
	packagePrefixes := parsePrefixes(pkg.SelectAttr("prefix"))
	prefixMap := make(map[string]string)
	for k, v := range PackageReservedPrefixes {
		prefixMap[k] = v
	}
	for k, v := range packagePrefixes {
		prefixMap[k] = v
	}

	// Version
	epubVersion := 1.2
	rv := pkg.SelectAttr("version")
	if rv != "" {
		ev, err := strconv.ParseFloat(rv, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed parsing package version")
		}
		epubVersion = ev
	}

	metadata := NewMetadataParser(epubVersion, prefixMap).Parse(document, filePath)
	if metadata == nil {
		return nil, errors.New("failed parsing package metadata")
	}
	manifestElement := pkg.SelectElement(
		"/*[namespace-uri()='" + NamespaceOPF + "' and local-name()='manifest']",
	)
	if manifestElement == nil {
		return nil, errors.New("package manifest not found")
	}
	spineElement := pkg.SelectElement("/*[namespace-uri()='" + NamespaceOPF + "' and local-name()='spine']")
	if spineElement == nil {
		return nil, errors.New("package spine not found")
	}

	mels := manifestElement.SelectElements("/*[namespace-uri()='" + NamespaceOPF + "' and local-name()='item']")
	manifest := make([]Item, 0, len(mels))
	for _, mel := range mels {
		item := ParseItem(mel, filePath, prefixMap)
		if item == nil {
			// return nil, errors.New("failed parsing package manifest item at index " + strconv.Itoa(i))
			continue
		}
		manifest = append(manifest, *item)
	}

	return &PackageDocument{
		Path:               filePath,
		EPUBVersion:        epubVersion,
		uniqueIdentifierID: pkg.SelectAttr("unique-identifier"),
		metadata:           *metadata,
		Manifest:           manifest,
		Spine:              ParseSpine(spineElement, prefixMap, epubVersion),
	}, nil

}

type Item struct {
	Href         string
	ID           string
	fallback     string
	mediaOverlay string
	MediaType    string
	Properties   []string
}

func ParseItem(element *xmlquery.Node, filePath string, prefixMap map[string]string) *Item {
	rawHref := element.SelectAttr("href")
	if rawHref == "" {
		return nil
	}
	href, err := util.NewHREF(rawHref, filePath).String()
	if err != nil {
		return nil
	}
	item := &Item{
		Href:         href,
		ID:           element.SelectAttr("id"),
		fallback:     element.SelectAttr("fallback"),
		mediaOverlay: element.SelectAttr("media-overlay"),
		MediaType:    element.SelectAttr("media-type"),
	}
	pp := parseProperties(element.SelectAttr("properties"))
	if len(pp) > 0 {
		item.Properties = make([]string, 0, len(pp))
		for _, prop := range parseProperties(element.SelectAttr("properties")) {
			if prop == "" {
				continue
			}
			item.Properties = append(item.Properties, resolveProperty(prop, prefixMap, DefaultVocabItem))
		}
	}
	return item
}

type Spine struct {
	itemrefs  []ItemRef
	direction manifest.ReadingProgression
	TOC       string
}

func ParseSpine(element *xmlquery.Node, prefixMap map[string]string, epubVersion float64) Spine {
	itemrefs := make([]ItemRef, 0)
	for _, itemref := range element.SelectElements(
		"/*[namespace-uri()='" + NamespaceOPF + "' and local-name()='itemref']",
	) {
		itemref := ParseItemRef(itemref, prefixMap)
		if itemref == nil {
			continue
		}
		itemrefs = append(itemrefs, *itemref)
	}

	pageProgressionDiretion := manifest.Auto
	switch element.SelectAttr("page-progression-direction") {
	case "ltr":
		pageProgressionDiretion = manifest.LTR
	case "rtl":
		pageProgressionDiretion = manifest.RTL
	}

	ncx := ""
	if epubVersion < 3.0 {
		ncx = element.SelectAttr("toc")
	}

	return Spine{
		itemrefs:  itemrefs,
		direction: pageProgressionDiretion,
		TOC:       ncx,
	}
}

type ItemRef struct {
	idref      string
	linear     bool
	properties []string
}

func ParseItemRef(element *xmlquery.Node, prefixMap map[string]string) *ItemRef {
	idref := element.SelectAttr("idref")
	if idref == "" {
		return nil
	}

	pp := parseProperties(element.SelectAttr("properties"))
	properties := make([]string, 0, len(pp))
	for _, prop := range parseProperties(element.SelectAttr("properties")) {
		if prop == "" {
			continue
		}
		properties = append(properties, resolveProperty(prop, prefixMap, DefaultVocabItemref))
	}

	return &ItemRef{
		idref:      idref,
		linear:     element.SelectAttr("linear") != "no",
		properties: properties,
	}
}
