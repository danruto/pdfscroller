# Description
Quick and dirty PDF reader with auto scrolling

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
j: Decrement speed by 0.2
k: Increment speed by 0.2
h: Decrement speed by 1.0
l: Increment speed by 1.0
s: Pause scrolling
p: Jump to previous page
n: Jump to next page
```