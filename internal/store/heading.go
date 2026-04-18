package store

import "regexp"

// plainProjectRe matches a heading body that looks like a directory name:
// alphanumeric, underscore, or dash. A heading like "my-project" is
// assumed to be auto-generated and rewritten to follow directory renames;
// prose like "My Project" is treated as user-customized and left alone.
var plainProjectRe = regexp.MustCompile(`^[A-Za-z0-9_\-]+$`)

// ensureProjectHeadingDoc mutates doc so it begins with a ProjectHeading
// for project. Rules:
//
//   - No existing ProjectHeading: prepend one.
//   - Existing ProjectHeading with Name == project: leave it alone.
//   - Existing ProjectHeading whose Name is a plain project-like token
//     (alphanumeric + -_): rewrite to project (picks up dir renames).
//   - Existing ProjectHeading with prose-like text: leave it alone.
func ensureProjectHeadingDoc(doc *Document, project string) {
	heading := ProjectHeading{Name: project, RawLine: "# " + project}

	if len(doc.Items) == 0 {
		doc.Items = []Item{heading}
		return
	}
	ph, ok := doc.Items[0].(ProjectHeading)
	if !ok {
		doc.Items = append([]Item{heading}, doc.Items...)
		return
	}
	if ph.Name == project {
		return
	}
	if plainProjectRe.MatchString(ph.Name) {
		doc.Items[0] = heading
	}
}
