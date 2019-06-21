package main

import (
	"flag"
	"fmt"
	logging "log"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

var (
	log = logging.New(os.Stderr, "HtmlNoJS", 0)
)

type NodeAction func(*goquery.Document, *goquery.Selection) *goquery.Document

func main() {
	flag.Parse()
	args := flag.Args()
	infname := args[0]

	docfile, err := os.Open(infname)
	if err != nil {
		log.Fatalln(err)
	}

	odoc, err := goquery.NewDocumentFromReader(docfile)
	if err != nil {
		log.Fatalln(err)
	}

	doc := JavaScriptCleaner(odoc)

	html, err := doc.Html()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(html)
}

func JavaScriptCleaner(odoc *goquery.Document) *goquery.Document {
	doc := goquery.CloneDocument(odoc)

	doc = RemoveScriptTags(doc)

	allTagActions := []NodeAction{StripInvalidSrc}
	allTagActions = append(allTagActions, StripAttribsByPrefixActionFactory("on")...)
	allTagActions = append(allTagActions, StripAttrsActionFactory([]string{"seekSegmentTime"})...)
	ProcessNodes(doc, []string{"*"}, allTagActions)

	stripSrcAction := StripAttrsActionFactory([]string{"src"})
	ProcessNodes(doc, []string{"iframe", "frame"}, stripSrcAction)

	stripRefreshAction := StripAttrsActionFactory([]string{"http-equiv"})
	ProcessNodes(doc, []string{"meta"}, stripRefreshAction)

	bodyBackGround := StripAttrsActionFactory([]string{"background"})
	ProcessNodes(doc, []string{"body"}, bodyBackGround)

	return doc
}

//TODO
// OBJECT elements?

var StripInvalidSrc NodeAction = NodeAction(stripInvalidSrc)

func stripInvalidSrc(doc *goquery.Document, sel *goquery.Selection) *goquery.Document {
	for _, node := range sel.Nodes {
		for ia, attr := range node.Attr {
			if attr.Key == "src" || attr.Key == "dynsrc" || attr.Key == "lowsrc" {
				if strings.Index(attr.Val, "https://") != 0 && strings.Index(attr.Val, "http://") != 0 && strings.Index(attr.Val, "/") != 0 {
					log.Printf("[-] Invalid @src '%s' in '%s'\n", attr.Val, node.Data)
					node.Attr[ia].Val = ""
				}
			}
		}
	}
	return doc
}

func StripAttrsActionFactory(attrNames []string) (handlers []NodeAction) {
	for _, attr := range attrNames {
		handlers = append(handlers, func(doc *goquery.Document, sel *goquery.Selection) *goquery.Document {
			// for _, n := range sel.Nodes {
			// 	log.Printf("[-]  @'%s' from '%s'\n", attr, n.Data)
			// }
			sel.RemoveAttr(attr)
			return doc
		})
	}
	return handlers
}

func StripAttribsByPrefixActionFactory(prefix string) (handlers []NodeAction) {
	return append(handlers, NodeAction(func(doc *goquery.Document, sel *goquery.Selection) *goquery.Document {
		for _, node := range sel.Nodes {
			newAttribs := []html.Attribute{}
			stripped := []string{}
			for _, attr := range node.Attr {
				aName := attr.Key
				if strings.Index(aName, "on") != 0 {
					newAttribs = append(newAttribs, attr)
				} else {
					stripped = append(stripped, attr.Key)
				}
			}
			if len(stripped) > 0 {
				log.Printf("[-] attrByPrefix '%s' from %s%v\n", prefix, node.Data, stripped)
			}
			node.Attr = newAttribs
		}

		return doc
	}))
}

func ProcessNodes(doc *goquery.Document, tagList []string, actions []NodeAction) *goquery.Document {
	for _, tag := range tagList {
		sel := doc.Find(tag)
		for _, f := range actions {
			doc = f(doc, sel)
		}
	}
	return doc
}

func RemoveScriptTags(doc *goquery.Document) *goquery.Document {
	s := doc.Find("script")
	log.Println("[-] Remove", len(s.Nodes), "script nodes.")
	for i := range s.Nodes {
		nodeSel := s.Eq(i)
		nodeSel.Remove()
	}
	return doc
}

func exampleNodeIterate(doc *goquery.Document) (*goquery.Document, error) {
	s := doc.Find("script")
	log.Printf("%+v\n", s)
	log.Printf("%+v\n", s.Nodes)
	for i, node := range s.Nodes {
		log.Printf("%+v\n", node.Data)
		nodeSel := s.Eq(i)
		d := nodeSel.Text()
		// d := nodeSel.Contents()
		log.Println(d)
	}
	return doc, nil
}
