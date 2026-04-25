package tui

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"

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

// Name returns a stable short string identifying this filter ("all",
// "default" (the pending+active+blocked combo), or a single status name).
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
	promptInput       textinput.Model
	promptAbove       bool
	confirmingDelete  bool
	banner            string
	searching         bool
	searchInput       textinput.Model
	searchActive      string
	preSearchCursor   int
	showingHelp       bool
	pickerOpen        bool
	pickerCursor      int
	pickerInput       textinput.Model
	editing           bool
	editInput         textinput.Model
	editOrig          string
	allProjects       bool
	allProjectsCache  []*store.Project
	editingProjectDir string
	undoStack         []historyEntry
	redoStack         []historyEntry
	styles            *Styles
	sort              Sort
}

// NewModel builds a fresh Model for the given project.
func NewModel(p *store.Project) Model {
	return NewModelWithOptions(p, StyleOptions{})
}

// NewModelWithOptions is like NewModel but takes style options.
func NewModelWithOptions(p *store.Project, opts StyleOptions) Model {
	st := NewStyles(opts)
	return Model{
		project:     p,
		filter:      DefaultFilter(),
		styles:      st,
		searchInput: newStyledInput(st),
		pickerInput: newStyledInput(st),
		promptInput: newStyledInput(st),
		editInput:   newStyledInput(st),
	}
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
	st := NewStyles(opts)
	return Model{
		project:     p,
		filter:      DefaultFilter(),
		baseDir:     baseDir,
		projects:    projects,
		styles:      st,
		searchInput: newStyledInput(st),
		pickerInput: newStyledInput(st),
		promptInput: newStyledInput(st),
		editInput:   newStyledInput(st),
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

// WithSort returns m with its initial sort mode replaced. Used by main.go
// to restore a persisted sort on launch.
func (m Model) WithSort(s Sort) Model {
	m.sort = s
	return m
}

// SessionSort is the current sort encoded as a stable name for state
// persistence.
func (m Model) SessionSort() string {
	return sortName(m.sort)
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
	case m.editing:
		return ModeEdit
	}
	return ModeNormal
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case urlOpenedMsg:
		if msg.err != nil {
			m.banner = "open URL: " + msg.err.Error()
		}
	case editorReturnedMsg:
		m.undoStack = nil
		m.redoStack = nil
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
		if m.editing {
			return m.updateEdit(msg), nil
		}
		if m.searching {
			switch msg.Type {
			case tea.KeyEsc:
				m.searching = false
				m.searchInput.Reset()
				m.cursor = m.preSearchCursor
			case tea.KeyEnter:
				m.searchActive = m.searchInput.Value()
				m.searching = false
				m.searchInput.Reset()
				m.cursor = 0
			default:
				m.searchInput, _ = m.searchInput.Update(msg)
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
		case "`":
			m.filter = DefaultFilter()
			m.cursor = 0
		case "esc":
			if m.allProjects {
				m.exitAllProjects()
			} else if m.searchActive != "" {
				m.searchActive = ""
				m.cursor = 0
			}
		case "tab":
			if m.allProjects {
				m.exitAllProjects()
			}
			m.switchProject(+1)
		case "shift+tab":
			if m.allProjects {
				m.exitAllProjects()
			}
			m.switchProject(-1)
		case "P":
			m.pickerOpen = true
			m.pickerCursor = 0
			m.pickerInput.Reset()
			m.pickerInput.Focus()
		case "A":
			if m.allProjects {
				m.exitAllProjects()
			} else {
				m.enterAllProjects()
			}
		case "v":
			m.hidePreview = !m.hidePreview
		case "s":
			if m.sort == SortStatus {
				m.sort = SortNatural
			} else {
				m.sort = SortStatus
			}
			m.cursor = 0
		case "u":
			m.undo()
		case "ctrl+r":
			m.redo()
		case " ":
			m.cycleStatus()
		case "o":
			m.prompting = true
			m.promptInput.Reset()
			m.promptInput.Focus()
			m.promptAbove = false
		case "O":
			m.prompting = true
			m.promptInput.Reset()
			m.promptInput.Focus()
			m.promptAbove = true
		case "/":
			m.preSearchCursor = m.cursor
			m.searching = true
			m.searchInput.Reset()
			m.searchInput.Focus()
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
		case "ctrl+o":
			_, doc, idx, ok := m.currentTaskItem()
			if !ok {
				return m, nil
			}
			task := doc.Items[idx].(store.TaskItem).Task
			u := task.URL()
			if u == "" {
				m.banner = "no URL in notes"
				return m, nil
			}
			return m, tea.ExecProcess(
				exec.Command("open", u),
				func(err error) tea.Msg { return urlOpenedMsg{err: err} },
			)
		case "e":
			_, doc, idx, ok := m.currentTaskItem()
			if !ok {
				return m, nil
			}
			m.editing = true
			m.editOrig = doc.Items[idx].(store.TaskItem).Task.Title
			m.editInput.Reset()
			m.editInput.SetValue(m.editOrig)
			m.editInput.Focus()
			return m, nil
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
				out = append(out, taskLoc{proj: p, doc: d, idx: i})
			}
		}
	}
	if m.sort == SortStatus && len(out) > 1 {
		projIdx := make(map[*store.Project]int, len(projects))
		for i, p := range projects {
			projIdx[p] = i
		}
		sort.SliceStable(out, func(i, j int) bool {
			a, b := out[i], out[j]
			ra := statusSortRank(a.doc.Items[a.idx].(store.TaskItem).Task.Status)
			rb := statusSortRank(b.doc.Items[b.idx].(store.TaskItem).Task.Status)
			if ra != rb {
				return ra < rb
			}
			return projIdx[a.proj] < projIdx[b.proj]
		})
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
	m.commitUndo()
	m.clampCursor()
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

func (m Model) updatePrompt(msg tea.KeyMsg) Model {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		m.prompting = false
		m.promptInput.Reset()
		m.promptAbove = false
	case tea.KeyEnter:
		if v := m.promptInput.Value(); v != "" {
			m.insertNewTask(v)
		}
		m.prompting = false
		m.promptInput.Reset()
		m.promptAbove = false
	default:
		// Static cursor: textinput.Update only emits a non-nil Cmd for
		// clipboard paste, which himo doesn't wire through.
		m.promptInput, _ = m.promptInput.Update(msg)
	}
	return m
}

func (m Model) updateEdit(msg tea.KeyMsg) Model {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		m.clearEdit()
	case tea.KeyEnter:
		m.commitEdit()
	default:
		m.editInput, _ = m.editInput.Update(msg)
	}
	return m
}

func (m *Model) clearEdit() {
	m.editing = false
	m.editInput.Reset()
	m.editOrig = ""
}

// commitEdit writes the buffered title to the cursor task, rolling back the
// in-memory mutation and dropping the undo entry if the save fails.
func (m *Model) commitEdit() {
	defer m.clearEdit()
	proj, doc, idx, ok := m.currentTaskItem()
	if !ok {
		return
	}
	v := m.editInput.Value()
	if v == "" || v == m.editOrig {
		return
	}
	m.pushUndo(proj)
	old := doc.Items[idx].(store.TaskItem)
	ti := old
	ti.Task.Title = v
	ti.RawLines[0] = store.RenderTaskLine(ti.Task)
	doc.Items[idx] = ti
	if err := m.saveWithBanner(proj, "edit"); err != nil {
		doc.Items[idx] = old
		m.popUndo()
		return
	}
	m.commitUndo()
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
	m.commitUndo()
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
	m.pushUndo(target)
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
		m.popUndo()
		return
	}
	m.commitUndo()
	// With "o", the new task now sits below the cursor; advance so the
	// cursor lands on it (matches vim's o behavior).
	if !m.promptAbove && haveCursor && m.cursor+1 < len(m.visibleTasks()) {
		m.cursor++
	}
}

// newStyledInput returns a textinput with himo's static accent caret.
// CursorStatic avoids wiring blink commands through Update; cursor.Focus
// flips Blink to false on entry so View renders the cell as an
// accent-bg reverse block in color terminals.
func newStyledInput(st *Styles) textinput.Model {
	ti := textinput.New()
	ti.Prompt = ""
	ti.CharLimit = 0
	ti.Cursor.SetMode(cursor.CursorStatic)
	ti.Cursor.Style = st.Accent
	return ti
}

// insertAtSlice returns s with it inserted at idx.
func insertAtSlice(s []store.Item, idx int, it store.Item) []store.Item {
	s = append(s, nil)
	copy(s[idx+1:], s[idx:])
	s[idx] = it
	return s
}

type urlOpenedMsg struct{ err error }

func (m Model) isBacklogFilter() bool {
	return !m.filter.All &&
		len(m.filter.Statuses) == 1 &&
		m.filter.Statuses[0] == model.StatusBacklog
}

// filteredProjects returns project names that contain the picker filter (case-insensitive).
func (m Model) filteredProjects() []string {
	filter := m.pickerInput.Value()
	if filter == "" {
		return m.projects
	}
	needle := strings.ToLower(filter)
	out := make([]string, 0, len(m.projects))
	for _, n := range m.projects {
		if strings.Contains(strings.ToLower(n), needle) {
			out = append(out, n)
		}
	}
	return out
}

func (m Model) updatePicker(msg tea.KeyMsg) Model {
	switch msg.Type {
	case tea.KeyEsc:
		m.pickerOpen = false
		m.pickerInput.Reset()
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
		m.pickerInput.Reset()
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
	default:
		before := m.pickerInput.Value()
		m.pickerInput, _ = m.pickerInput.Update(msg)
		if m.pickerInput.Value() != before {
			m.pickerCursor = 0
		}
	}
	return m
}

func (m *Model) enterAllProjects() {
	if m.baseDir == "" {
		return
	}
	m.reloadAllProjects()
	// Rebind m.project to the cached pointer so mutations, mtime updates,
	// and projectByDir lookups all reference the same *store.Project.
	if m.project != nil {
		want := m.project.Dir
		for _, p := range m.allProjectsCache {
			if p != nil && p.Dir == want {
				m.project = p
				break
			}
		}
	}
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
