package comments

import "github.com/lakerszhy/thn/domain"

type node struct {
	comment  domain.Comment
	parentID int64
	children []int64
	loaded   bool
	loading  bool
	expanded bool
	err      error
}

type visible struct {
	id    int64
	depth int
	line  int
}

type fetchRequest struct {
	parentID int64
	ids      []int64
	ok       bool
}

type tree struct {
	itemID int64

	roots []int64
	nodes map[int64]*node

	selectedID int64
	visible    []visible
}

func newTree(itemID int64) *tree {
	return &tree{
		itemID: itemID,
		nodes:  make(map[int64]*node),
	}
}

func (t *tree) RootCount() int {
	return len(t.roots)
}

func (t *tree) SelectedID() int64 {
	return t.selectedID
}

func (t *tree) Node(id int64) *node {
	return t.nodes[id]
}

func (t *tree) SetRoots(comments []domain.Comment) {
	t.roots = t.roots[:0]
	for _, comment := range comments {
		t.upsertComment(comment, t.itemID)
		t.roots = append(t.roots, comment.ID)
	}

	if t.selectedID == 0 && len(t.roots) > 0 {
		t.selectedID = t.roots[0]
	}
	t.rebuildVisible()
}

func (t *tree) StartLoading(parentID int64) {
	if node := t.nodes[parentID]; node != nil {
		node.loading = true
		node.err = nil
	}
}

func (t *tree) SetChildren(parentID int64, comments []domain.Comment) {
	parent := t.nodes[parentID]
	if parent == nil {
		return
	}

	parent.children = parent.children[:0]
	parent.loaded = true
	parent.loading = false
	parent.err = nil
	for _, comment := range comments {
		t.upsertComment(comment, parent.comment.ID)
		parent.children = append(parent.children, comment.ID)
	}
	t.rebuildVisible()
}

func (t *tree) FailLoading(parentID int64, err error) {
	if node := t.nodes[parentID]; node != nil {
		node.loading = false
		node.err = err
	}
}

func (t *tree) ToggleSelected() fetchRequest {
	node := t.nodes[t.selectedID]
	if node == nil || len(node.comment.KIDs) == 0 {
		return fetchRequest{}
	}

	if node.expanded {
		node.expanded = false
		t.rebuildVisible()
		return fetchRequest{}
	}

	node.expanded = true
	t.rebuildVisible()
	if node.loaded || node.loading {
		return fetchRequest{}
	}

	return fetchRequest{
		parentID: node.comment.ID,
		ids:      node.comment.KIDs,
		ok:       true,
	}
}

func (t *tree) SelectVisible(delta int) {
	t.rebuildVisible()

	index := t.visibleIndex(t.selectedID)
	if index == -1 {
		return
	}

	index += delta
	if index < 0 || index >= len(t.visible) {
		return
	}

	t.selectedID = t.visible[index].id
}

func (t *tree) SelectParent() {
	node := t.nodes[t.selectedID]
	if node == nil || node.parentID == t.itemID {
		return
	}

	t.selectedID = node.parentID
}

func (t *tree) SelectSibling(delta int) {
	node := t.nodes[t.selectedID]
	if node == nil {
		return
	}

	siblings := t.roots
	if parent := t.nodes[node.parentID]; parent != nil {
		siblings = parent.children
	}

	for i, id := range siblings {
		if id != t.selectedID {
			continue
		}

		next := i + delta
		if next < 0 || next >= len(siblings) {
			return
		}

		t.selectedID = siblings[next]
		return
	}
}

func (t *tree) SelectFirst() {
	t.rebuildVisible()
	if len(t.visible) > 0 {
		t.selectedID = t.visible[0].id
	}
}

func (t *tree) SelectLast() {
	t.rebuildVisible()
	if len(t.visible) > 0 {
		t.selectedID = t.visible[len(t.visible)-1].id
	}
}

func (t *tree) Visible() []visible {
	return t.visible
}

func (t *tree) SetVisibleLine(id int64, line int) {
	for i := range t.visible {
		if t.visible[i].id == id {
			t.visible[i].line = line
			return
		}
	}
}

func (t *tree) HasLoading() bool {
	for _, node := range t.nodes {
		if node.loading {
			return true
		}
	}
	return false
}

func (t *tree) upsertComment(comment domain.Comment, parentID int64) {
	n, ok := t.nodes[comment.ID]
	if !ok {
		t.nodes[comment.ID] = &node{
			comment:  comment,
			parentID: parentID,
		}
		return
	}

	n.comment = comment
	n.parentID = parentID
}

func (t *tree) rebuildVisible() {
	t.visible = t.visible[:0]
	for _, id := range t.roots {
		t.appendVisible(id, 0)
	}
}

func (t *tree) appendVisible(id int64, depth int) {
	node := t.nodes[id]
	if node == nil {
		return
	}

	t.visible = append(t.visible, visible{id: id, depth: depth})
	if !node.expanded {
		return
	}

	for _, childID := range node.children {
		t.appendVisible(childID, depth+1)
	}
}

func (t *tree) visibleIndex(id int64) int {
	for i, visible := range t.visible {
		if visible.id == id {
			return i
		}
	}
	return -1
}
