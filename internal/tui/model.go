package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/npaolopepito/himo/internal/model"
	"github.com/npaolopepito/himo/internal/store"
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

// Model is the top-level Bubble Tea model.
type Model struct {
	project     *store.Project
	filter      Filter
	cursor      int
	width       int
	height      int
	quit        bool
	baseDir     string
	projects    []string
	hidePreview bool
	prompting        bool
	promptBuf        string
	promptAbove      bool
	confirmingDelete bool
	banner           string
	searching        bool
	searchBuf        string
	searchActive     string
	showingHelp      bool
	pickerOpen       bool
	pickerCursor     int
	pickerFilter     string
	allProjects      bool
	allProjectsCache  []*store.Project
	editingProjectDir string
}

// NewModel builds a fresh Model for the given project.
func NewModel(p *store.Project) Model {
	return Model{project: p, filter: DefaultFilter()}
}

// NewModelFromBase loads the named project from baseDir and returns a Model
// seeded with the list of sibling projects for Tab cycling.
func NewModelFromBase(baseDir, name string) (Model, error) {
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
	}, nil
}

func (m Model) Init() tea.Cmd { return nil }

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
		if err := store.SaveProject(reloaded); err != nil {
			if store.IsConflict(err) {
				m.banner = "save conflict: " + err.Error()
			} else {
				m.banner = "save: " + err.Error()
			}
		}
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
		case "1":
			m.filter = Filter{Statuses: []model.Status{model.StatusBacklog}}
			m.cursor = 0
		case "2":
			m.filter = Filter{Statuses: []model.Status{model.StatusPending}}
			m.cursor = 0
		case "3":
			m.filter = Filter{Statuses: []model.Status{model.StatusActive}}
			m.cursor = 0
		case "4":
			m.filter = Filter{Statuses: []model.Status{model.StatusBlocked}}
			m.cursor = 0
		case "5":
			m.filter = Filter{Statuses: []model.Status{model.StatusDone}}
			m.cursor = 0
		case "6":
			m.filter = Filter{Statuses: []model.Status{model.StatusCancelled}}
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
		case "x":
			m.setStatus(model.StatusDone)
		case "-":
			m.setStatus(model.StatusCancelled)
		case "!":
			m.setStatus(model.StatusBlocked)
		case "p":
			m.setStatus(model.StatusPending)
		case "a":
			m.setStatus(model.StatusActive)
		case "b":
			m.setStatus(model.StatusBacklog)
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

// switchProject cycles m.project by delta through m.projects, in place.
// Delta of +1 is next, -1 is previous; wraps around.
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

// taskLoc locates a TaskItem by its (project, document, index) position.
type taskLoc struct {
	proj *store.Project
	doc  *store.Document
	idx  int
}

// visibleTaskLocations walks Active, Backlog, Done in order and returns one
// taskLoc per TaskItem whose status passes the current filter. In all-projects
// mode, the walk spans every project in allProjectsCache.
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

// currentTaskItem returns the (project, document, index) of the task under the
// cursor, or (nil, nil, -1, false) if the cursor is out of range.
func (m Model) currentTaskItem() (*store.Project, *store.Document, int, bool) {
	locs := m.visibleTaskLocations()
	if m.cursor < 0 || m.cursor >= len(locs) {
		return nil, nil, -1, false
	}
	loc := locs[m.cursor]
	return loc.proj, loc.doc, loc.idx, true
}

// setStatus changes the highlighted task's status, updates its rendered line,
// re-normalizes, and persists. No-op if no task is selected.
func (m *Model) setStatus(s model.Status) {
	proj, doc, idx, ok := m.currentTaskItem()
	if !ok {
		return
	}
	old := doc.Items[idx].(store.TaskItem)
	ti := old
	ti.Task.Status = s
	ti.RawLines[0] = store.RenderTaskLine(ti.Task)
	doc.Items[idx] = ti
	if err := store.Normalize(proj, today()); err != nil {
		doc.Items[idx] = old
		m.banner = "normalize: " + err.Error()
		return
	}
	if err := store.SaveProject(proj); err != nil {
		if store.IsConflict(err) {
			m.banner = "save conflict: " + err.Error()
		} else {
			m.banner = "save: " + err.Error()
		}
		return
	}
	if n := len(m.visibleTasks()); m.cursor >= n && n > 0 {
		m.cursor = n - 1
	}
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

// updatePrompt handles keystrokes while the new-task prompt is active.
// Returns the updated model; Enter commits via insertNewTask, Esc cancels.
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
	removed := doc.Items[idx]
	doc.Items = append(doc.Items[:idx], doc.Items[idx+1:]...)
	if err := store.SaveProject(proj); err != nil {
		// Restore the removed item at its original index.
		doc.Items = append(doc.Items, nil)
		copy(doc.Items[idx+1:], doc.Items[idx:])
		doc.Items[idx] = removed
		if store.IsConflict(err) {
			m.banner = "save conflict: " + err.Error()
		} else {
			m.banner = "save: " + err.Error()
		}
		return
	}
	if m.cursor >= len(m.visibleTasks()) && m.cursor > 0 {
		m.cursor--
	}
}

// insertNewTask inserts a new task into the document matching the current
// filter (backlog.md when the filter is exclusively backlog, active.md
// otherwise) and persists. In all-projects mode the owner is the cursor task's
// project. When promptAbove is set the task lands at the cursor's position in
// the target document (append otherwise).
func (m *Model) insertNewTask(title string) {
	target := m.project
	cursorProj, cursorDoc, cursorIdx, haveCursor := m.currentTaskItem()
	if m.allProjects && haveCursor {
		target = cursorProj
	}
	var ti store.Item
	var targetDoc *store.Document
	if m.isBacklogFilter() {
		ti = store.TaskItem{
			Task:     model.Task{Status: model.StatusBacklog, Title: title},
			RawLines: []string{"- " + title},
		}
		targetDoc = target.Backlog
	} else {
		ti = store.TaskItem{
			Task:     model.Task{Status: model.StatusPending, Title: title},
			RawLines: []string{"- [ ] " + title},
		}
		targetDoc = target.Active
	}
	insertAt := len(targetDoc.Items)
	if m.promptAbove && haveCursor && cursorProj == target && cursorDoc == targetDoc {
		insertAt = cursorIdx
	}
	targetDoc.Items = insertAtSlice(targetDoc.Items, insertAt, ti)
	if err := store.SaveProject(target); err != nil {
		// Roll back the insert.
		targetDoc.Items = append(targetDoc.Items[:insertAt], targetDoc.Items[insertAt+1:]...)
		if store.IsConflict(err) {
			m.banner = "save conflict: " + err.Error()
		} else {
			m.banner = "save: " + err.Error()
		}
	}
}

// insertAtSlice returns s with it inserted at idx.
func insertAtSlice(s []store.Item, idx int, it store.Item) []store.Item {
	s = append(s, nil)
	copy(s[idx+1:], s[idx:])
	s[idx] = it
	return s
}

// isBacklogFilter reports whether the current filter is exclusively backlog.
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

// enterAllProjects loads every project listed under baseDir into the cache.
func (m *Model) enterAllProjects() {
	if m.baseDir == "" {
		return
	}
	m.reloadAllProjects()
	m.allProjects = true
	m.cursor = 0
}

// exitAllProjects clears the cache and restores single-project mode.
func (m *Model) exitAllProjects() {
	m.allProjects = false
	m.allProjectsCache = nil
	m.cursor = 0
}

// reloadAllProjects refreshes the allProjectsCache from disk. On ListProjects
// failure the cache is left untouched; per-project load failures are
// accumulated into the banner.
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
