package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/opds-community/libopds2-go/opds2"
	"github.com/readium/go-toolkit/pkg/parser"
	"github.com/readium/go-toolkit/pkg/pub"
	"github.com/urfave/negroni"
)

type PublicationServer struct {
	config          ServerConfig
	currentBookList []currentBook
	zipMutex        sync.Mutex
	feed            *opds2.Feed
}

func NewPublicationServer(config ServerConfig) *PublicationServer {
	return &PublicationServer{
		config: config,
		feed:   new(opds2.Feed),
	}
}

func (s *PublicationServer) Init() http.Handler {
	go s.createOPDSFeed()

	n := negroni.Classic()
	n.Use(negroni.NewStatic(http.Dir(s.config.StaticPath)))
	n.UseHandler(s.bookHandler(false))
	return n
}

func (s *PublicationServer) bookHandler(test bool) http.Handler {
	serv := mux.NewRouter()

	serv.HandleFunc("/{filename}/manifest.json", s.getManifest)
	// serv.HandleFunc("/{filename}/search", s.search)
	// serv.HandleFunc("/{filename}/media-overlay", s.mediaOverlay)
	serv.HandleFunc("/{filename}/{asset:.*}", s.getAsset)
	serv.HandleFunc("/publications.json", s.opdsFeedHandler)

	return serv
}

func (s *PublicationServer) getManifest(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	filename := vars["filename"]

	publication, err := s.getPublication(filename, req)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	j, _ := json.Marshal(publication)

	var identJSON bytes.Buffer

	json.Indent(&identJSON, j, "", " ")
	w.Header().Set("Content-Type", "application/webpub+json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*") // TODO replace with CORS middleware

	hashJSONRaw := sha256.Sum256(identJSON.Bytes())
	hashJSON := base64.RawStdEncoding.EncodeToString(hashJSONRaw[:])

	if match := req.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, hashJSON) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}
	w.Header().Set("Etag", hashJSON)

	links := publication.GetPreFetchResources()
	if len(links) > 0 {
		prefetch := ""
		for _, l := range links {
			prefetch = prefetch + "<" + l.Href + ">;" + "rel=prefetch,"
		}
		w.Header().Set("Link", prefetch)
	}

	identJSON.WriteTo(w)
}

func (s *PublicationServer) getAsset(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(501)

	/*
		vars := mux.Vars(req)
		assetname := vars["asset"]

		publication, err := s.getPublication(vars["filename"], req)
		if err != nil {
			w.WriteHeader(500)
			return
		}

		epubReader, mediaType, err := fetcher.Fetch(publication, assetname)
		if err != nil {
			if err.Error() == "missing or bad key" {
				w.WriteHeader(401)
				return
			}
			w.WriteHeader(404)
			return
		}

		w.Header().Set("Content-Type", mediaType)
		w.Header().Set("Access-Control-Allow-Origin", "*") // TODO replace with CORS middleware
		w.Header().Set("Cache-Control", "public,max-age=86400")
		http.ServeContent(w, req, assetname, time.Time{}, epubReader)
	*/
}

/*
func (s *PublicationServer) search(w http.ResponseWriter, req *http.Request) {
	var returnJSON bytes.Buffer
	vars := mux.Vars(req)

	publication, err := s.getPublication(vars["filename"], req)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	searchTerm := req.URL.Query().Get("query")
	searchReturn, err := searcher.Search(*publication, searchTerm)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	j, _ := json.Marshal(searchReturn)
	json.Indent(&returnJSON, j, "", "  ")
	returnJSON.WriteTo(w)
}

func (s *PublicationServer) mediaOverlay(w http.ResponseWriter, req *http.Request) {
	var returnJSON bytes.Buffer
	var media []pub.MediaOverlayNode

	vars := mux.Vars(req)
	var mediaOverlay struct {
		MediaOverlay []pub.MediaOverlayNode `json:"media-overlay"`
	}

	publication, err := s.getPublication(vars["filename"], req)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	resource := req.URL.Query().Get("resource")
	if resource == "" {
		media = publication.FindAllMediaOverlay()

	} else {
		media = publication.FindMediaOverlayByHref(resource)
	}

	mediaOverlay.MediaOverlay = media
	j, _ := json.Marshal(mediaOverlay)
	json.Indent(&returnJSON, j, "", "  ")
	w.Header().Set("Content-Type", "application/vnd.readium.mo+json")
	returnJSON.WriteTo(w)
}*/

func (s *PublicationServer) getPublication(filename string, req *http.Request) (*pub.Manifest, error) {
	return nil, errors.New("not implemented")

	/*
		var current currentBook

		for _, book := range s.currentBookList {
			if filename == book.filename {
				current = book
			}
		}

		if current.filename == "" {
			manifestURL := "http://" + req.Host + "/" + filename + "/manifest.json"
			filenamePath, _ := base64.StdEncoding.DecodeString(filename)

			publication, err := parser.Parse(string(filenamePath))
			hasMediaOverlay := false
			for _, l := range publication.ReadingOrder {
				if l.Properties != nil && l.Properties.MediaOverlay != "" {
					hasMediaOverlay = true
				}
			}

			if err != nil {
				return &pub.Manifest{}, err
			}

			publication.AddLink("application/webpub+json", []string{"self"}, manifestURL, false)
			if hasMediaOverlay {
				publication.AddLink("application/vnd.readium.mo+json", []string{"media-overlay"}, "http://"+req.Host+"/"+filename+"/media-overlay?resource={path}", true)
			}
			// if searcher.CanBeSearch(publication) {
			// 	publication.AddLink("", []string{"search"}, "http://"+req.Host+"/"+filename+"/search?query={searchTerms}", true)
			// }
			current = currentBook{filename: filename, publication: publication, timestamp: time.Now(), indexed: false}
			s.currentBookList = append(s.currentBookList, current)
			// if searcher.CanBeSearch(publication) {
			// 	go indexBook(publication)
			// }
			return &publication, nil
		}
		return &current.publication, nil
	*/
	// if searcher.CanBeSearch(publication) {
	// 	go indexBook(publication)
	// }
}

func (s *PublicationServer) updatePublication(publicaton pub.Manifest, filename string) {
	for i, book := range s.currentBookList {
		if filename == book.filename {
			s.currentBookList[i].publication = publicaton
		}
	}

}

func (s *PublicationServer) createOPDSFeed() {
	t := time.Now()
	println(s.config.PublicationPath)
	files, err := ioutil.ReadDir(s.config.PublicationPath)
	if err != nil {
		return
	}
	for _, f := range files {
		pubPath := path.Join(s.config.PublicationPath, f.Name())
		pub, errParse := parser.Parse(pubPath)
		if errParse == nil {
			filename := base64.StdEncoding.EncodeToString([]byte(pubPath))
			baseURL := "http://" + s.config.Bind + "/" + filename + "/" // TODO remove hardcoded scheme
			AddPublicationToFeed(s.feed, pub, baseURL)
		}
	}
	if len(s.feed.Publications) > 0 {
		s.feed.Context = []string{"http://opds-spec.org/opds.jsonld"}
		l := opds2.Link{}
		l.Href = "http://" + s.config.Bind + "/publications.json" // TODO remove hardcoded scheme
		l.Rel = []string{"self"}
		l.TypeLink = "application/opds+json"
		s.feed.Links = append(s.feed.Links, l)
		s.feed.Metadata.Modified = &t
		s.feed.Metadata.RDFType = "http://schema.org/DataFeed"
		s.feed.Metadata.NumberOfItems = len(s.feed.Publications)
		s.feed.Metadata.Title = "Readium 2 OPDS 2.0 Feed"
	}
}

func (s *PublicationServer) opdsFeedHandler(w http.ResponseWriter, req *http.Request) {
	j, _ := json.Marshal(s.feed)

	var identJSON bytes.Buffer

	json.Indent(&identJSON, j, "", " ")
	w.Header().Set("Content-Type", "application/opds+json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*") // TODO replace with CORS middleware

	hashJSONRaw := sha256.Sum256(identJSON.Bytes())
	hashJSON := base64.RawStdEncoding.EncodeToString(hashJSONRaw[:])

	if match := req.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, hashJSON) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}
	w.Header().Set("Etag", hashJSON)

	identJSON.WriteTo(w)
}
