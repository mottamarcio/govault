package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mottamarcio/govault/internal/domain"
	"github.com/mottamarcio/govault/internal/tui/theme"
)

// SidebarItemType defines category vs custom tag items.
type SidebarItemType int

const (
	SidebarItemCategory SidebarItemType = iota
	SidebarItemTag
)

// SidebarItem represents an entry in the sidebar navigation tree.
type SidebarItem struct {
	Type       SidebarItemType
	Label      string
	Icon       string
	RecordType domain.RecordType
	Tag        string
	IsTrash    bool
	Count      int
}

// SidebarSelectMsg is emitted when an item in the sidebar is selected.
type SidebarSelectMsg struct {
	Item SidebarItem
}

// SidebarModel manages left-pane category and tag filtering.
type SidebarModel struct {
	Theme    *theme.Theme
	Items    []SidebarItem
	Cursor   int
	Focused  bool
	Width    int
	Height   int
}

// NewSidebarModel creates a new SidebarModel with standard vault categories.
func NewSidebarModel(th *theme.Theme) *SidebarModel {
	items := []SidebarItem{
		{Type: SidebarItemCategory, Label: "All Items", Icon: "📂", RecordType: ""},
		{Type: SidebarItemCategory, Label: "Logins", Icon: "🔑", RecordType: domain.RecordTypeLogin},
		{Type: SidebarItemCategory, Label: "Notes", Icon: "📝", RecordType: domain.RecordTypeNote},
		{Type: SidebarItemCategory, Label: "API Keys", Icon: "⚡", RecordType: domain.RecordTypeAPIKey},
		{Type: SidebarItemCategory, Label: "Custom", Icon: "📦", RecordType: domain.RecordTypeCustom},
		{Type: SidebarItemCategory, Label: "Trash", Icon: "🗑️", RecordType: "", IsTrash: true},
	}

	return &SidebarModel{
		Theme:   th,
		Items:   items,
		Cursor:  0,
		Focused: true,
		Width:   24,
		Height:  20,
	}
}

// Init initializes the sidebar model.
func (m *SidebarModel) Init() tea.Cmd {
	return nil
}

// SetDimensions updates width and height of the sidebar.
func (m *SidebarModel) SetDimensions(w, h int) {
	m.Width = w
	m.Height = h
}

// SetFocus sets active panel focus state.
func (m *SidebarModel) SetFocus(focused bool) {
	m.Focused = focused
}

// SetCountsAndTags updates category counts and dynamically lists custom tags.
func (m *SidebarModel) SetCountsAndTags(records []*domain.Record) {
	counts := make(map[string]int)
	tagCounts := make(map[string]int)

	allCount := 0
	trashCount := 0

	for _, r := range records {
		if r == nil {
			continue
		}
		if r.IsDeleted() {
			trashCount++
			continue
		}
		allCount++
		counts[string(r.Type)]++
		for _, tag := range r.Tags {
			tagCounts[tag]++
		}
	}

	// Update base categories
	m.Items[0].Count = allCount
	m.Items[1].Count = counts[string(domain.RecordTypeLogin)]
	m.Items[2].Count = counts[string(domain.RecordTypeNote)]
	m.Items[3].Count = counts[string(domain.RecordTypeAPIKey)]
	m.Items[4].Count = counts[string(domain.RecordTypeCustom)]
	m.Items[5].Count = trashCount

	// Rebuild items with custom tags
	baseItems := m.Items[:6]
	var tagItems []SidebarItem
	for tag, cnt := range tagCounts {
		tagItems = append(tagItems, SidebarItem{
			Type:  SidebarItemTag,
			Label: "# " + tag,
			Icon:  "🏷️",
			Tag:   tag,
			Count: cnt,
		})
	}

	m.Items = append(baseItems, tagItems...)
	if m.Cursor >= len(m.Items) {
		m.Cursor = len(m.Items) - 1
	}
}

// SelectedItem returns the currently highlighted sidebar item.
func (m *SidebarModel) SelectedItem() SidebarItem {
	if len(m.Items) == 0 {
		return SidebarItem{}
	}
	if m.Cursor < 0 {
		m.Cursor = 0
	}
	if m.Cursor >= len(m.Items) {
		m.Cursor = len(m.Items) - 1
	}
	return m.Items[m.Cursor]
}

// Update handles navigation keys for SidebarModel.
func (m *SidebarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.Focused {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
				return m, func() tea.Msg { return SidebarSelectMsg{Item: m.SelectedItem()} }
			}
		case "down", "j":
			if m.Cursor < len(m.Items)-1 {
				m.Cursor++
				return m, func() tea.Msg { return SidebarSelectMsg{Item: m.SelectedItem()} }
			}
		case "enter":
			return m, func() tea.Msg { return SidebarSelectMsg{Item: m.SelectedItem()} }
		}
	}
	return m, nil
}

// View renders the sidebar panel.
func (m *SidebarModel) View() string {
	var lines []string

	header := m.Theme.HeaderBannerStyle.Render("CATEGORIES")
	lines = append(lines, header, "")

	for i, item := range m.Items {
		var line string
		countBadge := fmt.Sprintf("(%d)", item.Count)

		itemText := fmt.Sprintf("%s %s", item.Icon, item.Label)
		// Truncate if too long
		maxLabelWidth := m.Width - 8
		if maxLabelWidth > 0 && len(itemText) > maxLabelWidth {
			itemText = itemText[:maxLabelWidth-1] + "…"
		}

		if i == m.Cursor {
			if m.Focused {
				line = lipgloss.NewStyle().
					Bold(true).
					Foreground(m.Theme.Background).
					Background(m.Theme.Primary).
					Width(m.Width - 4).
					Render(fmt.Sprintf(" ▸ %s %s", itemText, countBadge))
			} else {
				line = lipgloss.NewStyle().
					Bold(true).
					Foreground(m.Theme.Text).
					Background(m.Theme.Surface).
					Width(m.Width - 4).
					Render(fmt.Sprintf("   %s %s", itemText, countBadge))
			}
		} else {
			line = lipgloss.NewStyle().
				Foreground(m.Theme.TextMuted).
				Width(m.Width - 4).
				Render(fmt.Sprintf("   %s %s", itemText, countBadge))
		}

		lines = append(lines, line)
		if i == 5 && len(m.Items) > 6 {
			lines = append(lines, "", m.Theme.SubtitleStyle.Render("TAGS"), "")
		}
	}

	content := strings.Join(lines, "\n")
	borderStyle := m.Theme.PanelStyle
	if m.Focused {
		borderStyle = m.Theme.PanelFocusStyle
	}

	return borderStyle.
		Width(m.Width).
		Height(m.Height).
		Render(content)
}
