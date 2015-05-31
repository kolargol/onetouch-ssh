package main

import (
	"fmt"
	"github.com/authy/onetouch-ssh"
	"github.com/jroimartin/gocui"
)

var (
	listViewID         = "users-list"
	addUserViewID      = "add-user"
	addUserViewLabelID = "add-user-label"
	addUserViewInputID = "add-user-input"
)

// UsersList is a list of users.
type UsersList struct {
	gui       *gocui.Gui
	listeners []UsersListListener
}

// UsersListListener is a interface for listing users list events.
type UsersListListener interface {
	OnUserSelected(user *ssh.User)
	OnStartEditingUser(user *ssh.User)
}

// NewUsersList creates an instance of the UsersList
func NewUsersList(g *gocui.Gui) *UsersList {
	list := &UsersList{
		gui: g,
	}

	return list
}

func (list *UsersList) setHelp() {
	setHelp(list.gui, `enter: edit user | up/down: select user | ctrl-c close app`)
}

// AddListener adds a listener for list's events.
func (list *UsersList) AddListener(listener UsersListListener) {
	list.listeners = append(list.listeners, listener)
}

func (list *UsersList) drawLayout() {
	_, maxY := list.gui.Size()

	if v, err := list.gui.SetView(listViewID, -1, -1, 30, maxY-2); err != nil {
		v.Highlight = true
		v.Editable = false
		v.Wrap = false

		manager := ssh.NewUsersManager()
		for _, user := range manager.Users() {
			fmt.Fprintln(v, user.Username)
		}
	}
}

func (list *UsersList) view() *gocui.View {
	v, err := list.gui.View(listViewID)
	if err != nil {
		panic(err)
	}

	return v
}

func (list *UsersList) usernameToAdd() string {
	v, err := list.gui.View(addUserViewInputID)
	if err != nil {
		panic(err)
	}

	username, _ := v.Line(-1)

	return username
}

func (list *UsersList) setupKeyBindings() {
	list.gui.SetKeybinding(addUserViewInputID, gocui.KeyEnter, gocui.ModNone, list.validateAndAddUser)
	list.gui.SetKeybinding(addUserViewInputID, gocui.KeyCtrlU, gocui.ModNone, clearView)

	list.gui.SetKeybinding(listViewID, gocui.KeyEnter, gocui.ModNone, list.editCurrentUser)
	list.gui.SetKeybinding(listViewID, gocui.KeyArrowDown, gocui.ModNone, list.cursorDown)
	list.gui.SetKeybinding(listViewID, gocui.KeyArrowUp, gocui.ModNone, list.cursorUp)
}

func (list *UsersList) showAddUserView(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := g.Size()

	if v, err := g.SetView(addUserViewID, maxX/2-30, maxY/2, maxX/2+30, maxY/2+2); err != nil {
		v.Editable = false
		v.Frame = true
	}

	if v, err := g.SetView(addUserViewLabelID, maxX/2-30+1, maxY/2, maxX/2-20+1, maxY/2+2); err != nil {
		v.Frame = false
		fmt.Fprintln(v, "username:")
	}

	if v, err := g.SetView(addUserViewInputID, maxX/2-20+1, maxY/2, maxX/2+30, maxY/2+2); err != nil {
		v.Frame = false
		v.Editable = true
		if err := g.SetCurrentView(addUserViewInputID); err != nil {
			return err
		}
	}

	return nil
}

func (list *UsersList) validateAndAddUser(g *gocui.Gui, v *gocui.View) error {
	username := list.usernameToAdd()

	g.DeleteView(addUserViewID)
	g.DeleteView(addUserViewLabelID)
	g.DeleteView(addUserViewInputID)

	v = list.view()
	g.SetCurrentView(listViewID)

	manager := ssh.NewUsersManager()
	user := ssh.NewUser(username)

	err := manager.AddUser(user)
	if err == nil {
		fmt.Fprintln(v, username)
		return nil
	}

	return nil
}

func (list *UsersList) selectedUsername() string {
	v := list.view()

	_, cy := v.Cursor()
	selected, err := v.Line(cy)
	if err != nil {
		selected = ""
	}

	return selected
}

func (list *UsersList) focus() {
	v := list.view()
	list.gui.SetCurrentView(listViewID)

	list.selectCurrentUser(list.gui, v)

	list.setHelp()
}

func (list *UsersList) selectCurrentUser(g *gocui.Gui, v *gocui.View) error {
	username := list.selectedUsername()

	manager := ssh.NewUsersManager()
	user := manager.LoadUser(username)

	for _, listener := range list.listeners {
		listener.OnUserSelected(user)
	}

	return nil
}

func (list *UsersList) editCurrentUser(g *gocui.Gui, v *gocui.View) error {
	username := list.selectedUsername()

	manager := ssh.NewUsersManager()
	user := manager.LoadUser(username)

	for _, listener := range list.listeners {
		listener.OnStartEditingUser(user)
	}

	return nil
}

func (list *UsersList) cursorUp(g *gocui.Gui, v *gocui.View) error {
	cursorUp(g, v)
	list.selectCurrentUser(g, v)
	return nil
}

func (list *UsersList) cursorDown(g *gocui.Gui, v *gocui.View) error {
	cursorDown(g, v)
	list.selectCurrentUser(g, v)
	return nil
}