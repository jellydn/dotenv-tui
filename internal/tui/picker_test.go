package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPickerModelInit(t *testing.T) {
	model := PickerModel{
		items: []pickerItem{
			{text: "file1.env", filePath: "file1.env", isHeader: false},
			{text: "file2.env", filePath: "file2.env", isHeader: false},
		},
		selected: map[int]bool{0: true, 1: false},
		cursor:   0,
		mode:     GenerateExample,
		rootDir:  "/test",
	}

	cmd := model.Init()

	if cmd != nil {
		t.Errorf("PickerModel.Init() should return nil, got %v", cmd)
	}
}

func TestPickerModelUpdateWithInitMsg(t *testing.T) {
	model := PickerModel{}
	initMsg := pickerInitMsg{
		items: []pickerItem{
			{text: ".env", filePath: ".env", isHeader: false},
			{text: "test/.env", filePath: "test/.env", isHeader: false},
		},
		selected: map[int]bool{0: true, 1: true},
		mode:     GenerateEnv,
		rootDir:  "/project",
	}

	newModel, cmd := model.Update(initMsg)

	newPickerModel := newModel.(PickerModel)

	if len(newPickerModel.items) != 2 {
		t.Errorf("Update() items count = %d, expected 2", len(newPickerModel.items))
	}

	if len(newPickerModel.selected) != 2 {
		t.Errorf("Update() selected count = %d, expected 2", len(newPickerModel.selected))
	}

	if newPickerModel.mode != GenerateEnv {
		t.Errorf("Update() mode = %v, expected %v", newPickerModel.mode, GenerateEnv)
	}

	if newPickerModel.rootDir != "/project" {
		t.Errorf("Update() rootDir = %q, expected %q", newPickerModel.rootDir, "/project")
	}

	if cmd != nil {
		t.Errorf("Update() should return nil command, got %v", cmd)
	}
}

func TestPickerModelUpdateNavigation(t *testing.T) {
	tests := []struct {
		name           string
		initialCursor  int
		initialItems   []pickerItem
		keyMsg         tea.KeyMsg
		expectedCursor int
	}{
		{
			name:           "down key moves cursor down",
			initialCursor:  0,
			initialItems:   []pickerItem{{text: "file1.env", filePath: "file1.env", isHeader: false}, {text: "file2.env", filePath: "file2.env", isHeader: false}, {text: "file3.env", filePath: "file3.env", isHeader: false}},
			keyMsg:         tea.KeyMsg{Type: tea.KeyDown},
			expectedCursor: 1,
		},
		{
			name:           "j key moves cursor down",
			initialCursor:  0,
			initialItems:   []pickerItem{{text: "file1.env", filePath: "file1.env", isHeader: false}, {text: "file2.env", filePath: "file2.env", isHeader: false}},
			keyMsg:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			expectedCursor: 1,
		},
		{
			name:           "up key moves cursor up",
			initialCursor:  1,
			initialItems:   []pickerItem{{text: "file1.env", filePath: "file1.env", isHeader: false}, {text: "file2.env", filePath: "file2.env", isHeader: false}},
			keyMsg:         tea.KeyMsg{Type: tea.KeyUp},
			expectedCursor: 0,
		},
		{
			name:           "k key moves cursor up",
			initialCursor:  1,
			initialItems:   []pickerItem{{text: "file1.env", filePath: "file1.env", isHeader: false}, {text: "file2.env", filePath: "file2.env", isHeader: false}},
			keyMsg:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			expectedCursor: 0,
		},
		{
			name:           "up key at top stays at top",
			initialCursor:  0,
			initialItems:   []pickerItem{{text: "file1.env", filePath: "file1.env", isHeader: false}, {text: "file2.env", filePath: "file2.env", isHeader: false}},
			keyMsg:         tea.KeyMsg{Type: tea.KeyUp},
			expectedCursor: 0,
		},
		{
			name:           "down key at bottom stays at bottom",
			initialCursor:  1,
			initialItems:   []pickerItem{{text: "file1.env", filePath: "file1.env", isHeader: false}, {text: "file2.env", filePath: "file2.env", isHeader: false}},
			keyMsg:         tea.KeyMsg{Type: tea.KeyDown},
			expectedCursor: 1,
		},
		{
			name:           "down key with single file stays at 0",
			initialCursor:  0,
			initialItems:   []pickerItem{{text: "file1.env", filePath: "file1.env", isHeader: false}},
			keyMsg:         tea.KeyMsg{Type: tea.KeyDown},
			expectedCursor: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := PickerModel{
				items:    tt.initialItems,
				selected: make(map[int]bool),
				cursor:   tt.initialCursor,
			}

			newModel, cmd := model.Update(tt.keyMsg)

			newPickerModel := newModel.(PickerModel)

			if newPickerModel.cursor != tt.expectedCursor {
				t.Errorf("Update() cursor = %d, expected %d", newPickerModel.cursor, tt.expectedCursor)
			}

			if cmd != nil {
				t.Errorf("Update() should return nil command, got %v", cmd)
			}
		})
	}
}

func TestPickerModelUpdateToggleSelection(t *testing.T) {
	model := PickerModel{
		items: []pickerItem{
			{text: "file1.env", filePath: "file1.env", isHeader: false},
			{text: "file2.env", filePath: "file2.env", isHeader: false},
		},
		selected: map[int]bool{0: true, 1: false},
		cursor:   1,
	}

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	newModel, cmd := model.Update(spaceKey)

	newPickerModel := newModel.(PickerModel)

	if !newPickerModel.selected[1] {
		t.Errorf("Update() space key should toggle selection, expected index 1 to be true")
	}

	if cmd != nil {
		t.Errorf("Update() should return nil command, got %v", cmd)
	}
}

func TestPickerModelUpdateToggleSelectionFromTrue(t *testing.T) {
	model := PickerModel{
		items: []pickerItem{
			{text: "file1.env", filePath: "file1.env", isHeader: false},
			{text: "file2.env", filePath: "file2.env", isHeader: false},
		},
		selected: map[int]bool{0: true, 1: true},
		cursor:   1,
	}

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	newModel, cmd := model.Update(spaceKey)

	newPickerModel := newModel.(PickerModel)

	if newPickerModel.selected[1] {
		t.Errorf("Update() space key should toggle selection, expected index 1 to be false")
	}

	if cmd != nil {
		t.Errorf("Update() should return nil command, got %v", cmd)
	}
}

func TestPickerModelUpdateEnterWithSelection(t *testing.T) {
	model := PickerModel{
		items: []pickerItem{
			{text: ".env", filePath: ".env", isHeader: false},
			{text: "test/.env", filePath: "test/.env", isHeader: false},
			{text: "prod/.env", filePath: "prod/.env", isHeader: false},
		},
		selected: map[int]bool{0: true, 1: false, 2: true},
		cursor:   1,
		mode:     GenerateExample,
	}

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := model.Update(enterKey)

	if cmd == nil {
		t.Errorf("Update() enter key should return a command")
		return
	}

	msg := cmd()
	finishedMsg := msg.(PickerFinishedMsg)

	if len(finishedMsg.Selected) != 2 {
		t.Errorf("PickerFinishedMsg.Selected = %d files, expected 2", len(finishedMsg.Selected))
	}

	expectedFiles := map[string]bool{
		".env":      true,
		"prod/.env": true,
	}

	for _, file := range finishedMsg.Selected {
		if !expectedFiles[file] {
			t.Errorf("Unexpected selected file: %q", file)
		}
	}
}

func TestGroupFilesByDirectory(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected []pickerItem
	}{
		{
			name:  "files in different directories",
			files: []string{"apps/api/.env", "apps/web/.env", "services/auth/.env", "packages/db/.env"},
			expected: []pickerItem{
				{text: "apps/api", filePath: "", isHeader: true},
				{text: "apps/api/.env", filePath: "apps/api/.env", isHeader: false},
				{text: "apps/web", filePath: "", isHeader: true},
				{text: "apps/web/.env", filePath: "apps/web/.env", isHeader: false},
				{text: "packages/db", filePath: "", isHeader: true},
				{text: "packages/db/.env", filePath: "packages/db/.env", isHeader: false},
				{text: "services/auth", filePath: "", isHeader: true},
				{text: "services/auth/.env", filePath: "services/auth/.env", isHeader: false},
			},
		},
		{
			name:  "files in current directory",
			files: []string{".env", ".env.local"},
			expected: []pickerItem{
				{text: "Current Directory", filePath: "", isHeader: true},
				{text: ".env", filePath: ".env", isHeader: false},
				{text: ".env.local", filePath: ".env.local", isHeader: false},
			},
		},
		{
			name:     "empty list",
			files:    []string{},
			expected: []pickerItem{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := groupFilesByDirectory(tt.files)

			if len(result) != len(tt.expected) {
				t.Errorf("groupFilesByDirectory() returned %d items, expected %d", len(result), len(tt.expected))
				return
			}

			for i, item := range result {
				if i >= len(tt.expected) {
					t.Errorf("Result has more items than expected")
					break
				}
				expected := tt.expected[i]
				if item.text != expected.text || item.filePath != expected.filePath || item.isHeader != expected.isHeader {
					t.Errorf("Item %d mismatch:\n  got:      {text: %q, filePath: %q, isHeader: %v}\n  expected: {text: %q, filePath: %q, isHeader: %v}",
						i, item.text, item.filePath, item.isHeader, expected.text, expected.filePath, expected.isHeader)
				}
			}
		})
	}
}

func TestPickerModelUpdateEnterWithNoSelection(t *testing.T) {
	model := PickerModel{
		items: []pickerItem{
			{text: ".env", filePath: ".env", isHeader: false},
		},
		selected: map[int]bool{0: false},
		cursor:   0,
		mode:     GenerateEnv,
	}

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := model.Update(enterKey)

	if cmd != nil {
		t.Errorf("Update() enter key with no selection should return nil command, got %v", cmd)
	}
}

func TestPickerModelUpdateEmptyFilesNavigation(t *testing.T) {
	model := PickerModel{
		items:    []pickerItem{},
		selected: map[int]bool{},
		cursor:   0,
	}

	tests := []struct {
		name   string
		keyMsg tea.KeyMsg
	}{
		{name: "down key with empty files", keyMsg: tea.KeyMsg{Type: tea.KeyDown}},
		{name: "up key with empty files", keyMsg: tea.KeyMsg{Type: tea.KeyUp}},
		{name: "space key with empty files", keyMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newModel, cmd := model.Update(tt.keyMsg)

			newPickerModel := newModel.(PickerModel)

			if newPickerModel.cursor != 0 {
				t.Errorf("Update() cursor = %d, expected 0", newPickerModel.cursor)
			}

			if cmd != nil {
				t.Errorf("Update() should return nil command, got %v", cmd)
			}
		})
	}
}

func TestPickerModelScrolling(t *testing.T) {
	makeItems := func(n int) []pickerItem {
		items := make([]pickerItem, n)
		for i := range n {
			items[i] = pickerItem{text: fmt.Sprintf("file%d.env", i), filePath: fmt.Sprintf("file%d.env", i), isHeader: false}
		}
		return items
	}

	t.Run("cursor remains visible after scrolling down", func(t *testing.T) {
		items := makeItems(20)
		selected := make(map[int]bool)
		for i := range items {
			selected[i] = false
		}
		m := PickerModel{
			items:        items,
			selected:     selected,
			cursor:       0,
			windowHeight: 12,
		}

		for range 8 {
			newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
			m = newModel.(PickerModel)
		}

		if m.cursor != 8 {
			t.Errorf("cursor position = %d, expected 8", m.cursor)
		}

		visible := m.visibleLines()
		if m.cursor < m.offset || m.cursor >= m.offset+visible {
			t.Errorf("cursor at position %d is not visible in viewport (offset=%d, visible=%d)", m.cursor, m.offset, visible)
		}
	})

	t.Run("cursor remains visible after scrolling up", func(t *testing.T) {
		items := makeItems(20)
		selected := make(map[int]bool)
		for i := range items {
			selected[i] = false
		}
		m := PickerModel{
			items:        items,
			selected:     selected,
			cursor:       10,
			offset:       8,
			windowHeight: 12,
		}

		for range 5 {
			newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
			m = newModel.(PickerModel)
		}

		visible := m.visibleLines()
		if m.cursor < m.offset || m.cursor >= m.offset+visible {
			t.Errorf("cursor at position %d is not visible in viewport (offset=%d, visible=%d)", m.cursor, m.offset, visible)
		}
	})

	t.Run("window resize keeps cursor visible", func(t *testing.T) {
		items := makeItems(20)
		selected := make(map[int]bool)
		for i := range items {
			selected[i] = false
		}
		m := PickerModel{
			items:        items,
			selected:     selected,
			cursor:       19,
			offset:       18,
			windowHeight: 12,
		}

		newModel, _ := m.Update(tea.WindowSizeMsg{Height: 30})
		m = newModel.(PickerModel)

		visible := m.visibleLines()
		if m.cursor < m.offset || m.cursor >= m.offset+visible {
			t.Errorf("cursor at position %d is not visible after window resize (offset=%d, visible=%d)", m.cursor, m.offset, visible)
		}
	})

	t.Run("no scrolling when all items fit on screen", func(t *testing.T) {
		items := makeItems(3)
		selected := make(map[int]bool)
		for i := range items {
			selected[i] = false
		}
		m := PickerModel{
			items:        items,
			selected:     selected,
			cursor:       0,
			windowHeight: 30,
		}

		view := m.View()

		if strings.Contains(view, "more items") {
			t.Error("view should not show scroll indicators when all items fit on screen")
		}
	})

	t.Run("scroll indicators shown when items overflow viewport", func(t *testing.T) {
		items := makeItems(20)
		selected := make(map[int]bool)
		for i := range items {
			selected[i] = false
		}
		m := PickerModel{
			items:        items,
			selected:     selected,
			cursor:       10,
			offset:       5,
			windowHeight: 12,
		}

		view := m.View()

		if !strings.Contains(view, "more items above") {
			t.Error("view should show 'more items above' indicator when scrolled down")
		}
		if !strings.Contains(view, "more items below") {
			t.Error("view should show 'more items below' indicator when more items exist below viewport")
		}
	})

	t.Run("only top scroll indicator when at bottom", func(t *testing.T) {
		items := makeItems(20)
		selected := make(map[int]bool)
		for i := range items {
			selected[i] = false
		}
		m := PickerModel{
			items:        items,
			selected:     selected,
			cursor:       19,
			offset:       14,
			windowHeight: 12,
		}

		view := m.View()

		if !strings.Contains(view, "more items above") {
			t.Error("view should show 'more items above' when scrolled to bottom")
		}
		if strings.Contains(view, "more items below") {
			t.Error("view should not show 'more items below' when at the bottom")
		}
	})

	t.Run("only bottom scroll indicator when at top", func(t *testing.T) {
		items := makeItems(20)
		selected := make(map[int]bool)
		for i := range items {
			selected[i] = false
		}
		m := PickerModel{
			items:        items,
			selected:     selected,
			cursor:       5,
			offset:       0,
			windowHeight: 12,
		}

		view := m.View()

		if strings.Contains(view, "more items above") {
			t.Error("view should not show 'more items above' when at the top")
		}
		if !strings.Contains(view, "more items below") {
			t.Error("view should show 'more items below' when more items exist below")
		}
	})
}

func TestPickerModelSetWindowHeight(t *testing.T) {
	t.Run("window height is set correctly", func(t *testing.T) {
		m := &PickerModel{}

		m.SetWindowHeight(42)

		if m.windowHeight != 42 {
			t.Errorf("windowHeight = %d, expected 42", m.windowHeight)
		}
	})
}

func TestPickerModelVisibleLines(t *testing.T) {
	tests := []struct {
		name         string
		windowHeight int
		itemCount    int
		wantMin      int
		wantMax      int
	}{
		{
			name:         "small window shows all items when overhead is larger",
			windowHeight: 5,
			itemCount:    10,
			wantMin:      10,
			wantMax:      10,
		},
		{
			name:         "large window shows all items when they fit",
			windowHeight: 50,
			itemCount:    10,
			wantMin:      10,
			wantMax:      10,
		},
		{
			name:         "medium window shows subset when items overflow",
			windowHeight: 20,
			itemCount:    30,
			wantMin:      14,
			wantMax:      14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := make([]pickerItem, tt.itemCount)
			for i := range items {
				items[i] = pickerItem{text: fmt.Sprintf("item%d", i), filePath: fmt.Sprintf("item%d", i), isHeader: false}
			}
			m := PickerModel{
				items:        items,
				windowHeight: tt.windowHeight,
			}

			visible := m.visibleLines()

			if visible < tt.wantMin || visible > tt.wantMax {
				t.Errorf("visibleLines() = %d, expected range [%d, %d]", visible, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestPickerModelUpdateSelectAllToggle(t *testing.T) {
	tests := []struct {
		name                string
		initialSelection    map[int]bool
		expectedAfterToggle map[int]bool
		items               []pickerItem
	}{
		{
			name: "select all when none are selected",
			initialSelection: map[int]bool{
				1: false,
				2: false,
				4: false,
				5: false,
			},
			expectedAfterToggle: map[int]bool{
				1: true,
				2: true,
				4: true,
				5: true,
			},
			items: []pickerItem{
				{text: "Group 1", filePath: "", isHeader: true},
				{text: "file1.env", filePath: "file1.env", isHeader: false},
				{text: "file2.env", filePath: "file2.env", isHeader: false},
				{text: "Group 2", filePath: "", isHeader: true},
				{text: "file3.env", filePath: "file3.env", isHeader: false},
				{text: "file4.env", filePath: "file4.env", isHeader: false},
			},
		},
		{
			name: "deselect all when all are selected",
			initialSelection: map[int]bool{
				1: true,
				2: true,
			},
			expectedAfterToggle: map[int]bool{
				1: false,
				2: false,
			},
			items: []pickerItem{
				{text: "Group 1", filePath: "", isHeader: true},
				{text: "file1.env", filePath: "file1.env", isHeader: false},
				{text: "file2.env", filePath: "file2.env", isHeader: false},
			},
		},
		{
			name: "select all when some are selected",
			initialSelection: map[int]bool{
				1: true,
				2: false,
				3: true,
			},
			expectedAfterToggle: map[int]bool{
				1: true,
				2: true,
				3: true,
			},
			items: []pickerItem{
				{text: "Group 1", filePath: "", isHeader: true},
				{text: "file1.env", filePath: "file1.env", isHeader: false},
				{text: "file2.env", filePath: "file2.env", isHeader: false},
				{text: "file3.env", filePath: "file3.env", isHeader: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := PickerModel{
				items:    tt.items,
				selected: tt.initialSelection,
				cursor:   1,
			}

			aKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
			newModel, cmd := model.Update(aKey)

			newPickerModel := newModel.(PickerModel)

			for i := range tt.items {
				if !tt.items[i].isHeader {
					if newPickerModel.selected[i] != tt.expectedAfterToggle[i] {
						t.Errorf("Update() 'a' key: item %d selection = %v, expected %v",
							i, newPickerModel.selected[i], tt.expectedAfterToggle[i])
					}
				}
			}

			if cmd != nil {
				t.Errorf("Update() should return nil command, got %v", cmd)
			}
		})
	}
}

func TestPickerModelNavigationWithHeaders(t *testing.T) {
	tests := []struct {
		name           string
		initialCursor  int
		initialItems   []pickerItem
		keyMsg         tea.KeyMsg
		expectedCursor int
	}{
		{
			name:          "cursor skips headers when moving down",
			initialCursor: 0,
			initialItems: []pickerItem{
				{text: "Group 1", filePath: "", isHeader: true},
				{text: "file1.env", filePath: "file1.env", isHeader: false},
				{text: "file2.env", filePath: "file2.env", isHeader: false},
			},
			keyMsg:         tea.KeyMsg{Type: tea.KeyDown},
			expectedCursor: 1,
		},
		{
			name:          "cursor skips headers when moving up",
			initialCursor: 3,
			initialItems: []pickerItem{
				{text: "file1.env", filePath: "file1.env", isHeader: false},
				{text: "Group 1", filePath: "", isHeader: true},
				{text: "file2.env", filePath: "file2.env", isHeader: false},
				{text: "file3.env", filePath: "file3.env", isHeader: false},
			},
			keyMsg:         tea.KeyMsg{Type: tea.KeyUp},
			expectedCursor: 2,
		},
		{
			name:          "cursor stays at last selectable when header follows",
			initialCursor: 2,
			initialItems: []pickerItem{
				{text: "file1.env", filePath: "file1.env", isHeader: false},
				{text: "file2.env", filePath: "file2.env", isHeader: false},
				{text: "Group 1", filePath: "", isHeader: true},
			},
			keyMsg:         tea.KeyMsg{Type: tea.KeyDown},
			expectedCursor: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := PickerModel{
				items:    tt.initialItems,
				selected: make(map[int]bool),
				cursor:   tt.initialCursor,
			}

			newModel, cmd := model.Update(tt.keyMsg)

			newPickerModel := newModel.(PickerModel)

			if newPickerModel.cursor != tt.expectedCursor {
				t.Errorf("Update() cursor = %d, expected %d", newPickerModel.cursor, tt.expectedCursor)
			}

			if cmd != nil {
				t.Errorf("Update() should return nil command, got %v", cmd)
			}
		})
	}
}

func TestPickerModelInitMsgPositionsCursorAtFirstSelectable(t *testing.T) {
	tests := []struct {
		name           string
		items          []pickerItem
		expectedCursor int
	}{
		{
			name: "first item is selectable",
			items: []pickerItem{
				{text: "file1.env", filePath: "file1.env", isHeader: false},
				{text: "file2.env", filePath: "file2.env", isHeader: false},
			},
			expectedCursor: 0,
		},
		{
			name: "skips header to find first selectable",
			items: []pickerItem{
				{text: "Group 1", filePath: "", isHeader: true},
				{text: "file1.env", filePath: "file1.env", isHeader: false},
			},
			expectedCursor: 1,
		},
		{
			name: "multiple headers before first selectable",
			items: []pickerItem{
				{text: "Header 1", filePath: "", isHeader: true},
				{text: "Header 2", filePath: "", isHeader: true},
				{text: "file1.env", filePath: "file1.env", isHeader: false},
			},
			expectedCursor: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := PickerModel{}
			selected := make(map[int]bool)
			for i := range tt.items {
				if !tt.items[i].isHeader {
					selected[i] = false
				}
			}
			initMsg := pickerInitMsg{
				items:    tt.items,
				selected: selected,
				mode:     GenerateEnv,
				rootDir:  "/test",
			}

			newModel, cmd := model.Update(initMsg)

			newPickerModel := newModel.(PickerModel)

			if newPickerModel.cursor != tt.expectedCursor {
				t.Errorf("Update() with initMsg positioned cursor at %d, expected %d",
					newPickerModel.cursor, tt.expectedCursor)
			}

			if cmd != nil {
				t.Errorf("Update() should return nil command, got %v", cmd)
			}
		})
	}
}

func TestPickerModelUpdateQuit(t *testing.T) {
	tests := []struct {
		name   string
		keyMsg tea.KeyMsg
	}{
		{name: "q key returns nil command", keyMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}},
		{name: "esc key returns nil command", keyMsg: tea.KeyMsg{Type: tea.KeyEsc}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := PickerModel{
				items: []pickerItem{
					{text: ".env", filePath: ".env", isHeader: false},
				},
				selected: map[int]bool{0: true},
				cursor:   0,
			}

			newModel, cmd := model.Update(tt.keyMsg)

			_ = newModel.(PickerModel)

			if cmd != nil {
				t.Errorf("Update(%q) should return nil command, got %v", tt.name, cmd)
			}
		})
	}
}
