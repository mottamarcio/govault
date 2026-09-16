package screens

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mottamarcio/govault/internal/domain"
	"github.com/mottamarcio/govault/internal/service"
	"github.com/mottamarcio/govault/internal/tui/theme"
)

// DashboardFocus represents which pane currently owns keyboard focus.
type DashboardFocus int

const (
	FocusSidebar DashboardFocus = iota
	FocusTable
	FocusSearch
	FocusHistory
)

// ToastMsg carries transient notification text.
type ToastMsg struct {
	Message string
}

// ClearToastMsg clears active toast message.
type ClearToastMsg struct{}

// DashboardModel manages split-pane navigation, records table, search, detail view, and clipboard.
type DashboardModel struct {
	Theme         *theme.Theme
	RecordService *service.RecordService
	Clipboard     *service.ClipboardService

	Sidebar      *SidebarModel
	Detail       *DetailModel
	History      *HistoryModel
	Table        table.Model
	SearchInput  textinput.Model
	Searching    bool
	Focus        DashboardFocus

	// Modals & Overlays
	ConfirmModal   *ModalConfirm
	HelpModal      *ModalHelp
	GeneratorModal *GeneratorModel

	AllRecords      []*domain.Record
	FilteredRecords []*domain.Record
	CurrentCategory SidebarItem

	ToastMessage string
	Width        int
	Height       int
}

// NewDashboardModel creates a new DashboardModel.
func NewDashboardModel(
	th *theme.Theme,
	rs *service.RecordService,
	cs *service.ClipboardService,
	w, h int,
) *DashboardModel {
	sb := NewSidebarModel(th)
	dm := NewDetailModel(th)
	hm := NewHistoryModel(th, rs)
	cm := NewModalConfirm(th)
	hlp := NewModalHelp(th)
	gen := NewGeneratorModel(th)

	ti := textinput.New()
	ti.Placeholder = "Type to search records (/)..."
	ti.Prompt = "🔍 "
	ti.PromptStyle = th.InputPromptStyle
	ti.TextStyle = lipgloss.NewStyle().Foreground(th.Text)
	ti.CharLimit = 64

	columns := []table.Column{
		{Title: "Title", Width: 22},
		{Title: "Type", Width: 10},
		{Title: "Tags", Width: 16},
		{Title: "Updated", Width: 16},
	}

	tbl := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(th.Border).
		BorderBottom(true).
		Bold(true).
		Foreground(th.Primary)
	s.Selected = s.Selected.
		Foreground(th.Background).
		Background(th.Primary).
		Bold(true)
	tbl.SetStyles(s)

	m := &DashboardModel{
		Theme:           th,
		RecordService:   rs,
		Clipboard:       cs,
		Sidebar:         sb,
		Detail:          dm,
		History:         hm,
		Table:           tbl,
		SearchInput:     ti,
		Searching:       false,
		Focus:           FocusSidebar,
		ConfirmModal:    cm,
		HelpModal:       hlp,
		GeneratorModal:  gen,
		AllRecords:      nil,
		FilteredRecords: nil,
		CurrentCategory: sb.SelectedItem(),
		ToastMessage:    "",
		Width:           w,
		Height:          h,
	}

	m.resizePanels()
	return m
}

// Init initializes child components.
func (m *DashboardModel) Init() tea.Cmd {
	return nil
}

// SetDimensions handles window size adjustments.
func (m *DashboardModel) SetDimensions(w, h int) {
	m.Width = w
	m.Height = h
	m.resizePanels()
}

func (m *DashboardModel) resizePanels() {
	sidebarWidth := 26
	rightWidth := m.Width - sidebarWidth - 4
	if rightWidth < 30 {
		rightWidth = 30
	}
	paneHeight := m.Height - 4
	if paneHeight < 10 {
		paneHeight = 10
	}

	m.Sidebar.SetDimensions(sidebarWidth, paneHeight)
	m.Detail.SetDimensions(rightWidth, paneHeight/2)
	m.History.SetDimensions(m.Width-8, m.Height-6)
	m.ConfirmModal.SetDimensions(m.Width, m.Height)
	m.HelpModal.SetDimensions(m.Width, m.Height)
	m.GeneratorModal.SetDimensions(m.Width, m.Height)

	tableHeight := paneHeight - (paneHeight / 2) - 3
	if tableHeight < 4 {
		tableHeight = 4
	}
	m.Table.SetHeight(tableHeight)

	// Adjust column widths dynamically
	colTitle := rightWidth * 35 / 100
	colType := 10
	colTags := rightWidth * 25 / 100
	colUpdated := rightWidth - colTitle - colType - colTags - 6
	if colUpdated < 12 {
		colUpdated = 12
	}

	m.Table.SetColumns([]table.Column{
		{Title: "Title", Width: colTitle},
		{Title: "Type", Width: colType},
		{Title: "Tags", Width: colTags},
		{Title: "Updated", Width: colUpdated},
	})
}

// ReloadRecords fetches active records from RecordService and rebuilds views.
func (m *DashboardModel) ReloadRecords() error {
	records, err := m.RecordService.List(context.Background(), true)
	if err != nil {
		return err
	}
	m.AllRecords = records
	m.Sidebar.SetCountsAndTags(records)
	m.ApplyFilter()
	return nil
}

// ApplyFilter filters records based on current sidebar selection and active search query.
func (m *DashboardModel) ApplyFilter() {
	var filter service.SearchFilter

	if m.CurrentCategory.IsTrash {
		filter.IncludeTrash = true
	} else if m.CurrentCategory.Tag != "" {
		filter.Tags = []string{m.CurrentCategory.Tag}
	} else if m.CurrentCategory.RecordType != "" {
		filter.Type = m.CurrentCategory.RecordType
	}

	filter.Query = m.SearchInput.Value()

	filtered := service.FilterAndRankRecords(m.AllRecords, filter)
	m.FilteredRecords = filtered

	var rows []table.Row
	for _, r := range filtered {
		tagsStr := strings.Join(r.Tags, ", ")
		updatedStr := r.UpdatedAt.Format("2006-01-02 15:04")
		rows = append(rows, table.Row{
			r.Title,
			string(r.Type),
			tagsStr,
			updatedStr,
		})
	}

	m.Table.SetRows(rows)
	if len(rows) > 0 && (m.Table.Cursor() < 0 || m.Table.Cursor() >= len(rows)) {
		m.Table.SetCursor(0)
	}

	// Update selected record in Detail view
	m.syncSelectedDetail()
}

func (m *DashboardModel) syncSelectedDetail() {
	if len(m.FilteredRecords) == 0 {
		m.Detail.SetRecord(nil)
		return
	}
	cursor := m.Table.Cursor()
	if cursor < 0 {
		cursor = 0
		m.Table.SetCursor(0)
	}
	if cursor >= len(m.FilteredRecords) {
		cursor = len(m.FilteredRecords) - 1
		m.Table.SetCursor(cursor)
	}
	m.Detail.SetRecord(m.FilteredRecords[cursor])
}

// Update processes dashboard keyboard navigation and child components.
func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ToastMsg:
		m.ToastMessage = msg.Message
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return ClearToastMsg{}
		})

	case ClearToastMsg:
		m.ToastMessage = ""
		return m, nil

	case HistoryCloseMsg:
		m.Focus = FocusTable
		return m, nil

	case SidebarSelectMsg:
		m.CurrentCategory = msg.Item
		m.ApplyFilter()
		return m, nil

	case ModalConfirmMsg:
		if msg.Confirmed {
			switch msg.Action {
			case ModalActionTrash:
				if err := m.RecordService.Delete(context.Background(), msg.TargetID); err == nil {
					_ = m.ReloadRecords()
					return m, func() tea.Msg { return ToastMsg{Message: "✓ Record moved to trash"} }
				}
			case ModalActionPurge:
				if err := m.RecordService.PurgeTrash(context.Background()); err == nil {
					_ = m.ReloadRecords()
					return m, func() tea.Msg { return ToastMsg{Message: "✓ Trash permanently purged"} }
				}
			case ModalActionRestore:
				if err := m.RecordService.Restore(context.Background(), msg.TargetID); err == nil {
					_ = m.ReloadRecords()
					return m, func() tea.Msg { return ToastMsg{Message: "✓ Record restored to active vault"} }
				}
			}
		}
		return m, nil

	case GeneratorResultMsg:
		if msg.Secret != "" {
			if err := m.Clipboard.Copy(context.Background(), msg.Secret, 45*time.Second); err == nil {
				return m, func() tea.Msg {
					return ToastMsg{Message: "✓ Generated password copied to clipboard (auto-clears in 45s)"}
				}
			}
		}
		return m, nil

	case HelpCloseMsg:
		return m, nil

	case tea.KeyMsg:
		// 1. Delegate to Help Modal if active
		if m.HelpModal.Active {
			var cmd tea.Cmd
			_, cmd = m.HelpModal.Update(msg)
			return m, cmd
		}

		// 2. Delegate to Confirm Modal if active
		if m.ConfirmModal.Active {
			var cmd tea.Cmd
			_, cmd = m.ConfirmModal.Update(msg)
			return m, cmd
		}

		// 3. Delegate to Generator Modal if active
		if m.GeneratorModal.Active {
			var cmd tea.Cmd
			_, cmd = m.GeneratorModal.Update(msg)
			return m, cmd
		}

		// 4. When revision history modal is active, delegate directly to it
		if m.Focus == FocusHistory {
			var cmd tea.Cmd
			_, cmd = m.History.Update(msg)
			return m, cmd
		}

		// When live search is focused
		if m.Focus == FocusSearch {
			switch msg.Type {
			case tea.KeyEsc:
				m.Focus = FocusTable
				m.Searching = false
				m.SearchInput.Blur()
				return m, nil
			case tea.KeyEnter:
				m.Focus = FocusTable
				m.Searching = false
				m.SearchInput.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.SearchInput, cmd = m.SearchInput.Update(msg)
				m.ApplyFilter()
				return m, cmd
			}
		}

		// Global dashboard hotkeys
		switch msg.String() {
		case "?":
			m.HelpModal.Toggle()
			return m, nil

		case "g":
			m.GeneratorModal.Open()
			return m, nil

		case "a":
			recType := domain.RecordTypeLogin
			if m.CurrentCategory.RecordType != "" {
				recType = m.CurrentCategory.RecordType
			}
			return m, func() tea.Msg {
				return OpenEditorMsg{
					Mode:       EditorModeCreate,
					RecordType: recType,
				}
			}

		case "e":
			if m.Detail.Record != nil {
				rec := m.Detail.Record
				return m, func() tea.Msg {
					return OpenEditorMsg{
						Mode:   EditorModeEdit,
						Record: rec,
					}
				}
			}

		case "d":
			if m.Detail.Record != nil {
				if m.Detail.Record.IsDeleted() || m.CurrentCategory.IsTrash {
					m.ConfirmModal.Open(
						ModalActionPurge,
						"Permanent Purge",
						fmt.Sprintf("Permanently delete %q? This cannot be undone! [y/N]", m.Detail.Record.Title),
						m.Detail.Record.ID,
					)
				} else {
					m.ConfirmModal.Open(
						ModalActionTrash,
						"Move to Trash",
						fmt.Sprintf("Move %q to Trash? [y/N]", m.Detail.Record.Title),
						m.Detail.Record.ID,
					)
				}
				return m, nil
			}

		case "r":
			if m.Detail.Record != nil && (m.Detail.Record.IsDeleted() || m.CurrentCategory.IsTrash) {
				m.ConfirmModal.Open(
					ModalActionRestore,
					"Restore Record",
					fmt.Sprintf("Restore %q to active vault? [y/N]", m.Detail.Record.Title),
					m.Detail.Record.ID,
				)
				return m, nil
			}

		case "/":
			m.Focus = FocusSearch
			m.Searching = true
			m.SearchInput.Focus()
			return m, textinput.Blink

		case "tab":
			if m.Focus == FocusSidebar {
				m.Focus = FocusTable
				m.Sidebar.SetFocus(false)
				m.Table.Focus()
			} else {
				m.Focus = FocusSidebar
				m.Sidebar.SetFocus(true)
				m.Table.Blur()
			}
			return m, nil

		case "h":
			if m.Focus == FocusTable {
				m.Focus = FocusSidebar
				m.Sidebar.SetFocus(true)
				m.Table.Blur()
				return m, nil
			}

		case "l":
			if m.Focus == FocusSidebar {
				m.Focus = FocusTable
				m.Sidebar.SetFocus(false)
				m.Table.Focus()
				return m, nil
			}

		case "H":
			// Open history modal
			if m.Detail.Record != nil {
				if err := m.History.LoadRecordHistory(m.Detail.Record.ID, m.Detail.Record.Title); err == nil {
					m.Focus = FocusHistory
					return m, nil
				}
			}

		case "v":
			m.Detail.ToggleSecret()
			return m, nil

		case "c":
			// Copy primary secret
			if m.Detail.Record != nil {
				secret := m.extractPrimarySecret(m.Detail.Record)
				if secret != "" {
					if err := m.Clipboard.Copy(context.Background(), secret, 45*time.Second); err == nil {
						return m, func() tea.Msg {
							return ToastMsg{Message: "✓ Secret copied to clipboard (auto-clears in 45s)"}
						}
					}
				}
			}

		case "u":
			// Copy username
			if m.Detail.Record != nil {
				user := m.extractUsername(m.Detail.Record)
				if user != "" {
					if err := m.Clipboard.Copy(context.Background(), user, 0); err == nil {
						return m, func() tea.Msg {
							return ToastMsg{Message: "✓ Username copied to clipboard"}
						}
					}
				}
			}
		}

		// Delegate to focused submodel
		switch m.Focus {
		case FocusSidebar:
			var (
				newModel tea.Model
				cmd      tea.Cmd
			)
			newModel, cmd = m.Sidebar.Update(msg)
			if sm, ok := newModel.(*SidebarModel); ok {
				m.Sidebar = sm
			}
			if cmd != nil {
				if selMsg, ok := cmd().(SidebarSelectMsg); ok {
					m.CurrentCategory = selMsg.Item
					m.ApplyFilter()
				} else {
					cmds = append(cmds, cmd)
				}
			}

		case FocusTable:
			var cmd tea.Cmd
			m.Table, cmd = m.Table.Update(msg)
			m.syncSelectedDetail()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *DashboardModel) extractPrimarySecret(r *domain.Record) string {
	switch p := r.Payload.(type) {
	case *domain.LoginPayload:
		return p.Password
	case domain.LoginPayload:
		return p.Password
	case *domain.NotePayload:
		return p.Content
	case domain.NotePayload:
		return p.Content
	case *domain.APIKeyPayload:
		return p.Secret
	case domain.APIKeyPayload:
		return p.Secret
	case *domain.CustomPayload:
		for _, f := range p.Fields {
			if f.Masked {
				return f.Value
			}
		}
		if len(p.Fields) > 0 {
			return p.Fields[0].Value
		}
	case domain.CustomPayload:
		for _, f := range p.Fields {
			if f.Masked {
				return f.Value
			}
		}
		if len(p.Fields) > 0 {
			return p.Fields[0].Value
		}
	}
	return ""
}

func (m *DashboardModel) extractUsername(r *domain.Record) string {
	switch p := r.Payload.(type) {
	case *domain.LoginPayload:
		return p.Username
	case domain.LoginPayload:
		return p.Username
	case *domain.APIKeyPayload:
		return p.Service
	case domain.APIKeyPayload:
		return p.Service
	}
	return ""
}

// View renders the split-pane dashboard workspace.
func (m *DashboardModel) View() string {
	if m.HelpModal.Active {
		return m.HelpModal.View()
	}
	if m.ConfirmModal.Active {
		return m.ConfirmModal.View()
	}
	if m.GeneratorModal.Active {
		return m.GeneratorModal.View()
	}
	if m.Focus == FocusHistory && m.History.Active {
		return lipgloss.Place(
			m.Width,
			m.Height,
			lipgloss.Center,
			lipgloss.Center,
			m.History.View(),
		)
	}

	sidebarView := m.Sidebar.View()

	// Right pane layout: Search (if active or typed), Table, Detail
	var rightComponents []string

	// Search bar line
	searchStyle := m.Theme.SubtitleStyle
	if m.Focus == FocusSearch {
		searchStyle = m.Theme.TitleStyle
	}
	searchView := searchStyle.Render(m.SearchInput.View())
	rightComponents = append(rightComponents, searchView)

	// Records table
	tableBorder := m.Theme.PanelStyle
	if m.Focus == FocusTable {
		tableBorder = m.Theme.PanelFocusStyle
	}
	rightComponents = append(rightComponents, tableBorder.Render(m.Table.View()))

	// Record detail inspector
	rightComponents = append(rightComponents, m.Detail.View())

	// Toast alert banner if active
	if m.ToastMessage != "" {
		toast := m.Theme.SuccessBadgeStyle.Render(m.ToastMessage)
		rightComponents = append(rightComponents, toast)
	}

	rightPane := lipgloss.JoinVertical(lipgloss.Left, rightComponents...)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		sidebarView,
		"  ",
		rightPane,
	)
}
