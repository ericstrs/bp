# bp (blueprint)

bp is a kanban manager with nested sub-board support to handle complex projects.


https://github.com/ericstrs/bp/assets/98285990/b92452f5-db64-4349-ae42-9fdbbb618641


The traditional TODO list is great for managing immediate tasks, but once the complexity increases it quickly becomes difficult to use. The kanban board is great for mapping out longer-term tasks, but is not the ideal daily task manager. This project is an attempt to integrate the two systems. Additionally, most existing tools aren't focused on breaking down complex problems into their granular parts.

## Features

* TUI support.
* Sub-board support.
* Tree-view support.
* Quickly change task priority.

### Limitations

* Currently no third party integrations.
* Currently no collaboration support.
* Tree view does not support horizontal scrolling. Thus, a heavily nested board may run off the screen.
* A given child board may only have a single parent. This prevents two or more board tasks from referencing the same board.

## Install

The command can be built from source or directly installed:

```
go install github.com/ericstrs/bp/cmd/bp@latest
```

## Documentation

Usage, controls, and other documentation has been embedded into the source code. See the source or run the application with the `help` command.

Global:

|Keys|Description|
|----|-----------|
|<kbd>q</kbd>|Quit the program|
|<kbd>z</kbd>|Toggle panel zoom|
|<kbd>TAB</kbd>|Switch between right and left panel|
|<kbd>k</kbd>, <kbd>j</kbd>|Move up and down|


TODO list:

|Keys|Description|
|----|-----------|
|<kbd>a</kbd>|Create new task|
|<kbd>e</kbd>|Edit the current task|
|<kbd>x</kbd>|Toggle the current task completion status|
|<kbd>y</kbd>|Yank the current task|
|<kbd>d</kbd>|Delete the current task|
|<kbd>p</kbd>|Paste the buffered task|
|<kbd>space</kbd>|Toggle the current task description|

Treeview:

|Keys|Description|
|----|-----------|
|<kbd>L</kbd>|If current node is a board node, enter it|
|<kbd>a</kbd>|If current node is the root node, append a new root board|
|<kbd>e</kbd>|If current node is a root board, edit it|
|<kbd>y</kbd>|If current node is a root board, yank it|
|<kbd>d</kbd>|If current node is a root board, delete it and all its children|
|<kbd>p</kbd>|If current node is the root node, paste the buffered root board|

Kanban board:

|Keys|Description|
|----|-----------|
|<kbd>l</kbd>|Move right|
|<kbd>L</kbd>|If the current task references a board, enter it|
|<kbd>h</kbd>|If in left most column then return to tree view, otherwise move left|
|<kbd>H</kbd>|Navigate back to parent board|
|<kbd>0</kbd>|Navigate to the left most column|
|<kbd>$</kbd>|Navigate to the right most column|
|<kbd>Enter</kbd>|Move the current task to the next board column, with warp around enabled|
|<kbd>k</kbd>|If the current task is first then switch focus to entire column, otherwise move up|
|<kbd>a</kbd>|If the column is selected, then add new column to the right, otherwise add new board task underneath the current task|
|<kbd>e</kbd>|If the entire column is selected, then edit it. Otherwise, add a new board task underneath the current task|
|<kbd>y</kbd>|If the entire column is selected, then yank it. Otherwise, yank the current task|
|<kbd>d</kbd>|If the entire column is selected, then delete it and all its sub tasks and sub boards. Otherwise, delete it and all its children.|
|<kbd>p</kbd>|If the entire column is selected, then paste buffered board column. Otherwise, paste the buffered board task.|
|<kbd>j</kbd>|If the entire column is selected, then move down to next item|
|<kbd>space</kbd>|Toggle task/board description|

Note: Delete operation buffers the deleted item (and all its children if it has any).
