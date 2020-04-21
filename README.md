tdraw
---
Inspired by [Explaining Code using ASCII Art](https://blog.regehr.org/archives/1653).<br/>
Supports box / line drawing, text input and eraser.

#### Install
```
go get -u github.com/aca/tdraw
```
#### Usage
```
tdraw > output
```

```
tdraw has 3 mode.

ESC: Box Mode(default)
L: Draw Line
I: Text
---
MouseR: Eraser
CTRL-C/CTRL-D: exit
```

```
  ┌──────────────┐
  │              │ BOX
  └──────────────┘
  <--------------- LINES
  --------------->

  TEXT

  ----  ---  ----> ERASE WITH MouseR
```

```
             +----------+
             v          |
 ┌─────────────────┐    |
 │   STATE A       │    |
 └─────────────────┘    |
             |          |
             v          |
 ┌─────────────────┐    |
 │                 │    |       ┌───────────────┐
 │   STATE B       │ ---------> │     FAIL      │
 │                 │    |       └───────────────┘
 └─────────────────┘    |
             |          |
             v          |
 ┌─────────────────┐    |
 │                 │    |
 │   STATE C       │    |
 │                 │ ---+
 └─────────────────┘

```
