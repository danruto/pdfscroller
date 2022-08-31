# Description
Quick and dirty PDF image reader with auto scrolling

# Installation
## Locally
```sh
git clone git@github.com:danruto/pdfscroller.git
cd pdfscroller
go install .
```

## Remote
```sh
go install github.com/danruto/pdfscroller@latest
```

# Usage
```sh
pdfscroller <file.pdf>
```

# Keybinds
```
j: Decrement speed by 1.0
k: Increment speed by 1.0
h: Decrement speed by 40.0
l: Increment speed by 40.0
s: Pause scrolling
p: Jump to previous page
n: Jump to next page
u: Zoom in by 0.1
d: Zoom out by 0.1
```