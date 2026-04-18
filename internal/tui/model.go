package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/paopp2/himo/internal/model"
	"github.com/paopp2/himo/internal/store"
)

// Filter narrows the visible tasks by status.
type Filter struct {
	Statuses []model.Status
	All      bool
}

// DefaultFilter is the active+pending+blocked view shown on startup.
func DefaultFilter() Filter {
	return Filter{Statuses: []model.Status{model.StatusPending, model.StatusActive, model.StatusBlocked}}
}

// Name returns a stable short string identifying this filter — "all",
// "default" (the pending+active+blocked combo), or a single status name.
// Returns "" for any other shape (not persisted).
func (f Filter) Name() string {
	if f.All {
		return "all"
	}
	if isDefaultFilter(f) {
		return "default"
	}
	if len(f.Statuses) == 1 {
		return f.Statuses[0].String()
	}
	return ""
}

// FilterFromName is the inverse of Filter.Name. Unknown names fall back
// to the default filter.
func FilterFromName(name string) Filter {
	switch name {
	case "all":
		return Filter{All: true}
	case "", "default":
		return DefaultFilter()
	}
	if s, ok := model.ParseStatusName(name); ok {
		return Filter{Statuses: []model.Status{s}}
	}
	return DefaultFilter()
}

func isDefaultFilter(f Filter) bool {
	if f.All || len(f.Statuses) != 3 {
		return false
	}
	have := map[model.Status]bool{}
	for _, s := range f.Statuses {
		have[s] = true
	}
	return have[model.StatusPending] && have[model.StatusActive] && have[model.StatusBlocked]
}

var digitFilters = map[string]model.Status{
	"1": model.StatusBacklog,
	"2": model.StatusPending,
	"3": model.StatusActive,
	"4": model.StatusBlocked,
	"5": model.StatusDone,
	"6": model.StatusCancelled,
}

var statusActionKeys = map[string]model.Status{
	"b": model.StatusBacklog,
	"p": model.StatusPending,
	"a": model.StatusActive,
	"!": model.StatusBlocked,
	"x": model.StatusDone,
	"-": model.StatusCancelled,
}

// Model is the top-level Bubble Tea model.
type Model struct {
	project           *store.Project
	filter            Filter
	cursor            int
	width             int
	height            int
	quit              bool
	baseDir           string
	projects          []string
	hidePreview       bool
	prompting         bool
	promptBuf         string
	promptAbove       bool
	confirmingDelete  bool
	banner            string
	searching         bool
	searchBuf         string
	searchActive      string
	showingHelp       bool
	pickerOpen        bool
	pickerCursor      int
	pickerFilter      string
	allProjects       bool
	allProjectsCache  []*store.Project
	editingProjectDir string
	undoStack         []historyEntry
	redoStack         []historyEntry
	styles            *Styles
}

// NewModel builds a fresh Model for the given project.
func NewModel(p *store.Project) Model {
	return NewModelWithOptions(p, StyleOptions{})
}

// NewModelWithOptions is like NewModel but takes style options.
func NewModelWithOptions(p *store.Project, opts StyleOptions) Model {
	return Model{project: p, filter: DefaultFilter(), styles: NewStyles(opts)}
}

// NewModelFromBase loads the named project from baseDir and returns a Model
// seeded with the list of sibling projects for Tab cycling.
func NewModelFromBase(baseDir, name string, opts StyleOptions) (Model, error) {
	p, err := store.LoadProject(filepath.Join(baseDir, name))
	if err != nil {
		return Model{}, err
	}
	projects, err := store.ListProjects(baseDir)
	if err != nil {
		return Model{}, err
	}
	return Model{
		project:  p,
		filter:   DefaultFilter(),
		baseDir:  baseDir,
		projects: projects,
		styles:   NewStyles(opts),
	}, nil
}

// WithFilter returns m with its initial filter replaced. Used by main.go
// to restore a persisted filter on launch.
func (m Model) WithFilter(f Filter) Model {
	m.filter = f
	return m
}

// WithAllProjects returns m entered into all-projects mode. Used by main.go
// to restore a persisted all-projects scope on launch.
func (m Model) WithAllProjects() Model {
	m.enterAllProjects()
	return m
}

// SessionProject is the single-project scope the user was in at quit,
// regardless of whether they were in all-projects view at the moment.
func (m Model) SessionProject() string {
	if m.project == nil {
		return ""
	}
	return m.project.Name
}

// SessionFilter is the current filter encoded as a stable name for state
// persistence. Returns "" when the filter has no canonical name.
func (m Model) SessionFilter() string {
	return m.filter.Name()
}

// SessionAllProjects reports whether the user is in all-projects view.
func (m Model) SessionAllProjects() bool {
	return m.allProjects
}

func (m Model) Init() tea.Cmd { return nil }

// currentMode reports the interaction mode implied by m's overlay flags.
func (m Model) currentMode() Mode {
	switch {
	case m.showingHelp:
		return ModeHelp
	case m.searching:
		return ModeSearch
	case m.prompting:
		return ModePrompt
	case m.confirmingDelete:
		return ModeDelete
	case m.pickerOpen:
		return ModePicker
	}
	return ModeNormal
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case editorReturnedMsg:
		if msg.err != nil {
			m.banner = "editor: " + msg.err.Error()
		}
		dir := m.editingProjectDir
		if dir == "" {
			dir = m.project.Dir
		}
		reloaded, err := store.LoadProject(dir)
		if err != nil {
			m.banner = "reload: " + err.Error()
			m.editingProjectDir = ""
			return m, nil
		}
		if err := store.Normalize(reloaded, today()); err != nil {
			m.banner = "normalize: " + err.Error()
		}
		_ = m.saveWithBanner(reloaded, "save")
		// Bind reloaded onto m.project when it matches the current project.
		if reloaded.Dir == m.project.Dir {
			m.project = reloaded
		}
		if m.allProjects {
			m.reloadAllProjects()
			// After cache refresh, rebind m.project by Dir match.
			want := m.project.Dir
			for _, p := range m.allProjectsCache {
				if p.Dir == want {
					m.project = p
					break
				}
			}
		}
		m.editingProjectDir = ""
	case tea.KeyMsg:
		if m.showingHelp {
			switch msg.String() {
			case "?", "esc", "q":
				m.showingHelp = false
			}
			return m, nil
		}
		if m.prompting {
			return m.updatePrompt(msg), nil
		}
		if m.confirmingDelete {
			switch msg.String() {
			case "y":
				m.deleteCurrent()
				m.confirmingDelete = false
			case "n", "esc":
				m.confirmingDelete = false
			}
			return m, nil
		}
		if m.pickerOpen {
			return m.updatePicker(msg), nil
		}
		if m.searching {
			switch msg.Type {
			case tea.KeyEsc:
				m.searching = false
				m.searchBuf = ""
			case tea.KeyEnter:
				m.searchActive = m.searchBuf
				m.searching = false
				m.searchBuf = ""
				m.cursor = 0
			case tea.KeyBackspace:
				if len(m.searchBuf) > 0 {
					m.searchBuf = m.searchBuf[:len(m.searchBuf)-1]
				}
			case tea.KeyRunes:
				m.searchBuf += string(msg.Runes)
			case tea.KeySpace:
				m.searchBuf += " "
			}
			return m, nil
		}
		switch msg.String() {
		case "q", "ctrl+c":
			m.quit = true
			return m, tea.Quit
		case "j", "down":
			if m.cursor+1 < len(m.visibleTasks()) {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "g":
			m.cursor = 0
		case "G":
			if n := len(m.visibleTasks()); n > 0 {
				m.cursor = n - 1
			}
		case "ctrl+d":
			half := maxInt(m.height/2, 1)
			if n := len(m.visibleTasks()); n > 0 {
				m.cursor = minInt(m.cursor+half, n-1)
			}
		case "ctrl+u":
			half := maxInt(m.height/2, 1)
			m.cursor = maxInt(m.cursor-half, 0)
		case "0":
			m.filter = Filter{All: true}
			m.cursor = 0
		case "esc":
			if m.allProjects {
				m.exitAllProjects()
			} else if m.searchActive != "" {
				m.searchActive = ""
				m.cursor = 0
			} else {
				m.filter = DefaultFilter()
				m.cursor = 0
			}
		case "tab":
			if m.allProjects {
				return m, nil
			}
			m.switchProject(+1)
		case "shift+tab":
			if m.allProjects {
				return m, nil
			}
			m.switchProject(-1)
		case "P":
			m.pickerOpen = true
			m.pickerCursor = 0
			m.pickerFilter = ""
		case "A":
			if m.allProjects {
				m.exitAllProjects()
			} else {
				m.enterAllProjects()
			}
		case "v":
			m.hidePreview = !m.hidePreview
		case "u":
			m.undo()
		case "ctrl+r":
			m.redo()
		case " ":
			m.cycleStatus()
		case "o":
			m.prompting = true
			m.promptBuf = ""
			m.promptAbove = false
		case "O":
			m.prompting = true
			m.promptBuf = ""
			m.promptAbove = true
		case "/":
			m.searching = true
			m.searchBuf = ""
		case "d":
			m.confirmingDelete = true
		case "?":
			m.showingHelp = !m.showingHelp
		case "enter":
			ec, err := m.editorCmdForNotes()
			if err != nil {
				m.banner = err.Error()
				return m, nil
			}
			if proj, _, _, ok := m.currentTaskItem(); ok && proj != nil {
				m.editingProjectDir = proj.Dir
			} else {
				m.editingProjectDir = m.project.Dir
			}
			return m, m.openEditor(ec)
		case "e":
			target := m.project
			if m.allProjects {
				if p, _, _, ok := m.currentTaskItem(); ok {
					target = p
				}
			}
			path, err := fileForFilter(m.filter, target)
			if err != nil {
				m.banner = err.Error()
				return m, nil
			}
			m.editingProjectDir = target.Dir
			return m, m.openEditor(editorCmd{Path: path, Line: 0})
		default:
			if s, ok := digitFilters[msg.String()]; ok {
				m.filter = Filter{Statuses: []model.Status{s}}
				m.cursor = 0
				return m, nil
			}
			if s, ok := statusActionKeys[msg.String()]; ok {
				m.setStatus(s)
				return m, nil
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.quit {
		return ""
	}
	return renderView(m)
}

func (m *Model) switchProject(delta int) {
	if len(m.projects) == 0 {
		return
	}
	idx := -1
	for i, n := range m.projects {
		if n == m.project.Name {
			idx = i
			break
		}
	}
	if idx < 0 {
		idx = 0
	}
	next := (idx + delta + len(m.projects)) % len(m.projects)
	p, err := store.LoadProject(filepath.Join(m.baseDir, m.projects[next]))
	if err != nil {
		m.banner = "switch: " + err.Error()
		return
	}
	m.project = p
	m.cursor = 0
}

type taskLoc struct {
	proj *store.Project
	doc  *store.Document
	idx  int
}

func (m Model) visibleTaskLocations() []taskLoc {
	projects := []*store.Project{m.project}
	if m.allProjects {
		projects = m.allProjectsCache
	}
	var out []taskLoc
	for _, p := range projects {
		docs := []*store.Document{p.Active, p.Backlog, p.Done}
		for _, d := range docs {
			for i, it := range d.Items {
				ti, ok := it.(store.TaskItem)
				if !ok {
					continue
				}
				if !m.filter.All {
					match := false
					for _, s := range m.filter.Statuses {
						if ti.Task.Status == s {
							match = true
							break
						}
					}
					if !match {
						continue
					}
				}
				if m.searchActive != "" && !strings.Contains(strings.ToLower(ti.Task.Title), strings.ToLower(m.searchActive)) {
					continue
				}
				out = append(out, taskLoc{proj: p, doc: d, idx: i})
			}
		}
	}
	return out
}

func (m Model) visibleTasks() []model.Task {
	locs := m.visibleTaskLocations()
	out := make([]model.Task, len(locs))
	for i, loc := range locs {
		out[i] = loc.doc.Items[loc.idx].(store.TaskItem).Task
	}
	return out
}

// statusCounts returns the per-status totals across the current scope.
// Search does not narrow counts; only filters narrow what the list shows.
func (m Model) statusCounts() map[model.Status]int {
	out := make(map[model.Status]int, 6)
	add := func(p *store.Project) {
		for _, t := range p.AllTasks() {
			out[t.Status]++
		}
	}
	if m.allProjects {
		for _, p := range m.allProjectsCache {
			add(p)
		}
	} else {
		add(m.project)
	}
	return out
}

// currentTaskItem returns the cursor's task location, or ok=false if out of range.
func (m Model) currentTaskItem() (*store.Project, *store.Document, int, bool) {
	locs := m.visibleTaskLocations()
	if m.cursor < 0 || m.cursor >= len(locs) {
		return nil, nil, -1, false
	}
	loc := locs[m.cursor]
	return loc.proj, loc.doc, loc.idx, true
}

// setStatus changes the cursor task's status, normalizes, and saves.
func (m *Model) setStatus(s model.Status) {
	proj, doc, idx, ok := m.currentTaskItem()
	if !ok {
		return
	}
	m.pushUndo(proj)
	old := doc.Items[idx].(store.TaskItem)
	ti := old
	ti.Task.Status = s
	ti.RawLines[0] = store.RenderTaskLine(ti.Task)
	doc.Items[idx] = ti
	if err := store.Normalize(proj, today()); err != nil {
		doc.Items[idx] = old
		m.banner = "normalize: " + err.Error()
		m.popUndo()
		return
	}
	if err := m.saveWithBanner(proj, "save"); err != nil {
		m.popUndo()
		return
	}
	if n := len(m.visibleTasks()); m.cursor >= n && n > 0 {
		m.cursor = n - 1
	}
}

// saveWithBanner persists proj. Returns nil on success, the original error on
// failure. Sets m.banner with a conflict-specific message when applicable.
// Caller decides whether to roll back in-memory state.
func (m *Model) saveWithBanner(proj *store.Project, action string) error {
	err := store.SaveProject(proj)
	if err == nil {
		return nil
	}
	if store.IsConflict(err) {
		m.banner = action + " blocked: file changed on disk. Restart himo to reload."
	} else {
		m.banner = action + ": " + err.Error()
	}
	return err
}

func (m *Model) cycleStatus() {
	_, doc, idx, ok := m.currentTaskItem()
	if !ok {
		return
	}
	cur := doc.Items[idx].(store.TaskItem).Task.Status
	var next model.Status
	switch cur {
	case model.StatusBacklog:
		next = model.StatusPending
	case model.StatusPending:
		next = model.StatusActive
	case model.StatusActive:
		next = model.StatusDone
	case model.StatusBlocked:
		next = model.StatusActive
	case model.StatusDone, model.StatusCancelled:
		next = model.StatusPending
	default:
		return
	}
	m.setStatus(next)
}

func today() string {
	return time.Now().Format("2006-01-02")
}

// updatePrompt handles a keystroke while the new-task prompt is active.
func (m Model) updatePrompt(msg tea.KeyMsg) Model {
	switch msg.Type {
	case tea.KeyEsc:
		m.prompting = false
		m.promptBuf = ""
		m.promptAbove = false
	case tea.KeyEnter:
		if m.promptBuf != "" {
			m.insertNewTask(m.promptBuf)
		}
		m.prompting = false
		m.promptBuf = ""
		m.promptAbove = false
	case tea.KeyBackspace:
		if len(m.promptBuf) > 0 {
			m.promptBuf = m.promptBuf[:len(m.promptBuf)-1]
		}
	case tea.KeyRunes:
		m.promptBuf += string(msg.Runes)
	case tea.KeySpace:
		m.promptBuf += " "
	}
	return m
}

// deleteCurrent removes the task under the cursor, persists, and clamps cursor.
func (m *Model) deleteCurrent() {
	proj, doc, idx, ok := m.currentTaskItem()
	if !ok {
		return
	}
	m.pushUndo(proj)
	removed := doc.Items[idx]
	doc.Items = append(doc.Items[:idx], doc.Items[idx+1:]...)
	if err := m.saveWithBanner(proj, "delete"); err != nil {
		// Restore the removed item at its original index.
		doc.Items = append(doc.Items, nil)
		copy(doc.Items[idx+1:], doc.Items[idx:])
		doc.Items[idx] = removed
		m.popUndo()
		return
	}
	if m.cursor >= len(m.visibleTasks()) && m.cursor > 0 {
		m.cursor--
	}
}

func (m *Model) insertNewTask(title string) {
	target := m.project
	cursorProj, cursorDoc, cursorIdx, haveCursor := m.currentTaskItem()
	if m.allProjects && haveCursor {
		target = cursorProj
	}
	var task model.Task
	var targetDoc *store.Document
	if m.isBacklogFilter() {
		task = model.Task{Status: model.StatusBacklog, Title: title}
		targetDoc = target.Backlog
	} else {
		task = model.Task{Status: model.StatusPending, Title: title}
		targetDoc = target.Active
	}
	ti := store.TaskItem{Task: task, RawLines: []string{store.RenderTaskLine(task)}}
	insertAt := len(targetDoc.Items)
	if haveCursor && cursorProj == target && cursorDoc == targetDoc {
		if m.promptAbove {
			insertAt = cursorIdx
		} else {
			insertAt = cursorIdx + 1
			if insertAt > len(targetDoc.Items) {
				insertAt = len(targetDoc.Items)
			}
		}
	}
	targetDoc.Items = insertAtSlice(targetDoc.Items, insertAt, ti)
	if err := m.saveWithBanner(target, "new task"); err != nil {
		// Roll back the insert.
		targetDoc.Items = append(targetDoc.Items[:insertAt], targetDoc.Items[insertAt+1:]...)
		return
	}
	// With "o", the new task now sits below the cursor — advance so the
	// cursor lands on it (matches vim's o behavior).
	if !m.promptAbove && haveCursor && m.cursor+1 < len(m.visibleTasks()) {
		m.cursor++
	}
}

// insertAtSlice returns s with it inserted at idx.
func insertAtSlice(s []store.Item, idx int, it store.Item) []store.Item {
	s = append(s, nil)
	copy(s[idx+1:], s[idx:])
	s[idx] = it
	return s
}

func (m Model) isBacklogFilter() bool {
	return !m.filter.All &&
		len(m.filter.Statuses) == 1 &&
		m.filter.Statuses[0] == model.StatusBacklog
}

// filteredProjects returns project names that contain pickerFilter (case-insensitive).
func (m Model) filteredProjects() []string {
	if m.pickerFilter == "" {
		return m.projects
	}
	needle := strings.ToLower(m.pickerFilter)
	out := make([]string, 0, len(m.projects))
	for _, n := range m.projects {
		if strings.Contains(strings.ToLower(n), needle) {
			out = append(out, n)
		}
	}
	return out
}

// updatePicker handles keystrokes while the project picker is open.
func (m Model) updatePicker(msg tea.KeyMsg) Model {
	switch msg.Type {
	case tea.KeyEsc:
		m.pickerOpen = false
		m.pickerFilter = ""
		m.pickerCursor = 0
		return m
	case tea.KeyEnter:
		names := m.filteredProjects()
		if m.pickerCursor >= 0 && m.pickerCursor < len(names) {
			p, err := store.LoadProject(filepath.Join(m.baseDir, names[m.pickerCursor]))
			if err != nil {
				m.banner = "load: " + err.Error()
			} else {
				m.project = p
				m.cursor = 0
			}
		}
		m.pickerOpen = false
		m.pickerFilter = ""
		m.pickerCursor = 0
		return m
	case tea.KeyUp:
		if m.pickerCursor > 0 {
			m.pickerCursor--
		}
	case tea.KeyDown:
		if m.pickerCursor+1 < len(m.filteredProjects()) {
			m.pickerCursor++
		}
	case tea.KeyBackspace:
		if n := len(m.pickerFilter); n > 0 {
			m.pickerFilter = m.pickerFilter[:n-1]
			m.pickerCursor = 0
		}
	case tea.KeyRunes:
		m.pickerFilter += string(msg.Runes)
		m.pickerCursor = 0
	case tea.KeySpace:
		m.pickerFilter += " "
		m.pickerCursor = 0
	}
	return m
}

func (m *Model) enterAllProjects() {
	if m.baseDir == "" {
		return
	}
	m.reloadAllProjects()
	m.allProjects = true
	m.cursor = 0
}

func (m *Model) exitAllProjects() {
	m.allProjects = false
	m.allProjectsCache = nil
	m.cursor = 0
}

func (m *Model) reloadAllProjects() {
	names, err := store.ListProjects(m.baseDir)
	if err != nil {
		m.banner = "list: " + err.Error()
		return
	}
	next := make([]*store.Project, 0, len(names))
	var failed []string
	for _, n := range names {
		p, err := store.LoadProject(filepath.Join(m.baseDir, n))
		if err != nil {
			failed = append(failed, n)
			continue
		}
		next = append(next, p)
	}
	m.allProjectsCache = next
	if len(failed) > 0 {
		m.banner = fmt.Sprintf("skipped %d project(s): %s", len(failed), strings.Join(failed, ", "))
	}
}
