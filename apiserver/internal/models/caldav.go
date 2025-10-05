package models

import "encoding/xml"

const (
	DavNamespace            = "DAV:"
	CalDavNamespace         = "urn:ietf:params:xml:ns:caldav"
	CalendarServerNamespace = "http://calendarserver.org/ns/"
	SabreNamespace          = "http://sabredav.org/ns"
	AppleNamespace          = "http://apple.com/ns/ical/"
)

type (
	Multistatus struct {
		XMLName            xml.Name   `xml:"d:multistatus"`
		DAVAttr            string     `xml:"xmlns:d,attr"`
		CalDAVAttr         string     `xml:"xmlns:cal,attr"`
		CalendarServerAttr string     `xml:"xmlns:cs,attr"`
		SabreAttr          string     `xml:"xmlns:s,attr"`
		AppleAttr          string     `xml:"xmlns:x1,attr"`
		Responses          []Response `xml:"d:response"`
	}

	Response struct {
		Href     string     `xml:"d:href"`
		Propstat []Propstat `xml:"d:propstat"`
	}

	Propstat struct {
		Prop   Prop   `xml:"d:prop"`
		Status string `xml:"d:status"`
	}

	Prop struct {
		ResourceType        *ResourceType        `xml:"d:resourcetype,omitempty"`
		GetCTag             string               `xml:"cs:getctag,omitempty"`
		SyncToken           string               `xml:"s:sync-token,omitempty"`
		DisplayName         string               `xml:"d:displayname,omitempty"`
		CalendarTimeZone    string               `xml:"cal:calendar-timezone,omitempty"`
		SupportedComponents *SupportedComponents `xml:"cal:supported-calendar-component-set,omitempty"`
		GetLastModified     string               `xml:"d:getlastmodified,omitempty"`
		GetContentLength    int                  `xml:"d:getcontentlength,omitempty"`
		GetETag             string               `xml:"d:getetag,omitempty"`
		GetContentType      string               `xml:"d:getcontenttype,omitempty"`
		CalendarDescription string               `xml:"cal:calendar-description,omitempty"`
		CalendarOrder       string               `xml:"x1:calendar-order,omitempty"`
		CalendarColor       string               `xml:"x1:calendar-color,omitempty"`
		CalendarData        string               `xml:"cal:calendar-data,omitempty"`
	}

	CalendarMultiget struct {
		XMLName    xml.Name `xml:"calendar-multiget"`
		DAVAttr    string   `xml:"xmlns:D,attr"`
		CalDAVAttr string   `xml:"xmlns:C,attr"`
		Prop       struct {
			GetETag      *struct{} `xml:"D:getetag"`
			CalendarData *struct{} `xml:"C:calendar-data"`
		} `xml:"D:prop"`
		Hrefs []string `xml:"href"`
	}

	ResourceType struct {
		Collection  *struct{} `xml:"d:collection,omitempty"`
		Calendar    *struct{} `xml:"cal:calendar,omitempty"`
		SharedOwner *struct{} `xml:"cs:shared-owner,omitempty"`
	}

	SupportedComponents struct {
		Comp []CalComp `xml:"cal:comp"`
	}

	CalComp struct {
		Name string `xml:"name,attr"`
	}
)
