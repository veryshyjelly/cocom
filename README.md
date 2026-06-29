# Cocom (Competitive Companion Tui)

## Features:
- Integration with [CC](https://github.com/jmerle/competitive-companion)
- Creates solution file
- Open file in editor
- Library injection
- Testing and profiling
- Sandboxed running 

## Usage
```bash
cocom [options]

options:
  --config-file <file>: Path to config file
```

### Configuration
```yaml
author: Ayush Biswas # your name
# command for opening file
editor: "code {{ .Filename }} " 

filename: # rules for creating file
  rules:
      # base url
    - site: open.kattis.com
      # regex for capturing details from url
      regex: "problems/([^/?]+)/?"
      # template for file path
      template: "./src/bin/{{ index .Captures 1 }}.ml"

template: # config for templating
  source: ./src/main.ml
  modifier: |
    (* Created by {{ .Author }} at {{ .Time }} 
       {{ .Url }} *)
    {{ .Code }}

code: # post process code
  # useful for including library code
  modifier: |
    {{ .Code }}

lib: # register libraries
  regex: "open Lib"
  include:
    dsu: "./src/lib/dsu.ml"

# config for compiler
compiler:
  name: Ocaml
  compile: ocamlfind # empty if no compilation to be performed
  args:
    - ocamlopt
    - -O2
    - -o
    - a.out
    - main.ml
    - -linkpkg
    - -thread
    - -package
    - str,num,zarith,threads,containers,core,iter,batteries
  run: ./a.out

# control additional behaviours
create_file: on
run_on_save: on
```

### Shortcuts
```yaml
- r: run
- f: create file
- e: show error
- d: show diff 
- n: new test case
- a: answer + output
- i: input + answer (default)
- <space>: toggle output
```
