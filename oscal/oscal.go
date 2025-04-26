// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Contributors:
//	Fraunhofer AISEC

// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    coordinate, err := UnmarshalCoordinate(bytes)
//    bytes, err = coordinate.Marshal()
//
// This file was parsed from https://raw.githubusercontent.com/usnistgov/OSCAL/main/json/schema/oscal_catalog_schema.json

package oscal

import "encoding/json"

func UnmarshalCoordinate(data []byte) (Coordinate, error) {
	var r Coordinate
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Coordinate) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Coordinate struct {
	Catalog Catalog `json:"catalog"`
}

// A collection of controls.
type Catalog struct {
	BackMatter *BackMatter         `json:"back-matter,omitempty"`
	Controls   []Control           `json:"controls,omitempty"`
	Groups     []ControlGroup      `json:"groups,omitempty"`
	Metadata   PublicationMetadata `json:"metadata"`
	Params     []Parameter         `json:"params,omitempty"`
	UUID       string              `json:"uuid"` // A globally unique identifier with cross-instance scope for this catalog instance. This; UUID should be changed when this document is revised.
}

// A collection of resources, which may be included directly or by reference.
type BackMatter struct {
	Resources []Resource `json:"resources,omitempty"`
}

// A resource associated with content in the containing document. A resource may be directly
// included in the document base64 encoded or may point to one or more equivalent internet
// resources.
type Resource struct {
	Base64      *Base64              `json:"base64,omitempty"`      // The Base64 alphabet in RFC 2045 - aligned with XSD.
	Citation    *Citation            `json:"citation,omitempty"`    // A citation consisting of end note text and optional structured bibliographic data.
	Description *string              `json:"description,omitempty"` // A short summary of the resource used to indicate the purpose of the resource.
	DocumentIDS []DocumentIdentifier `json:"document-ids,omitempty"`
	Props       []Property           `json:"props,omitempty"`
	Remarks     *string              `json:"remarks,omitempty"`
	Rlinks      []ResourceLink       `json:"rlinks,omitempty"`
	Title       *string              `json:"title,omitempty"` // A name given to the resource, which may be used by a tool for display and navigation.
	UUID        string               `json:"uuid"`            // A machine-oriented, globally unique identifier with cross-instance scope that can be used; to reference this defined resource elsewhere in this or other OSCAL instances. This UUID; should be assigned per-subject, which means it should be consistently used to identify; the same subject across revisions of the document.
}

// The Base64 alphabet in RFC 2045 - aligned with XSD.
type Base64 struct {
	Filename  *string `json:"filename,omitempty"`   // Name of the file before it was encoded as Base64 to be embedded in a resource. This is; the name that will be assigned to the file when the file is decoded.
	MediaType *string `json:"media-type,omitempty"` // Specifies a media type as defined by the Internet Assigned Numbers Authority (IANA) Media; Types Registry.
	Value     string  `json:"value"`
}

// A citation consisting of end note text and optional structured bibliographic data.
type Citation struct {
	Links []Link     `json:"links,omitempty"`
	Props []Property `json:"props,omitempty"`
	Text  string     `json:"text"` // A line of citation text.
}

// A reference to a local or remote resource
type Link struct {
	Href      string  `json:"href"`                 // A resolvable URL reference to a resource.
	MediaType *string `json:"media-type,omitempty"` // Specifies a media type as defined by the Internet Assigned Numbers Authority (IANA) Media; Types Registry.
	Rel       *string `json:"rel,omitempty"`        // Describes the type of relationship provided by the link. This can be an indicator of the; link's purpose.
	Text      *string `json:"text,omitempty"`       // A textual label to associate with the link, which may be used for presentation in a tool.
}

// An attribute, characteristic, or quality of the containing object expressed as a
// namespace qualified name/value pair. The value of a property is a simple scalar value,
// which may be expressed as a list of values.
type Property struct {
	Class   *string `json:"class,omitempty"` // A textual label that provides a sub-type or characterization of the property's name. This; can be used to further distinguish or discriminate between the semantics of multiple; properties of the same object with the same name and ns.
	Name    string  `json:"name"`            // A textual label that uniquely identifies a specific attribute, characteristic, or quality; of the property's containing object.
	NS      *string `json:"ns,omitempty"`    // A namespace qualifying the property's name. This allows different organizations to; associate distinct semantics with the same name.
	Remarks *string `json:"remarks,omitempty"`
	UUID    *string `json:"uuid,omitempty"` // A machine-oriented, globally unique identifier with cross-instance scope that can be used; to reference this defined property elsewhere in this or other OSCAL instances. This UUID; should be assigned per-subject, which means it should be consistently used to identify; the same subject across revisions of the document.
	Value   string  `json:"value"`          // Indicates the value of the attribute, characteristic, or quality.
}

// A document identifier qualified by an identifier scheme. A document identifier provides a
// globally unique identifier with a cross-instance scope that is used for a group of
// documents that are to be treated as different versions of the same document. If this
// element does not appear, or if the value of this element is empty, the value of
// "document-id" is equal to the value of the "uuid" flag of the top-level root element.
type DocumentIdentifier struct {
	Identifier string  `json:"identifier"`
	Scheme     *string `json:"scheme,omitempty"` // Qualifies the kind of document identifier using a URI. If the scheme is not provided the; value of the element will be interpreted as a string of characters.
}

// A pointer to an external resource with an optional hash for verification and change
// detection.
type ResourceLink struct {
	Hashes    []Hash  `json:"hashes,omitempty"`
	Href      string  `json:"href"`                 // A resolvable URI reference to a resource.
	MediaType *string `json:"media-type,omitempty"` // Specifies a media type as defined by the Internet Assigned Numbers Authority (IANA) Media; Types Registry.
}

// A representation of a cryptographic digest generated over a resource using a specified
// hash algorithm.
type Hash struct {
	Algorithm string `json:"algorithm"` // Method by which a hash is derived
	Value     string `json:"value"`
}

// A structured information object representing a security or privacy control. Each security
// or privacy control within the Catalog is defined by a distinct control instance.
type Control struct {
	Class    *string     `json:"class,omitempty"` // A textual label that provides a sub-type or characterization of the control.
	Controls []Control   `json:"controls,omitempty"`
	ID       string      `json:"id"` // A human-oriented, locally unique identifier with instance scope that can be used to; reference this control elsewhere in this and other OSCAL instances (e.g., profiles). This; id should be assigned per-subject, which means it should be consistently used to identify; the same control across revisions of the document.
	Links    []Link      `json:"links,omitempty"`
	Params   []Parameter `json:"params,omitempty"`
	Parts    []Part      `json:"parts,omitempty"`
	Props    []Property  `json:"props,omitempty"`
	Title    string      `json:"title"` // A name given to the control, which may be used by a tool for display and navigation.
}

// Parameters provide a mechanism for the dynamic assignment of value(s) in a control.
type Parameter struct {
	Class       *string      `json:"class,omitempty"` // A textual label that provides a characterization of the parameter.
	Constraints []Constraint `json:"constraints,omitempty"`
	DependsOn   *string      `json:"depends-on,omitempty"` // **(deprecated)** Another parameter invoking this one. This construct has been deprecated; and should not be used.
	Guidelines  []Guideline  `json:"guidelines,omitempty"`
	ID          string       `json:"id"`              // A human-oriented, locally unique identifier with cross-instance scope that can be used to; reference this defined parameter elsewhere in this or other OSCAL instances. When; referenced from another OSCAL instance, this identifier must be referenced in the context; of the containing resource (e.g., import-profile). This id should be assigned; per-subject, which means it should be consistently used to identify the same subject; across revisions of the document.
	Label       *string      `json:"label,omitempty"` // A short, placeholder name for the parameter, which can be used as a substitute for a; value if no value is assigned.
	Links       []Link       `json:"links,omitempty"`
	Props       []Property   `json:"props,omitempty"`
	Remarks     *string      `json:"remarks,omitempty"`
	Select      *Selection   `json:"select,omitempty"`
	Usage       *string      `json:"usage,omitempty"` // Describes the purpose and use of a parameter
	Values      []string     `json:"values,omitempty"`
}

// A formal or informal expression of a constraint or test
type Constraint struct {
	Description *string          `json:"description,omitempty"` // A textual summary of the constraint to be applied.
	Tests       []ConstraintTest `json:"tests,omitempty"`
}

// A test expression which is expected to be evaluated by a tool.
type ConstraintTest struct {
	Expression string  `json:"expression"` // A formal (executable) expression of a constraint
	Remarks    *string `json:"remarks,omitempty"`
}

// A prose statement that provides a recommendation for the use of a parameter.
type Guideline struct {
	Prose string `json:"prose"` // Prose permits multiple paragraphs, lists, tables etc.
}

// Presenting a choice among alternatives
type Selection struct {
	Choice  []string              `json:"choice,omitempty"`
	HowMany *ParameterCardinality `json:"how-many,omitempty"` // Describes the number of selections that must occur. Without this setting, only one value; should be assumed to be permitted.
}

// A partition of a control's definition or a child of another part.
type Part struct {
	Class *string    `json:"class,omitempty"` // A textual label that provides a sub-type or characterization of the part's name. This can; be used to further distinguish or discriminate between the semantics of multiple parts of; the same control with the same name and ns.
	ID    *string    `json:"id,omitempty"`    // A human-oriented, locally unique identifier with cross-instance scope that can be used to; reference this defined part elsewhere in this or other OSCAL instances. When referenced; from another OSCAL instance, this identifier must be referenced in the context of the; containing resource (e.g., import-profile). This id should be assigned per-subject, which; means it should be consistently used to identify the same subject across revisions of the; document.
	Links []Link     `json:"links,omitempty"`
	Name  string     `json:"name"`         // A textual label that uniquely identifies the part's semantic type.
	NS    *string    `json:"ns,omitempty"` // A namespace qualifying the part's name. This allows different organizations to associate; distinct semantics with the same name.
	Parts []Part     `json:"parts,omitempty"`
	Props []Property `json:"props,omitempty"`
	Prose *string    `json:"prose,omitempty"` // Permits multiple paragraphs, lists, tables etc.
	Title *string    `json:"title,omitempty"` // A name given to the part, which may be used by a tool for display and navigation.
}

// A group of controls, or of groups of controls.
type ControlGroup struct {
	Class    *string        `json:"class,omitempty"` // A textual label that provides a sub-type or characterization of the group.
	Controls []Control      `json:"controls,omitempty"`
	Groups   []ControlGroup `json:"groups,omitempty"`
	ID       *string        `json:"id,omitempty"` // A human-oriented, locally unique identifier with cross-instance scope that can be used to; reference this defined group elsewhere in in this and other OSCAL instances (e.g.,; profiles). This id should be assigned per-subject, which means it should be consistently; used to identify the same group across revisions of the document.
	Links    []Link         `json:"links,omitempty"`
	Params   []Parameter    `json:"params,omitempty"`
	Parts    []Part         `json:"parts,omitempty"`
	Props    []Property     `json:"props,omitempty"`
	Title    string         `json:"title"` // A name given to the group, which may be used by a tool for display and navigation.
}

// Provides information about the publication and availability of the containing document.
type PublicationMetadata struct {
	DocumentIDS        []DocumentIdentifier        `json:"document-ids,omitempty"`
	LastModified       string                      `json:"last-modified"`
	Links              []Link                      `json:"links,omitempty"`
	Locations          []Location                  `json:"locations,omitempty"`
	OscalVersion       string                      `json:"oscal-version"`
	Parties            []PartyOrganizationOrPerson `json:"parties,omitempty"`
	Props              []Property                  `json:"props,omitempty"`
	Published          *string                     `json:"published,omitempty"`
	Remarks            *string                     `json:"remarks,omitempty"`
	ResponsibleParties []ResponsibleParty          `json:"responsible-parties,omitempty"`
	Revisions          []RevisionHistoryEntry      `json:"revisions,omitempty"`
	Roles              []Role                      `json:"roles,omitempty"`
	Title              string                      `json:"title"` // A name given to the document, which may be used by a tool for display and navigation.
	Version            string                      `json:"version"`
}

// A location, with associated metadata that can be referenced.
type Location struct {
	Address          Address           `json:"address"`
	EmailAddresses   []string          `json:"email-addresses,omitempty"`
	Links            []Link            `json:"links,omitempty"`
	Props            []Property        `json:"props,omitempty"`
	Remarks          *string           `json:"remarks,omitempty"`
	TelephoneNumbers []TelephoneNumber `json:"telephone-numbers,omitempty"`
	Title            *string           `json:"title,omitempty"` // A name given to the location, which may be used by a tool for display and navigation.
	Urls             []string          `json:"urls,omitempty"`
	UUID             string            `json:"uuid"` // A machine-oriented, globally unique identifier with cross-instance scope that can be used; to reference this defined location elsewhere in this or other OSCAL instances. The; locally defined UUID of the location can be used to reference the data item locally or; globally (e.g., from an importing OSCAL instance). This UUID should be assigned; per-subject, which means it should be consistently used to identify the same subject; across revisions of the document.
}

// A postal address for the location.
type Address struct {
	AddrLines  []string `json:"addr-lines,omitempty"`
	City       *string  `json:"city,omitempty"`        // City, town or geographical region for the mailing address.
	Country    *string  `json:"country,omitempty"`     // The ISO 3166-1 alpha-2 country code for the mailing address.
	PostalCode *string  `json:"postal-code,omitempty"` // Postal or ZIP code for mailing address
	State      *string  `json:"state,omitempty"`       // State, province or analogous geographical region for mailing address
	Type       *string  `json:"type,omitempty"`        // Indicates the type of address.
}

// Contact number by telephone.
type TelephoneNumber struct {
	Number string  `json:"number"`
	Type   *string `json:"type,omitempty"` // Indicates the type of phone number.
}

// A responsible entity which is either a person or an organization.
type PartyOrganizationOrPerson struct {
	Addresses             []Address                 `json:"addresses,omitempty"`
	EmailAddresses        []string                  `json:"email-addresses,omitempty"`
	ExternalIDS           []PartyExternalIdentifier `json:"external-ids,omitempty"`
	Links                 []Link                    `json:"links,omitempty"`
	LocationUuids         []string                  `json:"location-uuids,omitempty"`
	MemberOfOrganizations []string                  `json:"member-of-organizations,omitempty"`
	Name                  *string                   `json:"name,omitempty"` // The full name of the party. This is typically the legal name associated with the party.
	Props                 []Property                `json:"props,omitempty"`
	Remarks               *string                   `json:"remarks,omitempty"`
	ShortName             *string                   `json:"short-name,omitempty"` // A short common name, abbreviation, or acronym for the party.
	TelephoneNumbers      []TelephoneNumber         `json:"telephone-numbers,omitempty"`
	Type                  PartyType                 `json:"type"` // A category describing the kind of party the object describes.
	UUID                  string                    `json:"uuid"` // A machine-oriented, globally unique identifier with cross-instance scope that can be used; to reference this defined party elsewhere in this or other OSCAL instances. The locally; defined UUID of the party can be used to reference the data item locally or globally; (e.g., from an importing OSCAL instance). This UUID should be assigned per-subject, which; means it should be consistently used to identify the same subject across revisions of the; document.
}

// An identifier for a person or organization using a designated scheme. e.g. an Open
// Researcher and Contributor ID (ORCID)
type PartyExternalIdentifier struct {
	ID     string `json:"id"`
	Scheme string `json:"scheme"` // Indicates the type of external identifier.
}

// A reference to a set of organizations or persons that have responsibility for performing
// a referenced role in the context of the containing object.
type ResponsibleParty struct {
	Links      []Link     `json:"links,omitempty"`
	PartyUuids []string   `json:"party-uuids"`
	Props      []Property `json:"props,omitempty"`
	Remarks    *string    `json:"remarks,omitempty"`
	RoleID     string     `json:"role-id"` // A human-oriented identifier reference to roles served by the user.
}

// An entry in a sequential list of revisions to the containing document in reverse
// chronological order (i.e., most recent previous revision first).
type RevisionHistoryEntry struct {
	LastModified *string    `json:"last-modified,omitempty"`
	Links        []Link     `json:"links,omitempty"`
	OscalVersion *string    `json:"oscal-version,omitempty"`
	Props        []Property `json:"props,omitempty"`
	Published    *string    `json:"published,omitempty"`
	Remarks      *string    `json:"remarks,omitempty"`
	Title        *string    `json:"title,omitempty"` // A name given to the document revision, which may be used by a tool for display and; navigation.
	Version      string     `json:"version"`
}

// Defines a function assumed or expected to be assumed by a party in a specific situation.
type Role struct {
	Description *string    `json:"description,omitempty"` // A summary of the role's purpose and associated responsibilities.
	ID          string     `json:"id"`                    // A human-oriented, locally unique identifier with cross-instance scope that can be used to; reference this defined role elsewhere in this or other OSCAL instances. When referenced; from another OSCAL instance, the locally defined ID of the Role from the imported OSCAL; instance must be referenced in the context of the containing resource (e.g., import,; import-component-definition, import-profile, import-ssp or import-ap). This ID should be; assigned per-subject, which means it should be consistently used to identify the same; subject across revisions of the document.
	Links       []Link     `json:"links,omitempty"`
	Props       []Property `json:"props,omitempty"`
	Remarks     *string    `json:"remarks,omitempty"`
	ShortName   *string    `json:"short-name,omitempty"` // A short common name, abbreviation, or acronym for the role.
	Title       string     `json:"title"`                // A name given to the role, which may be used by a tool for display and navigation.
}

// Describes the number of selections that must occur. Without this setting, only one value
// should be assumed to be permitted.
type ParameterCardinality string

const (
	One       ParameterCardinality = "one"
	OneOrMore ParameterCardinality = "one-or-more"
)

// A category describing the kind of party the object describes.
type PartyType string

const (
	Organization PartyType = "organization"
	Person       PartyType = "person"
)
