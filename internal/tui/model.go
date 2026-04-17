package tui

import (
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
	allProjectsCache []*store.Project
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
		reloaded, err := store.LoadProject(m.project.Dir)
		if err != nil {
			m.banner = "reload: " + err.Error()
			return m, nil
		}
		m.project = reloaded
		_ = store.Normalize(m.project, today())
		_ = store.SaveProject(m.project)
		if m.allProjects {
			m.reloadAllProjects()
		}
	case tea.KeyMsg:
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
			return m, m.openEditor(ec)
		case "e":
			path, err := m.fileForFilter()
			if err != nil {
				m.banner = err.Error()
				return m, nil
			}
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
	ti := doc.Items[idx].(store.TaskItem)
	ti.Task.Status = s
	ti.RawLines[0] = store.RenderTaskLine(ti.Task)
	doc.Items[idx] = ti
	_ = store.Normalize(proj, today())
	_ = store.SaveProject(proj)
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
	case tea.KeyEnter:
		if m.promptBuf != "" {
			m.insertNewTask(m.promptBuf)
		}
		m.prompting = false
		m.promptBuf = ""
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
	doc.Items = append(doc.Items[:idx], doc.Items[idx+1:]...)
	_ = store.SaveProject(proj)
	if m.cursor >= len(m.visibleTasks()) && m.cursor > 0 {
		m.cursor--
	}
}

// insertNewTask appends a pending task to active.md and persists.
func (m *Model) insertNewTask(title string) {
	ti := store.TaskItem{
		Task:     model.Task{Status: model.StatusPending, Title: title},
		RawLines: []string{"- [ ] " + title},
	}
	m.project.Active.Items = append(m.project.Active.Items, ti)
	_ = store.SaveProject(m.project)
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
			if err == nil {
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

// reloadAllProjects refreshes the allProjectsCache from disk.
func (m *Model) reloadAllProjects() {
	names, err := store.ListProjects(m.baseDir)
	if err != nil {
		return
	}
	m.allProjectsCache = m.allProjectsCache[:0]
	for _, n := range names {
		p, err := store.LoadProject(filepath.Join(m.baseDir, n))
		if err != nil {
			continue
		}
		m.allProjectsCache = append(m.allProjectsCache, p)
	}
}
