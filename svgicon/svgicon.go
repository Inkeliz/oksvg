package svgicon

import (
	"encoding/xml"
	"io"
	"os"

	"golang.org/x/net/html/charset"
)

// ReadIconStream reads the Icon from the given io.Reader
// This only supports a sub-set of SVG, but
// is enough to draw many icons. errMode determines if the icon ignores, errors out, or logs a warning
// if it does not handle an element found in the icon file.
func ReadIconStream(stream io.Reader, errMode ErrorMode) (*SvgIcon, error) {
	icon := &SvgIcon{defs: make(map[string][]definition), grads: make(map[string]*Gradient), Transform: Identity}
	cursor := &iconCursor{styleStack: []PathStyle{DefaultStyle}, icon: icon}
	cursor.errorMode = errMode
	decoder := xml.NewDecoder(stream)
	decoder.CharsetReader = charset.NewReaderLabel
	for {
		t, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return icon, err
		}
		// Inspect the type of the XML token
		switch se := t.(type) {
		case xml.StartElement:
			// Reads all recognized style attributes from the start element
			// and places it on top of the styleStack
			err = cursor.pushStyle(se.Attr)
			if err != nil {
				return icon, err
			}
			err = cursor.readStartElement(se)
			if err != nil {
				return icon, err
			}
		case xml.EndElement:
			// pop style
			cursor.styleStack = cursor.styleStack[:len(cursor.styleStack)-1]
			switch se.Name.Local {
			case "g":
				if cursor.inDefs {
					cursor.currentDef = append(cursor.currentDef, definition{
						Tag: "endg",
					})
				}
			case "title":
				cursor.inTitleText = false
			case "desc":
				cursor.inDescText = false
			case "defs":
				if len(cursor.currentDef) > 0 {
					cursor.icon.defs[cursor.currentDef[0].ID] = cursor.currentDef
					cursor.currentDef = make([]definition, 0)
				}
				cursor.inDefs = false
			case "radialGradient", "linearGradient":
				cursor.inGrad = false
			}
		case xml.CharData:
			if cursor.inTitleText {
				icon.Titles[len(icon.Titles)-1] += string(se)
			}
			if cursor.inDescText {
				icon.Descriptions[len(icon.Descriptions)-1] += string(se)
			}
		}
	}
	return icon, nil
}

// ReadIcon reads the Icon from the named file
// This only supports a sub-set of SVG, but
// is enough to draw many icons. errMode determines if the icon ignores, errors out, or logs a warning
// if it does not handle an element found in the icon file.
func ReadIcon(iconFile string, errMode ErrorMode) (*SvgIcon, error) {
	fin, errf := os.Open(iconFile)
	if errf != nil {
		return nil, errf
	}
	defer fin.Close()
	return ReadIconStream(fin, errMode)
}