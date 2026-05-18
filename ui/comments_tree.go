package ui

import "github.com/lakerszhy/thn/domain"

type commentsTree struct {
	itemID int64

	roots []int64
	nodes map[int64]*commentNode

	selectedID int64
	visible    []visibleComment
}

type commentNode struct {
	comment  domain.Comment
	parentID int64
	children []int64
	loaded   bool
	loading  bool
	expanded bool
	err      error
}

type visibleComment struct {
	id    int64
	depth int
	line  int
}

type commentFetchRequest struct {
	parentID int64
	ids      []int64
	ok       bool
}

func newCommentsTree(itemID int64) *commentsTree {
	return &commentsTree{
		itemID: itemID,
		nodes:  make(map[int64]*commentNode),
	}
}

func (t *commentsTree) RootCount() int {
	return len(t.roots)
}

func (t *commentsTree) SelectedID() int64 {
	return t.selectedID
}

func (t *commentsTree) Node(id int64) *commentNode {
	return t.nodes[id]
}

func (t *commentsTree) SetRoots(comments []domain.Comment) {
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

func (t *commentsTree) StartLoading(parentID int64) {
	if node := t.nodes[parentID]; node != nil {
		node.loading = true
		node.err = nil
	}
}

func (t *commentsTree) SetChildren(parentID int64, comments []domain.Comment) {
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

func (t *commentsTree) FailLoading(parentID int64, err error) {
	if node := t.nodes[parentID]; node != nil {
		node.loading = false
		node.err = err
	}
}

func (t *commentsTree) ToggleSelected() commentFetchRequest {
	node := t.nodes[t.selectedID]
	if node == nil || len(node.comment.KIDs) == 0 {
		return commentFetchRequest{}
	}

	if node.expanded {
		node.expanded = false
		t.rebuildVisible()
		return commentFetchRequest{}
	}

	node.expanded = true
	t.rebuildVisible()
	if node.loaded || node.loading {
		return commentFetchRequest{}
	}

	return commentFetchRequest{
		parentID: node.comment.ID,
		ids:      node.comment.KIDs,
		ok:       true,
	}
}

func (t *commentsTree) SelectVisible(delta int) {
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

func (t *commentsTree) SelectParent() {
	node := t.nodes[t.selectedID]
	if node == nil || node.parentID == t.itemID {
		return
	}

	t.selectedID = node.parentID
}

func (t *commentsTree) SelectSibling(delta int) {
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

func (t *commentsTree) SelectFirst() {
	t.rebuildVisible()
	if len(t.visible) > 0 {
		t.selectedID = t.visible[0].id
	}
}

func (t *commentsTree) SelectLast() {
	t.rebuildVisible()
	if len(t.visible) > 0 {
		t.selectedID = t.visible[len(t.visible)-1].id
	}
}

func (t *commentsTree) Visible() []visibleComment {
	return t.visible
}

func (t *commentsTree) SetVisibleLine(id int64, line int) {
	for i := range t.visible {
		if t.visible[i].id == id {
			t.visible[i].line = line
			return
		}
	}
}

func (t *commentsTree) HasLoading() bool {
	for _, node := range t.nodes {
		if node.loading {
			return true
		}
	}
	return false
}

func (t *commentsTree) upsertComment(comment domain.Comment, parentID int64) {
	node, ok := t.nodes[comment.ID]
	if !ok {
		t.nodes[comment.ID] = &commentNode{
			comment:  comment,
			parentID: parentID,
		}
		return
	}

	node.comment = comment
	node.parentID = parentID
}

func (t *commentsTree) rebuildVisible() {
	t.visible = t.visible[:0]
	for _, id := range t.roots {
		t.appendVisible(id, 0)
	}
}

func (t *commentsTree) appendVisible(id int64, depth int) {
	node := t.nodes[id]
	if node == nil {
		return
	}

	t.visible = append(t.visible, visibleComment{id: id, depth: depth})
	if !node.expanded {
		return
	}

	for _, childID := range node.children {
		t.appendVisible(childID, depth+1)
	}
}

func (t *commentsTree) visibleIndex(id int64) int {
	for i, visible := range t.visible {
		if visible.id == id {
			return i
		}
	}
	return -1
}
