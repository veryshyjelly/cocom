# Cocom (Competitive Companion Tui)

## Features:
- Integration with [CC](https://github.com/jmerle/competitive-companion)
- Creates solution file
- Open file in editor
- Library injection
- Testing and profiling
- Sandboxed running

## Example usage with Vscode
<img width="1280" height="876" alt="cocom_tutorial" src="https://github.com/user-attachments/assets/9137a89e-df72-4f01-831b-3c37df89ad84" />


## Usage
```bash
cocom [options]

Options:
  -c, --config <file>
      Path to the configuration file.
      (default: ./cocom.yml)

  -C, --root <directory>
      Project root directory.
      (default: current working directory)

  -h, --help
      Show this help message.

  -v, --version
      Show version information.
```

### Configuration
```yaml
author: YOUR_NAME_HERE
# command for opening file
editor: "code {{ .Filename }}" 

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
  # main.extension automatically detected as solution file
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

# control additional behaviors
create_file: on
run_on_save: on
```

### Shortcuts
```text
r       Run
n       New test case
f       Create file
e       Show errors
tab     Nagivate cases
c       Copy Solution
q       Quit

i       Input  | Answer
Space   Input  | Output
o       Answer | Output
d       Input  | Diff
```
