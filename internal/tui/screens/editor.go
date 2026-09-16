package screens

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mottamarcio/govault/internal/domain"
	"github.com/mottamarcio/govault/internal/service"
	"github.com/mottamarcio/govault/internal/tui/theme"
)

// EditorMode defines creation vs editing of records.
type EditorMode int

const (
	EditorModeCreate EditorMode = iota
	EditorModeEdit
)

// EditorSaveMsg is emitted when a record is successfully saved.
type EditorSaveMsg struct {
	Record *domain.Record
}

// EditorCancelMsg is emitted when the editor is canceled.
type EditorCancelMsg struct{}

// OpenEditorMsg requests switching to the record form editor.
type OpenEditorMsg struct {
	Mode       EditorMode
	Record     *domain.Record
	RecordType domain.RecordType
}

// EditorCustomField tracks a dynamic custom field row.
type EditorCustomField struct {
	Key textinput.Model
	Val textinput.Model
}

// EditorFocus identifies the focused control within the editor form.
type EditorFocus int

const (
	FocusTypeSelector EditorFocus = iota
	FocusTitle
	FocusUsernameService
	FocusPasswordSecret
	FocusURI
	FocusTags
	FocusNotes
	FocusCustomStart
)

var recordTypes = []domain.RecordType{
	domain.RecordTypeLogin,
	domain.RecordTypeNote,
	domain.RecordTypeAPIKey,
	domain.RecordTypeCustom,
}

// EditorModel manages record creation and modification.
type EditorModel struct {
	Theme          *theme.Theme
	RecordService  *service.RecordService
	Mode           EditorMode
	RecordID       string
	OriginalRecord *domain.Record
	TypeIdx        int
	Active         bool

	// Inputs
	TitleInput      textinput.Model
	UsernameInput   textinput.Model
	SecretInput     textinput.Model
	SecretMasked    bool
	URIInput        textinput.Model
	TagsInput       textinput.Model
	NotesArea       textarea.Model
	CustomFields    []EditorCustomField

	FocusIndex      int // Flattened navigation index
	ValidationError string

	// Modals
	DiscardModal *ModalConfirm
	Generator    *GeneratorModel

	Width  int
	Height int
}

// NewEditorModel creates a new EditorModel.
func NewEditorModel(th *theme.Theme, rs *service.RecordService) *EditorModel {
	m := &EditorModel{
		Theme:         th,
		RecordService: rs,
		Mode:          EditorModeCreate,
		RecordID:      "",
		TypeIdx:       0,
		Active:        false,
		SecretMasked:  true,
		DiscardModal:  NewModalConfirm(th),
		Generator:     NewGeneratorModel(th),
		Width:         80,
		Height:        24,
	}

	m.initInputs()
	return m
}

func (m *EditorModel) initInputs() {
	th := m.Theme

	ti := textinput.New()
	ti.Placeholder = "e.g. GitHub, AWS Prod, Personal Wi-Fi"
	ti.Prompt = "Title: "
	ti.PromptStyle = th.InputPromptStyle
	ti.TextStyle = lipgloss.NewStyle().Foreground(th.Text)
	ti.CharLimit = 120
	m.TitleInput = ti

	userTi := textinput.New()
	userTi.Placeholder = "username / email / service"
	userTi.Prompt = "User/Identity: "
	userTi.PromptStyle = th.InputPromptStyle
	userTi.TextStyle = lipgloss.NewStyle().Foreground(th.Text)
	userTi.CharLimit = 120
	m.UsernameInput = userTi

	secTi := textinput.New()
	secTi.Placeholder = "password or secret token (Ctrl+G to generate)"
	secTi.Prompt = "Password/Secret: "
	secTi.PromptStyle = th.InputPromptStyle
	secTi.TextStyle = lipgloss.NewStyle().Foreground(th.Text)
	secTi.EchoMode = textinput.EchoPassword
	secTi.EchoCharacter = '•'
	secTi.CharLimit = 256
	m.SecretInput = secTi
	m.SecretMasked = true

	uriTi := textinput.New()
	uriTi.Placeholder = "https://github.com/login"
	uriTi.Prompt = "URI / URL: "
	uriTi.PromptStyle = th.InputPromptStyle
	uriTi.TextStyle = lipgloss.NewStyle().Foreground(th.Text)
	uriTi.CharLimit = 200
	m.URIInput = uriTi

	tagsTi := textinput.New()
	tagsTi.Placeholder = "work, dev, production (comma separated)"
	tagsTi.Prompt = "Tags: "
	tagsTi.PromptStyle = th.InputPromptStyle
	tagsTi.TextStyle = lipgloss.NewStyle().Foreground(th.Text)
	tagsTi.CharLimit = 150
	m.TagsInput = tagsTi

	ta := textarea.New()
	ta.Placeholder = "Secure multi-line notes..."
	ta.Prompt = "Notes: \n"
	ta.ShowLineNumbers = false
	ta.SetHeight(4)
	m.NotesArea = ta

	m.CustomFields = nil
	m.FocusIndex = 1 // Focus Title by default
	m.ValidationError = ""
}

// OpenCreate configures the editor for a new record.
func (m *EditorModel) OpenCreate(initialType domain.RecordType) {
	m.initInputs()
	m.Mode = EditorModeCreate
	m.RecordID = ""
	m.OriginalRecord = nil
	m.TypeIdx = 0
	for i, t := range recordTypes {
		if t == initialType {
			m.TypeIdx = i
			break
		}
	}
	m.Active = true
	m.FocusIndex = 1
	m.updateFocus()
}

// OpenEdit configures the editor with an existing record's contents.
func (m *EditorModel) OpenEdit(r *domain.Record) {
	m.initInputs()
	m.Mode = EditorModeEdit
	m.RecordID = r.ID
	m.OriginalRecord = r
	m.Active = true

	m.TitleInput.SetValue(r.Title)
	m.TagsInput.SetValue(strings.Join(r.Tags, ", "))

	for i, t := range recordTypes {
		if t == r.Type {
			m.TypeIdx = i
			break
		}
	}

	// Populate payload fields
	switch p := r.Payload.(type) {
	case *domain.LoginPayload:
		m.UsernameInput.SetValue(p.Username)
		m.SecretInput.SetValue(p.Password)
		m.URIInput.SetValue(p.URI)
		m.NotesArea.SetValue(p.Notes)
		m.loadCustomFields(p.CustomFields)
	case domain.LoginPayload:
		m.UsernameInput.SetValue(p.Username)
		m.SecretInput.SetValue(p.Password)
		m.URIInput.SetValue(p.URI)
		m.NotesArea.SetValue(p.Notes)
		m.loadCustomFields(p.CustomFields)

	case *domain.NotePayload:
		m.NotesArea.SetValue(p.Content)
		m.loadCustomFields(p.CustomFields)
	case domain.NotePayload:
		m.NotesArea.SetValue(p.Content)
		m.loadCustomFields(p.CustomFields)

	case *domain.APIKeyPayload:
		m.UsernameInput.SetValue(p.Service)
		m.URIInput.SetValue(p.Key)
		m.SecretInput.SetValue(p.Secret)
		m.NotesArea.SetValue(p.Notes)
		m.loadCustomFields(p.CustomFields)
	case domain.APIKeyPayload:
		m.UsernameInput.SetValue(p.Service)
		m.URIInput.SetValue(p.Key)
		m.SecretInput.SetValue(p.Secret)
		m.NotesArea.SetValue(p.Notes)
		m.loadCustomFields(p.CustomFields)

	case *domain.CustomPayload:
		m.NotesArea.SetValue(p.Notes)
		m.loadCustomFields(p.Fields)
	case domain.CustomPayload:
		m.NotesArea.SetValue(p.Notes)
		m.loadCustomFields(p.Fields)
	}

	m.FocusIndex = 1
	m.updateFocus()
}

func (m *EditorModel) loadCustomFields(fields []domain.Field) {
	m.CustomFields = nil
	for _, f := range fields {
		m.AddCustomFieldWithValues(f.Key, f.Value)
	}
}

// AddCustomField appends an empty custom field row.
func (m *EditorModel) AddCustomField() {
	m.AddCustomFieldWithValues("", "")
}

// AddCustomFieldWithValues appends a custom field row with initial key and value.
func (m *EditorModel) AddCustomFieldWithValues(k, v string) {
	th := m.Theme
	kIn := textinput.New()
	kIn.Placeholder = "Key / Label"
	kIn.Prompt = "K: "
	kIn.PromptStyle = th.InputPromptStyle
	kIn.TextStyle = lipgloss.NewStyle().Foreground(th.Text)
	kIn.CharLimit = 50
	kIn.Width = 16
	kIn.SetValue(k)

	vIn := textinput.New()
	vIn.Placeholder = "Value"
	vIn.Prompt = "V: "
	vIn.PromptStyle = th.InputPromptStyle
	vIn.TextStyle = lipgloss.NewStyle().Foreground(th.Text)
	vIn.CharLimit = 150
	vIn.Width = 32
	vIn.SetValue(v)

	m.CustomFields = append(m.CustomFields, EditorCustomField{Key: kIn, Val: vIn})
}

// RemoveFocusedCustomField removes the custom field row at the current focus.
func (m *EditorModel) RemoveFocusedCustomField() {
	if len(m.CustomFields) == 0 {
		return
	}
	customBase := int(FocusCustomStart)
	if m.FocusIndex >= customBase && m.FocusIndex < customBase+len(m.CustomFields)*2 {
		idx := (m.FocusIndex - customBase) / 2
		if idx >= 0 && idx < len(m.CustomFields) {
			m.CustomFields = append(m.CustomFields[:idx], m.CustomFields[idx+1:]...)
			if m.FocusIndex >= m.maxFocusIndex() {
				m.FocusIndex = m.maxFocusIndex() - 1
			}
			m.updateFocus()
		}
	}
}

// Reset clears all input fields and sensitive data.
func (m *EditorModel) Reset() {
	m.Active = false
	m.TitleInput.SetValue("")
	m.UsernameInput.SetValue("")
	m.SecretInput.SetValue("")
	m.URIInput.SetValue("")
	m.TagsInput.SetValue("")
	m.NotesArea.SetValue("")
	m.CustomFields = nil
	m.ValidationError = ""
	m.OriginalRecord = nil
	m.RecordID = ""
	m.DiscardModal.Close()
	m.Generator.Close()
}

// SetDimensions updates dimensions for responsive rendering.
func (m *EditorModel) SetDimensions(w, h int) {
	m.Width = w
	m.Height = h
	m.DiscardModal.SetDimensions(w, h)
	m.Generator.SetDimensions(w, h)

	contentWidth := w - 8
	if contentWidth < 30 {
		contentWidth = 30
	}
	m.TitleInput.Width = contentWidth
	m.UsernameInput.Width = contentWidth
	m.SecretInput.Width = contentWidth
	m.URIInput.Width = contentWidth
	m.TagsInput.Width = contentWidth
	m.NotesArea.SetWidth(contentWidth)
}

// IsDirty returns true if any input field differs from original state.
func (m *EditorModel) IsDirty() bool {
	if m.Mode == EditorModeCreate {
		return m.TitleInput.Value() != "" ||
			m.UsernameInput.Value() != "" ||
			m.SecretInput.Value() != "" ||
			m.URIInput.Value() != "" ||
			m.TagsInput.Value() != "" ||
			m.NotesArea.Value() != "" ||
			len(m.CustomFields) > 0
	}

	// In edit mode, compare with OriginalRecord
	if m.OriginalRecord == nil {
		return false
	}
	if m.TitleInput.Value() != m.OriginalRecord.Title {
		return true
	}
	if m.TagsInput.Value() != strings.Join(m.OriginalRecord.Tags, ", ") {
		return true
	}
	return true // Assume dirty if modified in edit mode
}

func (m *EditorModel) maxFocusIndex() int {
	// FocusTypeSelector (0), Title (1), Username (2), Secret (3), URI (4), Tags (5), Notes (6),
	// Custom fields (7 .. 7 + 2*len(CustomFields) - 1),
	// ButtonSave, ButtonCancel
	return int(FocusCustomStart) + len(m.CustomFields)*2 + 2
}

func (m *EditorModel) updateFocus() {
	m.TitleInput.Blur()
	m.UsernameInput.Blur()
	m.SecretInput.Blur()
	m.URIInput.Blur()
	m.TagsInput.Blur()
	m.NotesArea.Blur()
	for i := range m.CustomFields {
		m.CustomFields[i].Key.Blur()
		m.CustomFields[i].Val.Blur()
	}

	switch EditorFocus(m.FocusIndex) {
	case FocusTitle:
		m.TitleInput.Focus()
	case FocusUsernameService:
		m.UsernameInput.Focus()
	case FocusPasswordSecret:
		m.SecretInput.Focus()
	case FocusURI:
		m.URIInput.Focus()
	case FocusTags:
		m.TagsInput.Focus()
	case FocusNotes:
		m.NotesArea.Focus()
	default:
		customBase := int(FocusCustomStart)
		if m.FocusIndex >= customBase && m.FocusIndex < customBase+len(m.CustomFields)*2 {
			idx := (m.FocusIndex - customBase) / 2
			isVal := (m.FocusIndex-customBase)%2 == 1
			if isVal {
				m.CustomFields[idx].Val.Focus()
			} else {
				m.CustomFields[idx].Key.Focus()
			}
		}
	}
}

// Init initializes child components.
func (m *EditorModel) Init() tea.Cmd {
	return nil
}

// Update handles editor keystrokes and modal transitions.
func (m *EditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.Active {
		return m, nil
	}

	// 1. Delegate to Discard Modal if active
	if m.DiscardModal.Active {
		_, cmd := m.DiscardModal.Update(msg)
		if cmd != nil {
			if confirmMsg, ok := cmd().(ModalConfirmMsg); ok {
				if confirmMsg.Confirmed {
					m.Reset()
					return m, func() tea.Msg { return EditorCancelMsg{} }
				}
			}
		}
		return m, cmd
	}

	// 2. Delegate to Generator Modal if active
	if m.Generator.Active {
		_, cmd := m.Generator.Update(msg)
		if cmd != nil {
			innerMsg := cmd()
			switch genMsg := innerMsg.(type) {
			case GeneratorResultMsg:
				m.SecretInput.SetValue(genMsg.Secret)
				return m, nil
			case GeneratorCloseMsg:
				return m, nil
			}
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.Reset()
			return m, func() tea.Msg { return EditorCancelMsg{} }

		case tea.KeyEsc:
			if m.IsDirty() {
				m.DiscardModal.Open(ModalActionDiscard, "Discard Changes", "Discard unsaved changes? [y/N]", "")
				return m, nil
			}
			m.Reset()
			return m, func() tea.Msg { return EditorCancelMsg{} }

		case tea.KeyCtrlS:
			return m.handleSave()

		case tea.KeyCtrlG:
			m.Generator.Open()
			return m, nil

		case tea.KeyCtrlV:
			m.SecretMasked = !m.SecretMasked
			if m.SecretMasked {
				m.SecretInput.EchoMode = textinput.EchoPassword
			} else {
				m.SecretInput.EchoMode = textinput.EchoNormal
			}
			return m, nil

		case tea.KeyCtrlN:
			m.AddCustomField()
			m.FocusIndex = int(FocusCustomStart) + (len(m.CustomFields)-1)*2
			m.updateFocus()
			return m, nil

		case tea.KeyCtrlD:
			m.RemoveFocusedCustomField()
			return m, nil

		case tea.KeyTab:
			m.FocusIndex = (m.FocusIndex + 1) % m.maxFocusIndex()
			m.updateFocus()
			return m, nil

		case tea.KeyShiftTab:
			max := m.maxFocusIndex()
			m.FocusIndex = (m.FocusIndex - 1 + max) % max
			m.updateFocus()
			return m, nil

		case tea.KeyEnter:
			btnSaveIdx := m.maxFocusIndex() - 2
			btnCancelIdx := m.maxFocusIndex() - 1
			if m.FocusIndex == btnSaveIdx {
				return m.handleSave()
			}
			if m.FocusIndex == btnCancelIdx {
				if m.IsDirty() {
					m.DiscardModal.Open(ModalActionDiscard, "Discard Changes", "Discard unsaved changes? [y/N]", "")
					return m, nil
				}
				m.Reset()
				return m, func() tea.Msg { return EditorCancelMsg{} }
			}

		case tea.KeyLeft, tea.KeyRight:
			if m.FocusIndex == 0 && m.Mode == EditorModeCreate {
				delta := 1
				if msg.Type == tea.KeyLeft {
					delta = -1
				}
				m.TypeIdx = (m.TypeIdx + delta + len(recordTypes)) % len(recordTypes)
				return m, nil
			}
		}
	}

	// Dispatch input updates to focused element
	var cmd tea.Cmd
	switch EditorFocus(m.FocusIndex) {
	case FocusTitle:
		m.TitleInput, cmd = m.TitleInput.Update(msg)
	case FocusUsernameService:
		m.UsernameInput, cmd = m.UsernameInput.Update(msg)
	case FocusPasswordSecret:
		m.SecretInput, cmd = m.SecretInput.Update(msg)
	case FocusURI:
		m.URIInput, cmd = m.URIInput.Update(msg)
	case FocusTags:
		m.TagsInput, cmd = m.TagsInput.Update(msg)
	case FocusNotes:
		m.NotesArea, cmd = m.NotesArea.Update(msg)
	default:
		customBase := int(FocusCustomStart)
		if m.FocusIndex >= customBase && m.FocusIndex < customBase+len(m.CustomFields)*2 {
			idx := (m.FocusIndex - customBase) / 2
			isVal := (m.FocusIndex-customBase)%2 == 1
			if isVal {
				m.CustomFields[idx].Val, cmd = m.CustomFields[idx].Val.Update(msg)
			} else {
				m.CustomFields[idx].Key, cmd = m.CustomFields[idx].Key.Update(msg)
			}
		}
	}

	return m, cmd
}

func (m *EditorModel) handleSave() (tea.Model, tea.Cmd) {
	title := strings.TrimSpace(m.TitleInput.Value())
	if title == "" {
		m.ValidationError = "Title cannot be empty."
		return m, nil
	}

	// Parse tags
	rawTags := strings.Split(m.TagsInput.Value(), ",")
	var tags []string
	for _, t := range rawTags {
		tClean := strings.TrimSpace(t)
		if tClean != "" {
			tags = append(tags, tClean)
		}
	}

	// Parse custom fields
	var customFields []domain.Field
	seenKeys := make(map[string]bool)
	for _, cf := range m.CustomFields {
		k := strings.TrimSpace(cf.Key.Value())
		v := strings.TrimSpace(cf.Val.Value())
		if k == "" && v == "" {
			continue
		}
		if k == "" {
			m.ValidationError = "Custom field key cannot be empty."
			return m, nil
		}
		kLower := strings.ToLower(k)
		if seenKeys[kLower] {
			m.ValidationError = fmt.Sprintf("Duplicate custom field key: '%s'", k)
			return m, nil
		}
		seenKeys[kLower] = true
		customFields = append(customFields, domain.Field{Key: k, Value: v})
	}

	// Build payload
	curType := recordTypes[m.TypeIdx]
	var payload any
	switch curType {
	case domain.RecordTypeLogin:
		payload = domain.LoginPayload{
			Username:     m.UsernameInput.Value(),
			Password:     m.SecretInput.Value(),
			URI:          m.URIInput.Value(),
			Notes:        m.NotesArea.Value(),
			CustomFields: customFields,
		}
	case domain.RecordTypeNote:
		payload = domain.NotePayload{
			Content:      m.NotesArea.Value(),
			CustomFields: customFields,
		}
	case domain.RecordTypeAPIKey:
		payload = domain.APIKeyPayload{
			Service:      m.UsernameInput.Value(),
			Key:          m.URIInput.Value(),
			Secret:       m.SecretInput.Value(),
			Notes:        m.NotesArea.Value(),
			CustomFields: customFields,
		}
	case domain.RecordTypeCustom:
		payload = domain.CustomPayload{
			Notes:  m.NotesArea.Value(),
			Fields: customFields,
		}
	}

	input := service.RecordInput{
		Title:   title,
		Type:    curType,
		Tags:    tags,
		Payload: payload,
	}

	var savedRecord *domain.Record
	var err error

	if m.Mode == EditorModeCreate {
		savedRecord, err = m.RecordService.Create(context.Background(), input)
	} else {
		savedRecord, err = m.RecordService.Update(context.Background(), m.RecordID, input)
	}

	if err != nil {
		m.ValidationError = fmt.Sprintf("Save failed: %v", err)
		return m, nil
	}

	m.Reset()
	return m, func() tea.Msg {
		return EditorSaveMsg{Record: savedRecord}
	}
}

// View renders the multi-field form or active child modal.
func (m *EditorModel) View() string {
	if !m.Active {
		return ""
	}

	if m.DiscardModal.Active {
		return m.DiscardModal.View()
	}
	if m.Generator.Active {
		return m.Generator.View()
	}

	var sections []string

	// Header Banner
	bannerText := " ✏️  ADD NEW SECRET "
	if m.Mode == EditorModeEdit {
		bannerText = fmt.Sprintf(" ✏️  EDIT RECORD: %s ", m.TitleInput.Value())
	}
	sections = append(sections, m.Theme.HeaderBannerStyle.Render(bannerText), "")

	// Validation Error Banner
	if m.ValidationError != "" {
		sections = append(sections, m.Theme.ErrorBadgeStyle.Render(" Error: "+m.ValidationError+" "), "")
	}

	// Type Selector
	curType := recordTypes[m.TypeIdx]
	var typeTabs []string
	for i, t := range recordTypes {
		tab := string(t)
		if i == m.TypeIdx {
			tab = fmt.Sprintf("[● %s]", tab)
			if m.FocusIndex == 0 {
				typeTabs = append(typeTabs, m.Theme.KeyHintKeyStyle.Render(tab))
			} else {
				typeTabs = append(typeTabs, m.Theme.TitleStyle.Render(tab))
			}
		} else {
			tab = fmt.Sprintf("[○ %s]", tab)
			typeTabs = append(typeTabs, m.Theme.StatusMutedStyle.Render(tab))
		}
	}
	typeRow := fmt.Sprintf("Secret Type:   %s", strings.Join(typeTabs, "  "))
	sections = append(sections, typeRow, "")

	// Form Inputs
	sections = append(sections, m.TitleInput.View())

	if curType == domain.RecordTypeLogin || curType == domain.RecordTypeAPIKey {
		sections = append(sections, m.UsernameInput.View())
		sections = append(sections, m.SecretInput.View())
		sections = append(sections, m.URIInput.View())
	}

	sections = append(sections, m.TagsInput.View())
	sections = append(sections, m.NotesArea.View())

	// Custom Fields
	if len(m.CustomFields) > 0 {
		sections = append(sections, "", m.Theme.SubtitleStyle.Render("--- Custom Fields (Ctrl+N Add, Ctrl+D Delete) ---"))
		for i, cf := range m.CustomFields {
			row := fmt.Sprintf("  #%d  %s   %s", i+1, cf.Key.View(), cf.Val.View())
			sections = append(sections, row)
		}
	}

	sections = append(sections, "")

	// Buttons
	btnSaveIdx := m.maxFocusIndex() - 2
	btnCancelIdx := m.maxFocusIndex() - 1

	btnSave := "[ Ctrl+S / Save ]"
	btnCancel := "[ Esc / Cancel ]"

	if m.FocusIndex == btnSaveIdx {
		btnSave = m.Theme.KeyHintKeyStyle.Render(btnSave)
	} else {
		btnSave = m.Theme.StatusMutedStyle.Render(btnSave)
	}
	if m.FocusIndex == btnCancelIdx {
		btnCancel = m.Theme.KeyHintKeyStyle.Render(btnCancel)
	} else {
		btnCancel = m.Theme.StatusMutedStyle.Render(btnCancel)
	}

	helpHint := m.Theme.StatusMutedStyle.Render("Hotkeys: Tab switch field • Ctrl+G generator • Ctrl+V reveal secret • Ctrl+N add field")
	buttonsRow := fmt.Sprintf("%s    %s\n%s", btnSave, btnCancel, helpHint)
	sections = append(sections, buttonsRow)

	content := strings.Join(sections, "\n")
	boxWidth := m.Width - 4
	if boxWidth < 40 {
		boxWidth = 40
	}

	return m.Theme.PanelFocusStyle.
		Width(boxWidth).
		Render(content)
}
